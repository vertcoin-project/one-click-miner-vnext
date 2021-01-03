package miners

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

// Compile time assertion on interface
var _ MinerImpl = &VerthashMinerImpl{}

type VerthashMinerImpl struct {
	binaryRunner *BinaryRunner
	hashRates    map[int64]uint64
	hashRatesLock sync.Mutex
}

func NewVerthashMinerImpl(br *BinaryRunner) MinerImpl {
	return &VerthashMinerImpl{binaryRunner: br, hashRates: map[int64]uint64{}, hashRatesLock: sync.Mutex{}}
}

func (l *VerthashMinerImpl) Configure(args BinaryArguments) error {
	os.Remove(filepath.Join(util.DataDirectory(), "verthash-miner-tmpl.conf"))
	err := l.binaryRunner.launch([]string{"--gen-conf", filepath.Join(util.DataDirectory(), "verthash-miner-tmpl.conf")}, false)
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
		if strings.HasPrefix(line, "<Global") {
			out.WriteString(fmt.Sprintf("<Global Debug=\"false\" VerthashDataFileVerification=\"true\" VerthashDataFile=\"%s\">\n\n", filepath.Join(util.DataDirectory(), "verthash.dat")))
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

func (l *VerthashMinerImpl) ParseOutput(line string) {
	//if l.binaryRunner.Debug {
	logging.Debugf("[VerthashMiner] %s\n", line)
	//}
	line = strings.TrimSpace(line)
	if strings.Contains(line, "total hashrate:") && strings.HasSuffix(line, "H/s") {
		startMHs := strings.LastIndex(line, ": ")
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
			l.hashRates[0] = uint64(f)
			l.hashRatesLock.Unlock()
			
		}
	}
}

func (l *VerthashMinerImpl) HashRate() uint64 {
	totalHash := uint64(0)
	l.hashRatesLock.Lock()
	for _, h := range l.hashRates {
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
