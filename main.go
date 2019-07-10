package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leaanthony/mewn"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/mining"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/wailsapp/wails"
)

func main() {
	debugMiner := flag.Bool("debugminers", false, "Print the output of the miner binaries to the debug log")
	flag.Parse()
	js := mewn.String("./frontend/dist/app.js")
	css := mewn.String("./frontend/dist/app.css")

	tracking.StartTracker()

	tracking.Track(tracking.TrackingRequest{
		Category: "Lifecycle",
		Action:   "Startup",
		Name:     fmt.Sprintf("OCM/%s", tracking.GetVersion()),
	})

	logging.SetLogLevel(int(logging.LogLevelDebug))
	if _, err := os.Stat(util.DataDirectory()); os.IsNotExist(err) {
		logging.Infof("Creating data directory")
		os.MkdirAll(util.DataDirectory(), 0700)
	}

	logFilePath := filepath.Join(util.DataDirectory(), "debug.log")
	logFile, _ := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	logging.SetLogFile(logFile)
	defer logFile.Close()
	app := wails.CreateApp(&wails.AppConfig{
		Width:  800,
		Height: 400,
		Title:  "Vertcoin One Click Miner",
		JS:     js,
		CSS:    css,
		Colour: "#131313",
	})

	core := mining.NewMinerCore()
	core.DebugMiners = *debugMiner
	app.Bind(core)
	app.Run()
	core.StopMining()

	tracking.Track(tracking.TrackingRequest{
		Category: "Lifecycle",
		Action:   "Shutdown",
	})

	tracking.Stop()
}
