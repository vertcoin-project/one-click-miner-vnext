package miners

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

// Compile time assertion on interface
var _ MinerImpl = &CryptoDredgeMinerImpl{}

type CryptoDredgeMinerImpl struct {
	binaryRunner  *BinaryRunner
	hashRates     map[int64]uint64
	hashRatesLock sync.Mutex
	gpuCount      int8
}

func NewCryptoDredgeMinerImpl(br *BinaryRunner) MinerImpl {
	return &CryptoDredgeMinerImpl{binaryRunner: br, hashRates: map[int64]uint64{}, hashRatesLock: sync.Mutex{}}
}

func (l *CryptoDredgeMinerImpl) Configure(args BinaryArguments) error {
	return nil
}

func (l *CryptoDredgeMinerImpl) ParseOutput(line string) {
	if l.binaryRunner.Debug {
		logging.Debugf("[cryptodredge] %s\n", line)
	}
	line = strings.TrimSpace(line)

	if strings.Contains(line, "INFO  - GPU") && strings.Contains(line, "MB") {
		startCountIdx := strings.Index(line, "INFO  - GPU") + 11
		gpuCountString := line[startCountIdx : startCountIdx+1]
		gpuCount64, _ := strconv.ParseInt(gpuCountString, 10, 8)
		l.gpuCount = int8(gpuCount64) + 1
		logging.Debugf("Set GPU Count to %d", l.gpuCount)
	}
	if strings.Contains(line, "INFO  - GPU") && strings.Contains(line, "H/s") {
		startDeviceIdx := strings.Index(line, "INFO  - GPU")
		endDeviceIdx := strings.Index(line[startDeviceIdx+9:], " ")
		deviceIdxString := line[startDeviceIdx+11 : startDeviceIdx+9+endDeviceIdx]
		deviceIdx, err := strconv.ParseInt(deviceIdxString, 10, 64)
		if err != nil {
			return
		}

		endMHs := strings.Index(line, "H/s")
		if endMHs > -1 {
			hashRateUnit := strings.ToUpper(line[endMHs-1 : endMHs])
			line = line[:endMHs-1]
			line = line[strings.LastIndex(line, " ")+1:]
			line = strings.ReplaceAll(line, ",", ".")
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

func (l *CryptoDredgeMinerImpl) HashRate() uint64 {
	totalHash := uint64(0)
	l.hashRatesLock.Lock()
	for _, h := range l.hashRates {
		totalHash += h
	}
	l.hashRatesLock.Unlock()
	return totalHash
}

func (l *CryptoDredgeMinerImpl) ConstructCommandlineArgs(args BinaryArguments) []string {
	return []string{"--intensity", "5", "--no-color", "-a", "lyra2v3", "-o", args.StratumUrl, "-u", args.StratumUsername, "-p", args.StratumPassword}
}

func (l *CryptoDredgeMinerImpl) AvailableGPUs() int8 {
	l.binaryRunner.launch([]string{}, false)
	time.Sleep(time.Second)
	l.binaryRunner.Stop()
	// Output is caught by ParseOuput function above and this will set the gpuCount accordingly
	return l.gpuCount
}
