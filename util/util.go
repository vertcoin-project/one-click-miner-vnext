package util

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/fastsha256"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
)

const APP_NAME string = "vertcoin-ocm"

func DataDirectory() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), APP_NAME)
	} else if runtime.GOOS == "darwin" {
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", APP_NAME)
	} else if runtime.GOOS == "linux" {
		return filepath.Join(os.Getenv("HOME"), fmt.Sprintf(".%s", strings.ToLower(APP_NAME)))
	}
	return "."
}

func ReplaceInFile(file string, find string, replace string) error {
	input, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	output := bytes.Replace(input, []byte(find), []byte(replace), -1)

	if err = ioutil.WriteFile(file, output, 0666); err != nil {
		return err
	}

	return nil
}

type BlocksResponse struct {
	Blocks     []Block           `json:"blocks"`
	Pagination InsightPagination `json:"pagination"`
}

type InsightPagination struct {
	Next string `json:"next"`
	Prev string `json:"prev"`
}

type Block struct {
	Hash   string `json:"hash"`
	Height int64  `json:"height"`
}

type BlockResponse struct {
	Bits uint32 `json:"bits"`
}

type getInfoResponse struct {
	Difficulty       float64 `json:"difficulty"`
	TipHeight        int64   `json:"tipHeight"`
	BackendTipHeight int64   `json:"backendTipHeight"`
}

type infoData struct {
	Difficulty float64 `json:"difficulty"`
}

type VerthashMinerDeviceConfig struct {
	DeviceIndex int
	PCIeBus     string
	OpenCL      bool
	Name        string
	Platform    string
}

var genesisDiff *big.Int

func init() {
	genesisDiff = blockchain.CompactToBig(0x1e00ffff)
}

func targetToDiff(target *big.Int) float64 {
	f, _ := big.NewFloat(0).Quo(big.NewFloat(0).SetInt(genesisDiff), big.NewFloat(0).SetInt(target)).Float64()
	return f
}

func GetDifficulty() float64 {
	info := getInfoResponse{}
	url := fmt.Sprintf("%sinfo", networks.Active.OCMBackend)
	GetJson(url, &info)
	return info.Difficulty
}

func GetNetHash() uint64 {
	difficulty := big.NewFloat(GetDifficulty())
	factor := big.NewInt(0).Exp(big.NewInt(2), big.NewInt(48), nil)
	netHash := difficulty.Mul(difficulty, big.NewFloat(0).SetInt(factor))
	u, _ := netHash.Quo(netHash, big.NewFloat(9830250)).Uint64() // 0xffff * blocktime in seconds
	logging.Debugf("Nethash: %d", u)

	return u
}

var jsonClient = &http.Client{Timeout: 60 * time.Second}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetJson(url string, target interface{}) error {
	r, err := jsonClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func PostJson(url string, payload interface{}, target interface{}) error {
	var b bytes.Buffer
	json.NewEncoder(&b).Encode(payload)
	r, err := jsonClient.Post(url, "application/json", bytes.NewBuffer(b.Bytes()))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	logging.Infof("POST JSON response: %s", string(bodyBytes))

	buf := bytes.NewBuffer(bodyBytes)
	return json.NewDecoder(buf).Decode(target)
}

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		logging.Error(err)
	}
}

func UnpackZip(archive, unpackPath string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {

		targetPath := filepath.Join(unpackPath, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(targetPath, filepath.Clean(unpackPath)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", targetPath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(targetPath, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(targetPath), 0755)
		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

func UnpackTar(archive, unpackPath string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			targetPath := filepath.Join(unpackPath, name)
			os.MkdirAll(filepath.Dir(targetPath), 0755)
			outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			defer outFile.Close()
			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ShaSum(file string) ([]byte, error) {
	h := fastsha256.New()
	fp, err := os.Open(file)
	if err != nil {
		return []byte{}, err
	}
	defer fp.Close()
	buf := make([]byte, 4096)

	for {
		n, err := fp.Read(buf)

		if err != nil && err != io.EOF {
			return []byte{}, err
		}

		if err == io.EOF {
			break
		} else {
			h.Write(buf[:n])
		}
	}
	return h.Sum(nil), nil
}

func ParseVerthashMinerDeviceCfg(cfg string) map[int]VerthashMinerDeviceConfig {
	lines := strings.Split(cfg, "\n")
	deviceCFG := VerthashMinerDeviceConfig{}
	allDeviceCFG := make(map[int]VerthashMinerDeviceConfig)
	platforms := make(map[int]string)

	// line 0 will always have the platform type
	if strings.Contains(lines[0], "OpenCL") {
		deviceCFG.OpenCL = true
	} else {
		deviceCFG.OpenCL = false
	}

	lines = lines[1:] //dont need to process the first line anymore...

	isGettingDeviceInfo := false
	isGettingPlatformInfo := false
	for _, line := range lines {

		if strings.Contains(line, "Available platforms") {
			isGettingPlatformInfo = true
		}
		if strings.Contains(line, "Available devices") {
			isGettingPlatformInfo = false
		}

		if isGettingPlatformInfo && strings.Contains(line, "Platform name") {
			parsedNum, _ := strconv.Atoi(string(line[2]))

			platforms[parsedNum-1] = strings.TrimSpace(strings.Split(line, ": ")[1])
		}

		if strings.Contains(line, "DeviceIndex:") {
			// if we were already getting info, save that device and start again
			if isGettingDeviceInfo {
				allDeviceCFG[deviceCFG.DeviceIndex] = deviceCFG
				deviceCFG = VerthashMinerDeviceConfig{}
			}
			isGettingDeviceInfo = true
			tmpStr := strings.SplitAfter(line, " ")

			// in order to capture negative / 2 digit numbers
			deviceCFG.DeviceIndex, _ = strconv.Atoi(strings.TrimSpace(tmpStr[len(tmpStr)-1]))
		}

		if isGettingDeviceInfo && strings.Contains(line, "Name:") {
			deviceCFG.Name = strings.TrimSpace(strings.SplitAfter(line, ":")[1])
		}

		if isGettingDeviceInfo && strings.Contains(line, "PCIeBusId:") {
			tmpStr := strings.SplitAfter(line, " ")

			deviceCFG.PCIeBus = strings.TrimSpace(tmpStr[len(tmpStr)-1])
		}

		if isGettingDeviceInfo && strings.Contains(line, "Platform index:") {
			tmpStr := strings.SplitAfter(line, ": ")

			platformIdx, _ := strconv.Atoi(strings.TrimSpace(tmpStr[len(tmpStr)-1]))

			deviceCFG.Platform = platforms[platformIdx]
		}

		if isGettingDeviceInfo && strings.Contains(line, "#-#-#-#-#-#-#-#-#-#-#-#-#-#-#-#") {
			isGettingDeviceInfo = false
		}
	}

	allDeviceCFG[deviceCFG.DeviceIndex] = deviceCFG
	deviceCFG = VerthashMinerDeviceConfig{}

	return allDeviceCFG
}
