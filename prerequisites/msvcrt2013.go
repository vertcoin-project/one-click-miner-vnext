package prerequisites

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func vcrt2013Installed() bool {
	searchPath := strings.Split(os.Getenv("PATH"), ";")
	for _, p := range searchPath {
		search := filepath.Join(p, "MSVCR120.dll")
		if _, err := os.Stat(search); !os.IsNotExist(err) {
			return true
		}
	}
	logging.Infof("Visual C++ Redistributable is not found\n")
	return false
}

func installVCRT2013(install chan bool) error {
	if vcrt2013Installed() {
		return nil
	}

	install <- true

	err := downloadVCRT2013()
	if err != nil {
		return err
	}

	err = checkVCRT2013Hash()
	if err != nil {
		return err
	}

	installer := exec.Command(vcrt2013DownloadPath(), "/q", "/norestart")
	err = installer.Run()
	install <- false
	return err
}

func vcrt2013DownloadPath() string {
	downloadDir := filepath.Join(util.DataDirectory(), "prerequisites")
	downloadPath := filepath.Join(downloadDir, "vcredist_x64.exe")
	os.MkdirAll(downloadDir, 0755)
	return downloadPath
}

func checkVCRT2013Hash() error {
	expectedHash, _ := hex.DecodeString("e554425243e3e8ca1cd5fe550db41e6fa58a007c74fad400274b128452f38fb8")
	realHash, err := util.ShaSum(vcrt2013DownloadPath())
	if err != nil {
		return err
	}
	if !bytes.Equal(realHash, expectedHash) {
		return fmt.Errorf("Hash of downloaded VCRT runtime installer does not match")
	}
	return nil
}

func downloadVCRT2013() error {

	resp, err := http.Get("https://download.microsoft.com/download/2/E/6/2E61CFA4-993B-4DD4-91DA-3737CD5CD6E3/vcredist_x64.exe")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(vcrt2013DownloadPath())
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
