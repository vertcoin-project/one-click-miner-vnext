package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/leaanthony/mewn"
	"github.com/marcsauter/single"
	"github.com/vertiond/verthash-one-click-miner/backend"
	"github.com/vertiond/verthash-one-click-miner/logging"
	"github.com/vertiond/verthash-one-click-miner/networks"
	"github.com/vertiond/verthash-one-click-miner/tracking"
	"github.com/vertiond/verthash-one-click-miner/util"
	"github.com/wailsapp/wails"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			// Reopen log file, since it's closed now!
			logging.SetLogLevel(int(logging.LogLevelDebug))
			logFilePath := filepath.Join(util.DataDirectory(), "debug.log")
			logFile, _ := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			logging.SetLogFile(logFile)
			defer logFile.Close()
			logging.Errorf("%v\n%s\n", err, string(debug.Stack()))

			tracking.Track(tracking.TrackingRequest{
				Category: "Lifecycle",
				Action:   "Crash",
				Name:     fmt.Sprintf("%v", err),
			})

		}
	}()

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

	log.Printf("OCM v%s Started up\n", tracking.GetVersion())

	app := wails.CreateApp(&wails.AppConfig{
		Width:  800,
		Height: 400,
		Title:  "Verthash One Click Miner",
		JS:     js,
		CSS:    css,
		Colour: "#131313",
	})

	alreadyRunning := false
	s := single.New("verthash-ocm")
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
	networks.SetNetwork(backend.GetTestnet())

	backend.ResetWalletAddress()
	backend.ResetPool()
	backend.ResetCustomAddress()
	backend.ResetPayout()
	app.Bind(backend)
	app.Run()
	backend.StopMining()

	tracking.Track(tracking.TrackingRequest{
		Category: "Lifecycle",
		Action:   "Shutdown",
	})

	tracking.Stop()
}
