package main

import (
	"os"
	"path"

	"github.com/leaanthony/mewn"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/mining"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/wailsapp/wails"
)

func main() {
	js := mewn.String("./frontend/dist/app.js")
	css := mewn.String("./frontend/dist/app.css")
	logging.SetLogLevel(int(logging.LogLevelDebug))
	if _, err := os.Stat(util.DataDirectory()); os.IsNotExist(err) {
		logging.Infof("Creating data directory")
		os.MkdirAll(util.DataDirectory(), 0700)
	}

	logFilePath := path.Join(util.DataDirectory(), "debug.log")
	logFile, _ := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	logging.SetLogFile(logFile)

	app := wails.CreateApp(&wails.AppConfig{
		Width:  800,
		Height: 400,
		Title:  "Vertcoin One Click Miner",
		JS:     js,
		CSS:    css,
		Colour: "#131313",
	})

	app.Bind(mining.NewMinerCore())
	app.Run()
}
