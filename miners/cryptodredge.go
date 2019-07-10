package miners

import (
	"strconv"
	"strings"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

// Compile time assertion on interface
var _ MinerImpl = &CryptoDredgeMinerImpl{}

type CryptoDredgeMinerImpl struct {
	binaryRunner *BinaryRunner
	hashRates    map[int64]uint64
}

func NewCryptoDredgeMinerImpl(br *BinaryRunner) MinerImpl {
	return &CryptoDredgeMinerImpl{binaryRunner: br, hashRates: map[int64]uint64{}}
}

func (l *CryptoDredgeMinerImpl) Configure(args BinaryArguments) error {
	return nil
}

func (l *CryptoDredgeMinerImpl) ParseOutput(line string) {
	if l.binaryRunner.Debug {
		logging.Debugf("[cryptodredge] %s\n", line)
	}
	line = strings.TrimSpace(line)

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

			l.hashRates[deviceIdx] = uint64(f)
		}
	}
}

func (l *CryptoDredgeMinerImpl) HashRate() uint64 {
	totalHash := uint64(0)
	for _, h := range l.hashRates {
		totalHash += h
	}
	return totalHash
}

func (l *CryptoDredgeMinerImpl) ConstructCommandlineArgs(args BinaryArguments) []string {
	return []string{"--intensity", "5", "--no-color", "-a", "lyra2v3", "-o", args.StratumUrl, "-u", args.StratumUsername, "-p", args.StratumPassword}
}
