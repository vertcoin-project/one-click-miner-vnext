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
var _ MinerImpl = &LyclMinerImpl{}

type LyclMinerImpl struct {
	binaryRunner  *BinaryRunner
	hashRates     map[int64]uint64
	hashRatesLock sync.Mutex
}

func NewLyclMinerImpl(br *BinaryRunner) MinerImpl {
	return &LyclMinerImpl{binaryRunner: br, hashRates: map[int64]uint64{}, hashRatesLock: sync.Mutex{}}
}

func (l *LyclMinerImpl) Configure(args BinaryArguments) error {
	os.Remove(filepath.Join(util.DataDirectory(), "lyclMiner_tmpl.conf"))
	err := l.binaryRunner.launch([]string{"-g", filepath.Join(util.DataDirectory(), "lyclMiner_tmpl.conf")}, false)
	err2 := l.binaryRunner.cmd.Wait()
	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	if !l.binaryRunner.cmd.ProcessState.Success() {
		return fmt.Errorf("Was unable to configure lyclMiner. Exit code %d", l.binaryRunner.cmd.ProcessState.ExitCode())
	}

	in, err := os.Open(filepath.Join(util.DataDirectory(), "lyclMiner_tmpl.conf"))
	if err != nil {
		logging.Error(err)
		return err
	}
	defer in.Close()

	os.Remove(filepath.Join(util.DataDirectory(), "lyclMiner.conf"))
	out, err := os.Create(filepath.Join(util.DataDirectory(), "lyclMiner.conf"))
	defer out.Close()

	scanner := bufio.NewScanner(in)
	skip := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			skip = false
		}
		if strings.HasPrefix(line, "<Connection") {
			out.WriteString(fmt.Sprintf("<Connection Url = \"%s\"\n\tUsername = \"%s\"\n\tPassword = \"%s\"\n\tAlgorithm = \"Lyra2REv3\">\n\n", args.StratumUrl, args.StratumUsername, args.StratumPassword))
			skip = true
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

func (l *LyclMinerImpl) ParseOutput(line string) {
	if l.binaryRunner.Debug {
		logging.Debugf("[lyclMiner] %s\n", line)
	}
	line = strings.TrimSpace(line)
	if strings.Contains(line, "Device #") && strings.HasSuffix(line, "H/s") {
		startDeviceIdx := strings.Index(line, "Device #")
		endDeviceIdx := strings.Index(line[startDeviceIdx:], ":")
		deviceIdxString := line[startDeviceIdx+8 : startDeviceIdx+endDeviceIdx]
		deviceIdx, err := strconv.ParseInt(deviceIdxString, 10, 64)
		if err != nil {
			return
		}

		startMHs := strings.LastIndex(line, ", ")
		if startMHs > -1 {
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
			l.hashRates[deviceIdx] = uint64(f)
			l.hashRatesLock.Unlock()
		}
	}
}

func (l *LyclMinerImpl) HashRate() uint64 {
	totalHash := uint64(0)
	l.hashRatesLock.Lock()
	for _, h := range l.hashRates {
		totalHash += h
	}
	l.hashRatesLock.Unlock()
	return totalHash
}

func (l *LyclMinerImpl) ConstructCommandlineArgs(args BinaryArguments) []string {
	return []string{filepath.Join(util.DataDirectory(), "lyclMiner.conf")}
}

func (l *LyclMinerImpl) AvailableGPUs() int8 {
	tmpCfg := filepath.Join(util.DataDirectory(), "lyclMiner_tmp.conf")
	err := l.binaryRunner.launch([]string{"-g", tmpCfg}, false)
	err2 := l.binaryRunner.cmd.Wait()
	if err != nil {
		return 0
	}
	if err2 != nil {
		return 0
	}

	if !l.binaryRunner.cmd.ProcessState.Success() {
		return 0
	}

	in, err := os.Open(filepath.Join(util.DataDirectory(), "lyclMiner_tmpl.conf"))
	if err != nil {
		logging.Error(err)
		return 0
	}
	gpu := int8(0)
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "<Device") {
			gpu++
		}
	}
	in.Close()
	os.Remove(tmpCfg)
	return gpu
}
