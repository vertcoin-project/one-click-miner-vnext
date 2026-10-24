package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	_ "embed"

	"github.com/marcsauter/single"
	"github.com/vertcoin-project/one-click-miner-vnext/backend"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/wailsapp/wails"
)

//nolint:all
//go:embed frontend/dist/app.js
var js string

//nolint:all
//go:embed frontend/dist/app.css
var css string

// background matches the html background in frontend/src/assets/css/main.css
const background = "#191b1c"

// html is the page Wails loads before it injects app.css. Painting the
// background here keeps the window from showing white until then.
const html = `<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <style>html, body { background-color: ` + background + `; margin: 0; }</style>
</head>

<body>
  <div id="app"></div>
</body>

</html>`

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

	tracking.StartTracker()

	tracking.Track(tracking.TrackingRequest{
		Category: "Lifecycle",
		Action:   "Startup",
		Name:     fmt.Sprintf("OCM/%s", tracking.GetVersion()),
	})

	logging.SetLogLevel(int(logging.LogLevelDebug))
	if _, err := os.Stat(util.DataDirectory()); os.IsNotExist(err) {
		logging.Infof("Creating data directory")
		err = os.MkdirAll(util.DataDirectory(), 0700)
		if err != nil && !os.IsExist(err) {
			logging.Errorf("Error creating data directory, cannot continue")
			os.Exit(1)
		}
	}

	logFilePath := filepath.Join(util.DataDirectory(), "debug.log")
	logFile, _ := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	logging.SetLogFile(logFile)
	defer logFile.Close()

	log.Printf("OCM v%s Started up\n", tracking.GetVersion())

	app := wails.CreateApp(&wails.AppConfig{
		Width:  800,
		Height: 400,
		Title:  "Vertcoin One Click Miner",
		JS:     js,
		CSS:    css,
		HTML:   html,
		Colour: background,
	})

	alreadyRunning := false
	s := single.New("vertcoin-ocm")
	if err := s.CheckLock(); err != nil && err == single.ErrAlreadyRunning {
		alreadyRunning = true
	} else if err == nil {
		defer func() {
			err := s.TryUnlock()
			if err != nil {
				logging.Errorf("Error unlocking OCM: %v", err)
			}
		}()
	}

	backend, err := backend.NewBackend(alreadyRunning)
	if err != nil {
		logging.Errorf("Error creating Backend: %s", err.Error())
		panic(err)
	}
	networks.SetNetwork(backend.GetTestnet())

	go func() {
		backend.ResolveBackendServers()
		backend.BackendServerSelector()
	}()
	go backend.SelectP2PoolNode()

	backend.ResetPool()
	app.Bind(backend)
	err = app.Run()
	if err != nil {
		logging.Errorf("Error running app: %v", err)
	}
	backend.StopMining()

	tracking.Track(tracking.TrackingRequest{
		Category: "Lifecycle",
		Action:   "Shutdown",
	})

	tracking.Stop()
}
