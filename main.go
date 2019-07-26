package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/leaanthony/mewn"
	"github.com/marcsauter/single"
	"github.com/vertcoin-project/one-click-miner-vnext/backend"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/wailsapp/wails"
)

func main() {
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

	alreadyRunning := false
	s := single.New("vertcoin-ocm")
	if err := s.CheckLock(); err != nil && err == single.ErrAlreadyRunning {
		alreadyRunning = true
	} else if err == nil {
		defer s.TryUnlock()
	}

	backend, err := backend.NewBackend(alreadyRunning)
	if err != nil {
		logging.Errorf("Error creating Backend: %s", err.Error())
		panic(err)
	}
	app.Bind(backend)
	app.Run()
	backend.StopMining()

	tracking.Track(tracking.TrackingRequest{
		Category: "Lifecycle",
		Action:   "Shutdown",
	})

	tracking.Stop()
}
