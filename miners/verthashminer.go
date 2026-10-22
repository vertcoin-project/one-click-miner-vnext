package miners

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

// Compile time assertion on interface
var _ MinerImpl = &VerthashMinerImpl{}

var cfgPath = "verthash-miner-tmpl.conf"

// cuda11MaxComputeCap is the highest compute capability the bundled CUDA 11
// build of VerthashMiner can run. Newer cards (e.g. RTX 50xx at 12.0) start
// without error but never hash, so they have to mine through OpenCL instead.
const cuda11MaxComputeCap = 9.0

var deviceIndexRe = regexp.MustCompile(`DeviceIndex\s*=\s*"(-?\d+)"`)

// normalizePCIeID turns nvidia-smi's "00000000:2B:00.0" and VerthashMiner's
// "2b:00:0" into the same "bus:device:function" form.
func normalizePCIeID(id string) string {
	parts := strings.Split(strings.ReplaceAll(strings.ToLower(strings.TrimSpace(id)), ".", ":"), ":")
	if len(parts) > 3 {
		parts = parts[len(parts)-3:]
	}
	return strings.Join(parts, ":")
}

// nvidiaComputeCaps returns the compute capability of each NVIDIA GPU keyed by
// PCIe id, or nil when nvidia-smi isn't available.
func nvidiaComputeCaps() map[string]float64 {
	cmd := exec.Command("nvidia-smi", "--query-gpu=pci.bus_id,compute_cap", "--format=csv,noheader")
	util.PrepareBackgroundCommand(cmd)
	out, err := cmd.Output()
	if err != nil {
		logging.Debugf("nvidia-smi unavailable, keeping CUDA for NVIDIA GPUs: %v\n", err)
		return nil
	}
	caps := map[string]float64{}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Split(line, ",")
		if len(fields) != 2 {
			continue
		}
		if c, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64); err == nil {
			caps[normalizePCIeID(fields[0])] = c
		}
	}
	return caps
}

// selectVerthashDevices takes a config generated with --no-restrict-cuda, where
// NVIDIA cards appear as both OpenCL and CUDA devices, and keeps exactly one
// entry per card: CUDA when the CUDA 11 build supports it, OpenCL otherwise.
// Cards without compute capability info keep CUDA, as before.
func selectVerthashDevices(lines []string, caps map[string]float64) []string {
	clPCIe, cuPCIe := map[int]string{}, map[int]string{}
	section, cur := "", -1
	for _, line := range lines {
		switch {
		case strings.Contains(line, "OpenCL device config"):
			section, cur = "cl", -1
		case strings.Contains(line, "CUDA Device config"):
			section, cur = "cu", -1
		case strings.HasPrefix(line, "# DeviceIndex:"):
			cur, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "# DeviceIndex:")))
		case cur >= 0 && strings.Contains(line, "PCIeId:"):
			id := normalizePCIeID(line[strings.Index(line, "PCIeId:")+len("PCIeId:"):])
			if section == "cl" {
				clPCIe[cur] = id
			} else if section == "cu" {
				cuPCIe[cur] = id
			}
		}
	}

	useCUDA := map[string]bool{}
	for _, id := range cuPCIe {
		if id != "" {
			cc, known := caps[id]
			useCUDA[id] = !known || cc <= cuda11MaxComputeCap
		}
	}

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		isCL, isCU := strings.HasPrefix(line, "<CL_Device"), strings.HasPrefix(line, "<CU_Device")
		if isCL || isCU {
			idx := -1
			if m := deviceIndexRe.FindStringSubmatch(line); m != nil {
				idx, _ = strconv.Atoi(m[1])
			}
			if isCL {
				if cuda, nvidia := useCUDA[clPCIe[idx]]; nvidia && cuda {
					continue // this NVIDIA card mines through CUDA
				}
			} else if !useCUDA[cuPCIe[idx]] {
				continue // the CUDA 11 build can't run this card, OpenCL entry is used
			}
		}
		out = append(out, line)
	}
	return out
}

