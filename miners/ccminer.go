package miners

import (
	"strconv"
	"strings"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

// Compile time assertion on interface
var _ MinerImpl = &CCMinerImpl{}

type CCMinerImpl struct {
	binaryRunner *BinaryRunner
	hashRates    map[int64]uint64
}

func NewCCMinerImpl(br *BinaryRunner) MinerImpl {
	return &CCMinerImpl{binaryRunner: br, hashRates: map[int64]uint64{}}
}

func (l *CCMinerImpl) Configure(args BinaryArguments) error {
	return nil
}

func (l *CCMinerImpl) ParseOutput(line string) {
	if l.binaryRunner.Debug {
		logging.Debugf("[ccminer] %s\n", line)
	}
	line = strings.TrimSpace(line)

	if strings.Contains(line, "GPU #") && strings.HasSuffix(line, "H/s") {
		startDeviceIdx := strings.Index(line, "GPU #")
		endDeviceIdx := strings.Index(line[startDeviceIdx:], ":")
		deviceIdxString := line[startDeviceIdx+5 : startDeviceIdx+endDeviceIdx]
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

			l.hashRates[deviceIdx] = uint64(f)
		}
	}
}

func (l *CCMinerImpl) HashRate() uint64 {
	totalHash := uint64(0)
	for _, h := range l.hashRates {
		totalHash += h
	}
	return totalHash
}

func (l *CCMinerImpl) ConstructCommandlineArgs(args BinaryArguments) []string {
	return []string{"--max-log-rate", "0", "--no-color", "-a", "lyra2v3", "-o", args.StratumUrl, "-u", args.StratumUsername, "-p", args.StratumPassword}
}
