package miners

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

type MinerBinary struct {
	Platform           string `json:"platform"`
	GpuPlatformString  string `json:"gpuplatform"`
	Url                string `json:"url"`
	Hash               string `json:"sha256"`
	MainExecutableName string `json:"mainExecutableName"`
	ClosedSource       bool   `json:"closedSource"`
	GPUType            util.GPUType
}

func GetMinerBinaries() []MinerBinary {
	binaries := []MinerBinary{}
	util.GetJson("https://raw.githubusercontent.com/vertcoin-project/one-click-miner-vnext/master/miners.json", &binaries)
	for i := range binaries {
		if binaries[i].GpuPlatformString == "AMD" {
			binaries[i].GPUType = util.GPUTypeAMD
		} else if binaries[i].GpuPlatformString == "NVIDIA" {
			binaries[i].GPUType = util.GPUTypeNVidia
		} else {
			logging.Warnf("Found unrecognized platform [%s] in miners.json\n", binaries[i].GpuPlatformString)
		}
	}
	return binaries
}

type MinerImpl interface {
	ParseOutput(line string)
	Configure(args BinaryArguments) error
	HashRate() uint64
	ConstructCommandlineArgs(args BinaryArguments) []string
}

func NewBinaryRunner(m MinerBinary) (*BinaryRunner, error) {
	br := &BinaryRunner{MinerBinary: m}
	if strings.HasPrefix(m.MainExecutableName, "lycl") {
		br.MinerImpl = NewLyclMinerImpl(br)
	} else if strings.HasPrefix(m.MainExecutableName, "ccminer") {
		br.MinerImpl = NewCCMinerImpl(br)
	} else if strings.HasPrefix(m.MainExecutableName, "teamred") {
		br.MinerImpl = NewTeamRedMinerImpl(br)
	} else if strings.HasPrefix(m.MainExecutableName, "Cryptodredge") {
		br.MinerImpl = NewCryptoDredgeMinerImpl(br)
	} else {
		return nil, fmt.Errorf("Could not determine implementation for miner binary")
	}
	return br, nil
}

type BinaryArguments struct {
	StratumUrl      string
	StratumUsername string
	StratumPassword string
}

type BinaryRunner struct {
	MinerBinary MinerBinary
	MinerImpl   MinerImpl
	cmd         *exec.Cmd
	Debug       bool
}

func (b *BinaryRunner) logPrefix() string {
	return fmt.Sprintf("[Miner %s/%d]", b.MinerBinary.Platform, b.MinerBinary.GPUType)
}

func (b *BinaryRunner) Stop() error {

	if b.cmd == nil {
		// not started (yet)
		return nil
	}
	// Windows doesn't support Interrupt
	if runtime.GOOS == "windows" {
		_ = b.cmd.Process.Signal(os.Kill)
		return nil
	}

	go func() {
		time.Sleep(15 * time.Second)
		_ = b.cmd.Process.Signal(os.Kill)
	}()
	b.cmd.Process.Signal(os.Interrupt)

	return b.wait()
}

func (b *BinaryRunner) IsRunning() bool {
	if b.cmd == nil {
		return false
	} else {
		if b.cmd.Process == nil {
			return false
		} else {
			if b.cmd.ProcessState != nil {
				return false
			}
		}
	}
	return true
}

func (b *BinaryRunner) Install() error {
	// Check if the archive is available and it has the right SHA sum. Download if not
	err := b.ensureAvailable()
	if err != nil {
		return err
	}

	// Always re-unpack the archive to ensure no one tampered with the file on disk.
	err = b.unpack()
	if err != nil {
		return err
	}

	return nil
}

func (b *BinaryRunner) HashRate() uint64 {
	return b.MinerImpl.HashRate()
}

func (b *BinaryRunner) Start(args BinaryArguments) error {
	err := b.Install()
	if err != nil {
		return err
	}

	// Always do a fresh unpack of the executable to ensure there's been no funny
	// business. EnsureAvailable already checked the SHA hash.
	err = b.launch(b.MinerImpl.ConstructCommandlineArgs(args))
	if err != nil {
		return err
	}

	return nil
}

