package miners

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

// Compile time assertion on interface
var _ MinerImpl = &VerthashMinerImpl{}

var cfgPath = "verthash-miner-tmpl.conf"

type VerthashMinerImpl struct {
	binaryRunner  *BinaryRunner
	clhashRates   map[int64]uint64
	cuhashRates   map[int64]uint64
	hashRatesLock sync.Mutex
}

func (l *VerthashMinerImpl) generateTempConf() error {
	os.Remove(filepath.Join(util.DataDirectory(), cfgPath))
	err := l.binaryRunner.launch([]string{"--gen-conf", filepath.Join(util.DataDirectory(), cfgPath)}, false)
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
	defer out.Close()

	var parsedDevices map[int]util.VerthashMinerDeviceConfig

	scanner := bufio.NewScanner(in)
	skip := false
	insideDeviceBlock := false
	deviceBlockStr := ""

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			skip = false
		}
		if strings.HasPrefix(line, "<Connection") {
			out.WriteString(fmt.Sprintf("<Connection Url = \"%s\"\n\tUsername = \"%s\"\n\tPassword = \"%s\"\n\tAlgorithm = \"Verthash\">\n\n", args.StratumUrl, args.StratumUsername, args.StratumPassword))
			skip = true
		}
		if strings.HasPrefix(line, "<Global") {
			out.WriteString(fmt.Sprintf("<Global Debug=\"false\" VerthashDataFileVerification=\"false\" VerthashDataFile=\"%s\">\n\n", filepath.Join(util.DataDirectory(), "verthash.dat")))
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
			out.WriteString(fmt.Sprintf("%s\n", line))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
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
	return []string{"--conf", filepath.Join(util.DataDirectory(), "verthash-miner.conf")}
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