type VerthashMinerImpl struct {
	binaryRunner  *BinaryRunner
	clhashRates   map[int64]uint64
	cuhashRates   map[int64]uint64
	hashRatesLock sync.Mutex
}

func (l *VerthashMinerImpl) generateTempConf() error {
	os.Remove(filepath.Join(util.DataDirectory(), cfgPath))
	// --no-restrict-cuda lists NVIDIA cards under OpenCL too, so cards the CUDA 11
	// build can't run still have a usable device entry (see selectVerthashDevices).
	err := l.binaryRunner.launch([]string{"--gen-conf", filepath.Join(util.DataDirectory(), cfgPath), "--no-restrict-cuda"}, false)
	var err2 error
	if l.binaryRunner.cmd != nil {
		err2 = l.binaryRunner.cmd.Wait()
	}
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func NewVerthashMinerImpl(br *BinaryRunner) MinerImpl {
	return &VerthashMinerImpl{binaryRunner: br, clhashRates: map[int64]uint64{}, cuhashRates: map[int64]uint64{}, hashRatesLock: sync.Mutex{}}
}

func (l *VerthashMinerImpl) Configure(args BinaryArguments) error {
	err := l.generateTempConf()
	if err != nil {
		return err
	}

	if !l.binaryRunner.cmd.ProcessState.Success() {
		return fmt.Errorf("Was unable to configure VerthashMiner. Exit code %d", l.binaryRunner.cmd.ProcessState.ExitCode())
	}

	in, err := os.Open(filepath.Join(util.DataDirectory(), "verthash-miner-tmpl.conf"))
	if err != nil {
		logging.Error(err)
		return err
	}
	defer in.Close()

	os.Remove(filepath.Join(util.DataDirectory(), "verthash-miner.conf"))
	out, err := os.Create(filepath.Join(util.DataDirectory(), "verthash-miner.conf"))
	if err != nil {
		logging.Error(err)
		return err
	}
	defer func() {
		err := out.Close()
		if err != nil {
			logging.Error(err)
		}
	}()

	var parsedDevices map[int]util.VerthashMinerDeviceConfig

	var tmplLines []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		tmplLines = append(tmplLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	caps := nvidiaComputeCaps()
	for id, cc := range caps {
		if cc > cuda11MaxComputeCap {
			logging.Infof("GPU %s has compute capability %.1f, mining it through OpenCL", id, cc)
		}
	}
	tmplLines = selectVerthashDevices(tmplLines, caps)

	skip := false
	insideDeviceBlock := false
	deviceBlockStr := ""

	for _, line := range tmplLines {
		if strings.HasPrefix(line, "#") {
			skip = false
		}
		if strings.HasPrefix(line, "<Connection") {
			_, err = out.WriteString(fmt.Sprintf("<Connection Url = \"%s\"\n\tUsername = \"%s\"\n\tPassword = \"%s\"\n\tAlgorithm = \"Verthash\">\n\n", args.StratumUrl, args.StratumUsername, args.StratumPassword))
			if err != nil {
				return err
			}
			skip = true
		}
		if strings.HasPrefix(line, "<Global") {
			_, err = out.WriteString(fmt.Sprintf("<Global Debug=\"false\" VerthashDataFileVerification=\"false\" VerthashDataFile=\"%s\">\n\n", filepath.Join(util.DataDirectory(), "verthash.dat")))
			if err != nil {
				return err
			}
			skip = true
		}

		if strings.Contains(line, "OpenCL device config") || strings.Contains(line, "CUDA Device config") {
			logging.Debug("Entering device block")
			insideDeviceBlock = true
		} else if insideDeviceBlock {
			deviceBlockStr += line + "\n"
		}

		if strings.Contains(line, "#-#-#-#-#-#-#-#-#-#-#-") && insideDeviceBlock {
			insideDeviceBlock = false
			parsedDevices = util.ParseVerthashMinerDeviceCfg(deviceBlockStr)
			logging.Debug("Exiting device block")
			logging.Debug(parsedDevices[0])
			deviceBlockStr = ""
		}

		if strings.HasPrefix(line, "<CL_Device") {
			words := strings.SplitAfter(line, " ")
			thisDeviceIndexNumber, _ := strconv.Atoi(strings.Trim(words[3], "\""))

			if device, ok := parsedDevices[thisDeviceIndexNumber]; ok {
				if strings.Contains(device.Platform, "Intel") && !args.EnableIntegrated {
					logging.Debug("Intel disabled.")
					skip = true
				}
			}
		}

		if !skip {
			_, err = out.WriteString(fmt.Sprintf("%s\n", line))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *VerthashMinerImpl) ParseOutput(line string) {
	if l.binaryRunner.Debug {
		logging.Debugf("[VerthashMiner] %s\n", line)
	}
	line = strings.TrimSpace(line)
	if strings.Contains(line, "_device(") && strings.HasSuffix(line, "H/s") {
		startMHs := strings.LastIndex(line, ": ")
		if startMHs > -1 {
			deviceIdxStart := strings.Index(line, "_device(") + 8
			deviceTypeStart := strings.Index(line, "_device(") - 2
			deviceIdxEnd := strings.Index(line[deviceIdxStart:], ")")
			deviceIdxString := line[deviceIdxStart : deviceIdxStart+deviceIdxEnd]
			deviceIdx, _ := strconv.ParseInt(deviceIdxString, 10, 64)
			deviceType := line[deviceTypeStart : deviceTypeStart+2]

			hashRateUnit := strings.ToUpper(line[len(line)-4 : len(line)-3])
			line = line[startMHs+2 : len(line)-5]
			f, err := strconv.ParseFloat(line, 64)
			if err != nil {
				logging.Errorf("Error parsing hashrate: %s\n", err.Error())
			}
			if hashRateUnit == "K" {
				f = f * 1000
			} else if hashRateUnit == "M" {
				f = f * 1000 * 1000
			} else if hashRateUnit == "G" {
				f = f * 1000 * 1000 * 1000
			}

			l.hashRatesLock.Lock()
			if deviceType == "cu" {
				l.cuhashRates[deviceIdx] = uint64(f)
			} else {
				l.clhashRates[deviceIdx] = uint64(f)
			}
			l.hashRatesLock.Unlock()
		}
	}
}

func (l *VerthashMinerImpl) HashRate() uint64 {
	totalHash := uint64(0)
	l.hashRatesLock.Lock()
	for _, h := range l.cuhashRates {
		totalHash += h
	}
	for _, h := range l.clhashRates {
		totalHash += h
	}
	l.hashRatesLock.Unlock()

	return totalHash
}

func (l *VerthashMinerImpl) ConstructCommandlineArgs(args BinaryArguments) []string {
	// Must match the --no-restrict-cuda used to generate the config, or the
	// OpenCL device indices in it won't line up.
	return []string{"--conf", filepath.Join(util.DataDirectory(), "verthash-miner.conf"), "--no-restrict-cuda"}
}

func (l *VerthashMinerImpl) AvailableGPUs() int8 {
	logging.Debugf("AvailableGPUs called\n")
	tmpCfg := filepath.Join(util.DataDirectory(), "verthash-miner-tmp.conf")
	err := l.binaryRunner.launch([]string{"--gen-conf", tmpCfg}, false)
	err2 := l.binaryRunner.cmd.Wait()
	if err != nil {
		logging.Error(err)
		return 0
	}
	if err2 != nil {
		logging.Error(err)
		return 0
	}

	if !l.binaryRunner.cmd.ProcessState.Success() {
		logging.Errorf("Process state: %d", l.binaryRunner.cmd.ProcessState)
		return 0
	}

	in, err := os.Open(tmpCfg)
	if err != nil {
		logging.Error(err)
		return 0
	}
	gpu := int8(0)
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "<CL_Device") {
			gpu++
		}
		if strings.HasPrefix(line, "<CU_Device") {
			gpu++
		}
	}
	in.Close()
	os.Remove(tmpCfg)
	return gpu
}
