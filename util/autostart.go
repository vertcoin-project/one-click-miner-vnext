package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ProtonMail/go-autostart"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
)

var app *autostart.App
var fullPath string
var oldFullPathFile string

func init() {
	fullPath, _ = filepath.Abs(os.Args[0])
	oldFullPath := ""
	oldFullPathFile = filepath.Join(DataDirectory(), "auto_start")
	if FileExists(oldFullPathFile) {
		oldFullPathBytes, err := ioutil.ReadFile(oldFullPath)
		if err != nil {
			oldFullPath = string(oldFullPathBytes)
		}
	}
	app = &autostart.App{
		Name:        "vertcoin-ocm",
		DisplayName: "Vertcoin One-Click miner",
		Exec:        []string{fullPath},
	}

	if oldFullPath != "" && oldFullPath != fullPath {
		oldApp := &autostart.App{
			Name:        "vertcoin-ocm",
			DisplayName: "Vertcoin One-Click miner",
			Exec:        []string{oldFullPath},
		}
		if oldApp.IsEnabled() {
			// We enabled autostart on a different location. Move it to the current
			// full executable path. Disable it on the old path and move to the new.
			logging.Debugf("Autostart was enabled on a different location.\nOld location: %s\nNew location: %s\nMoving it.", oldFullPath, fullPath)
			err := oldApp.Disable()
			if err == nil {
				err := app.Enable()
				if err == nil {
					ioutil.WriteFile(oldFullPathFile, []byte(fullPath), 0644)
				}
			}
		}
	}
}

func GetAutoStart() bool {
	return app.IsEnabled()
}

func SetAutoStart(autoStart bool) string {
	if autoStart {
		err := app.Enable()
		if err != nil {
			return err.Error()
		}
		// Store the full path we created the autostart for, so we can
		// re-enable it on a new path when someone decides to download
		// an update to a different location.
		ioutil.WriteFile(oldFullPathFile, []byte(fullPath), 0644)

	} else {
		err := app.Disable()
		if err != nil {
			return err.Error()
		}
		os.Remove(oldFullPathFile)
	}

	return ""
}
