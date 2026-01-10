package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/marcsauter/single"
	"github.com/vertcoin-project/one-click-miner-vnext/backend"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

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

	backendInstance, err := backend.NewBackend(alreadyRunning)
	if err != nil {
		logging.Errorf("Error creating Backend: %s", err.Error())
		panic(err)
	}
	networks.SetNetwork(backendInstance.GetTestnet())

	err = wails.Run(&options.App{
		Title:  "Vertcoin One Click Miner",
		Width:  800,
		Height: 400,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 19, G: 19, B: 19, A: 1},
		OnStartup: func(ctx context.Context) {
			backendInstance.SetContext(ctx)
			go backendInstance.BackendServerSelector()
			go backendInstance.SelectP2PoolNode()
			backendInstance.ResetPool()
		},
		OnShutdown: func(ctx context.Context) {
			backendInstance.StopMining()
			tracking.Track(tracking.TrackingRequest{
				Category: "Lifecycle",
				Action:   "Shutdown",
			})
			tracking.Stop()
		},
		Bind: []interface{}{
			backendInstance,
		},
		Linux: &linux.Options{
			Icon:             []byte{}, // Will load icon if needed
			WebviewGpuPolicy: linux.WebviewGpuPolicyNever,
			ProgramName:      "vertcoin-ocm",
		},
	})
	if err != nil {
		logging.Errorf("Error running app: %v", err)
	}
}