func (b *BinaryRunner) unpackDir() string {
	return filepath.Join(util.DataDirectory(), "miners", fmt.Sprintf("unpacked-%s", b.MinerBinary.Hash))
}

func (b *BinaryRunner) downloadPath() string {
	return filepath.Join(util.DataDirectory(), "miners", b.MinerBinary.Hash)
}

func (b *BinaryRunner) launch(params []string) error {
	exePath := b.findExecutable()
	if exePath == "" {
		return fmt.Errorf("Cannot find main miner binary in unpack folder")
	}
	logging.Debugf("Launching %s %v\n", exePath, params)
	b.cmd = exec.Command(exePath, params...)
	util.PrepareBackgroundCommand(b.cmd)
	b.cmd.Dir = filepath.Dir(exePath)
	r, w := io.Pipe()
	go func(b *BinaryRunner, rd io.Reader) {
		br := bufio.NewReader(rd)

		for {
			l, _, e := br.ReadLine()
			if e != nil {
				logging.Debugf("%sError on readline from stdout/err: %s", b.logPrefix(), e.Error())
				return
			}
			b.MinerImpl.ParseOutput(string(l))
		}
	}(b, r)
	b.cmd.Stderr = w
	b.cmd.Stdout = w
	return b.cmd.Start()
}

func (b *BinaryRunner) wait() error {
	return b.cmd.Wait()
}

func (b *BinaryRunner) unpack() error {
	unpackDir := b.unpackDir()

	if _, err := os.Stat(unpackDir); !os.IsNotExist(err) {
		logging.Debugf("%sRemoving unpack directory", b.logPrefix())
		err = os.RemoveAll(unpackDir)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(unpackDir); os.IsNotExist(err) {
		logging.Debugf("%s(Re)creating unpack directory", b.logPrefix())
		err = os.MkdirAll(unpackDir, 0755)
		if err != nil {
			return err
		}
	}

	archive := b.downloadPath()
	if strings.HasSuffix(b.MinerBinary.Url, ".zip") {
		return util.UnpackZip(archive, unpackDir)
	} else if strings.HasSuffix(b.MinerBinary.Url, ".tar.gz") || strings.HasSuffix(b.MinerBinary.Url, ".tgz") {
		return util.UnpackTar(archive, unpackDir)
	}

	return fmt.Errorf("Unknown archive format, cannot unpack: %s", b.MinerBinary.Url)
}

func (b *BinaryRunner) findExecutable() string {
	mainExecutablePath := ""
	filepath.Walk(b.unpackDir(),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == b.MinerBinary.MainExecutableName {
				mainExecutablePath = path
			}
			return nil
		})
	return mainExecutablePath
}

func (b *BinaryRunner) ensureAvailable() error {
	freshDownload := false
	_ = os.Mkdir(filepath.Join(util.DataDirectory(), "miners"), 0700)
	nodePath := b.downloadPath()
	_, err := os.Stat(nodePath)
	if os.IsNotExist(err) {
		logging.Debugf("%sBinary not found, downloading...", b.logPrefix())
		freshDownload = true
		b.download()
	} else if err != nil {
		return err
	} else {
		logging.Debugf("%sDaemon file already exists", b.logPrefix())
	}

	shaSum, err := util.ShaSum(nodePath)
	if err != nil {
		return err
	}
	expectedHash, _ := hex.DecodeString(b.MinerBinary.Hash)
	if !bytes.Equal(shaSum, expectedHash) {
		logging.Warnf("%sHash differs: [%x] vs [%s]", b.logPrefix(), shaSum, b.MinerBinary.Hash)
		if !freshDownload {
			err = os.Remove(nodePath)
			if err != nil {
				return err
			}
			return b.ensureAvailable()
		} else {
			err = fmt.Errorf("%sFreshly downloaded node did not have correct SHA256 hash", b.logPrefix())
			logging.Error(err)
			return err
		}
	}

	logging.Debugf("%sDaemon file is available and correct", b.logPrefix())
	return nil
}

func (b *BinaryRunner) download() error {
	nodePath := b.downloadPath()

	resp, err := http.Get(b.MinerBinary.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(nodePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	logging.Debugf("%sDaemon file downloaded", b.logPrefix())
	return err
}
