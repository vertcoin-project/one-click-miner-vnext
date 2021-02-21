package miners

import (
	"strconv"
	"strings"
	"sync"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

// Compile time assertion on interface
var _ MinerImpl = &TeamRedMinerImpl{}

type TeamRedMinerImpl struct {
	binaryRunner  *BinaryRunner
	hashRates     map[int64]uint64
	hashRatesLock sync.Mutex
	gpuCount      int8
}

func NewTeamRedMinerImpl(br *BinaryRunner) MinerImpl {
	return &TeamRedMinerImpl{binaryRunner: br, hashRates: map[int64]uint64{}, hashRatesLock: sync.Mutex{}}
}

func (l *TeamRedMinerImpl) Configure(args BinaryArguments) error {
	return nil
}

func (l *TeamRedMinerImpl) ParseOutput(line string) {
	if l.binaryRunner.Debug {
		logging.Debugf("[teamRedMiner] %s\n", line)
	}
	line = strings.TrimSpace(line)
	if strings.Contains(line, "] Detected") && strings.Contains(line, "devices, ") {
		startCountIdx := strings.Index(line, "] Detected ") + 11
		gpuCountString := line[startCountIdx : startCountIdx+1]
		logging.Debugf("GPUCountString: %s", gpuCountString)
		gpuCount64, _ := strconv.ParseInt(gpuCountString, 10, 8)
		l.gpuCount = int8(gpuCount64)
		logging.Debugf("Set GPU Count to %d", l.gpuCount)
	}
	if strings.Contains(line, "] GPU ") && strings.Contains(line, "lyra2rev3") {
		startDeviceIdx := strings.Index(line, "] GPU ")
		endDeviceIdx := strings.Index(line[startDeviceIdx:], "[")
		deviceIdxString := line[startDeviceIdx+6 : startDeviceIdx+endDeviceIdx-1]
		deviceIdx, err := strconv.ParseInt(deviceIdxString, 10, 64)
		if err != nil {
			return
		}

		startMHs := strings.Index(line, "lyra2rev3: ")
		if startMHs > -1 {
			endMHs := strings.Index(line[startMHs:], "h/s")
			hashRateUnit := strings.ToUpper(line[startMHs+endMHs-1 : startMHs+endMHs])
			line = line[startMHs+11 : startMHs+endMHs-1]
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

func (l *TeamRedMinerImpl) HashRate() uint64 {
	totalHash := uint64(0)
	l.hashRatesLock.Lock()
	for _, h := range l.hashRates {
		totalHash += h
	}
	l.hashRatesLock.Unlock()
	return totalHash
}

func (l *TeamRedMinerImpl) ConstructCommandlineArgs(args BinaryArguments) []string {
	return []string{"--log_interval=10", "--disable_colors", "-a", "lyra2rev3", "-o", args.StratumUrl, "-u", args.StratumUsername, "-p", args.StratumPassword}
}

func (l *TeamRedMinerImpl) AvailableGPUs() int8 {
	l.binaryRunner.launch([]string{"--list_devices"}, false)
	l.binaryRunner.cmd.Wait()
	// Output is caught by ParseOuput function above and this will set the gpuCount accordingly
	return l.gpuCount
}
