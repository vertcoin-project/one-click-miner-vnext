package util

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const APP_NAME string = "vertcoin-ocm"

func DataDirectory() string {
	if runtime.GOOS == "windows" {
		return path.Join(os.Getenv("APPDATA"), APP_NAME)
	} else if runtime.GOOS == "darwin" {
		return path.Join(os.Getenv("HOME"), "Library", "Application Support", APP_NAME)
	} else if runtime.GOOS == "linux" {
		return path.Join(os.Getenv("HOME"), fmt.Sprintf(".%s", strings.ToLower(APP_NAME)))
	}
	return "."
}

func DownloadFile(url string, dest string) error {
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func UnzipFile(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return err
			}

		}
	}

	return nil
}

func GetGPU() string {
	if runtime.GOOS == "windows" {
		Info := exec.Command("cmd", "/C", "wmic path win32_VideoController get name")
		History, _ := Info.Output()
		return strings.TrimSpace(strings.Replace(string(History), "Name", "", -1))
	} else if runtime.GOOS == "linux" {
		Info := exec.Command("lspci")
		History, _ := Info.Output()
		lines := strings.Split(string(History), "\n")
		lastGpu := ""
		for _, l := range lines {
			if strings.Contains(l, "VGA compatible") {
				// this is a GPU
				lastGpu = l
			}
		}
		return lastGpu
	} else if runtime.GOOS == "darwin" {
		Info := exec.Command("system_profiler", "SPDisplaysDataType")
		History, _ := Info.Output()
		lines := strings.Split(string(History), "\n")
		lastGpu := ""
		for _, l := range lines {
			if strings.Contains(l, "Chipset Model:") {
				// this is a GPU
				lastGpu = l
			}
		}
		return lastGpu
	} else {
		return "Unknown OS, unable to detect GPU"
	}
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

var jsonClient = &http.Client{Timeout: 10 * time.Second}

func GetJson(url string, target interface{}) error {
	r, err := jsonClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
