package util

import (
	"os"
	"path/filepath"

	"github.com/ProtonMail/go-autostart"
)

var app *autostart.App

func init() {
	fullPath, _ := filepath.Abs(os.Args[0])

	app = &autostart.App{
		Name:        "vertcoin-ocm",
		DisplayName: "Vertcoin One-Click miner",
		Exec:        []string{fullPath},
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
	} else {
		err := app.Disable()
		if err != nil {
			return err.Error()
		}
	}

	return ""
}
