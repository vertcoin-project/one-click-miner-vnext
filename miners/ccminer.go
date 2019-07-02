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
	hashRate     uint64
}

func NewCCMinerImpl(br *BinaryRunner) MinerImpl {
	return &CCMinerImpl{binaryRunner: br}
}

func (l *CCMinerImpl) Configure(args BinaryArguments) error {
	return nil
}

func (l *CCMinerImpl) ParseOutput(line string) {
	line = strings.TrimSpace(line)
	logging.Debugf("[ccminer] %s", line)
	if strings.Contains(line, "MH/s") {
		endMHs := strings.LastIndex(line, "MH/s")
		startMHs := strings.LastIndex(line, ", ")
		if startMHs > -1 {
			line = line[startMHs+2 : endMHs-1]
			f, err := strconv.ParseFloat(line, 64)
			if err != nil {
				logging.Errorf("Error parsing hashrate: %s\n", err.Error())
			}
			f = f * 1000 * 1000
			l.hashRate = uint64(f)
		}
	}
}

func (l *CCMinerImpl) HashRate() uint64 {
	return l.hashRate
}

func (l *CCMinerImpl) ConstructCommandlineArgs(args BinaryArguments) []string {
	return []string{"-a", "lyra2v3", "-o", args.StratumUrl, "-u", args.StratumUsername, "-p", args.StratumPassword}
}
