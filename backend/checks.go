package backend

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"time"

	verthash "github.com/gertjaap/verthash-go"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/miners"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func (m *Backend) PerformChecks() string {
	m.runtime.Events.Emit("checkStatus", "rapidfail")
	if len(m.rapidFailures) > 0 {
		m.runtime.Events.Emit("checkStatus", "Failed")
		m.rapidFailures = make([]*miners.BinaryRunner, 0) // Clear the failures
		return "Rapid failures: Your GPU is likely incompatible; check FAQ for supported hardware. If compatible, GPU overclocks or antivirus may be the cause."
	}

	m.runtime.Events.Emit("checkStatus", "compatibility")
	err := m.CheckGPUCompatibility()
	if err != nil {
		tracking.Track(tracking.TrackingRequest{
			Category: "PerformChecks",
			Action:   "CheckGPUCompatibilityError",
			Name:     err.Error(),
		})
		m.runtime.Events.Emit("checkStatus", "Failed")
		return err.Error()
	}

	m.runtime.Events.Emit("checkStatus", "installing_miners")
	err = m.InstallMinerBinaries()
	if err != nil {
		tracking.Track(tracking.TrackingRequest{
			Category: "PerformChecks",
			Action:   "InstallMinerBinariesError",
			Name:     err.Error(),
		})
		m.runtime.Events.Emit("checkStatus", "Failed")
		return err.Error()
	}

	m.runtime.Events.Emit("checkStatus", "verthash")
	verthashFile := filepath.Join(util.DataDirectory(), "verthash.dat")

	doneChan := make(chan bool, 1)
	progress := make(chan float64, 1)

	go func() {
		if m.GetSkipVerthashExtendedVerify() {
			err = verthash.MakeVerthashDatafileIfNotExistsWithProgress(verthashFile, progress)
		} else {
			err = verthash.EnsureVerthashDatafileWithProgress(verthashFile, progress)
		}
		doneChan <- true
	}()

	for {
		done := false
		select {
		case done = <-doneChan:
			break
		case prog := <-progress:
			m.runtime.Events.Emit("verthashProgress", prog*100)
			break
		}
		if done {
			break
		}
	}

	if err != nil {
		errorString := fmt.Sprintf("Failed to create or verify Verthash data file: %s", err.Error())
		m.runtime.Events.Emit("checkStatus", "Failed")
		return errorString
	}

	for !m.p2poolNodeSelected {
		time.Sleep(time.Second)
  }
	for networks.Active.OCMBackend == "" {
		time.Sleep(500 * time.Millisecond)
  }
  
	args := m.GetArgs()

	for _, br := range m.minerBinaries {
		err := br.MinerImpl.Configure(args)
		if err != nil {
			errorString := fmt.Sprintf("Failure to configure %s: The data directory may have to be excluded from your antivirus. Check FAQ.", br.MinerBinary.MainExecutableName)
			tracking.Track(tracking.TrackingRequest{
				Category: "PerformChecks",
				Action:   "ConfigureError",
				Name:     errorString,
			})
			m.runtime.Events.Emit("checkStatus", "Failed")
			return errorString
		}

		if br.MinerImpl.AvailableGPUs() == 0 {
			m.runtime.Events.Emit("checkStatus", "Failed")
			return "Miner software reported no compatible GPUs. Check FAQ for supported hardware and ensure your GPU drivers are up to date."
		}
	}

	tracking.Track(tracking.TrackingRequest{
		Category: "PerformChecks",
		Action:   "Success",
	})

	return "ok"
}

func (m *Backend) CheckGPUCompatibility() error {
	gpus := util.GetGPUs()
	compat := 0
	gpustring := ""
	for _, g := range gpus {
		if g.Type != util.GPUTypeOther {
			compat++
		}
		if gpustring != "" {
			gpustring += " / "
		}
		gpustring += g.OSName
	}

	tracking.Track(tracking.TrackingRequest{
		Category: "EnumerateGPUs",
		Action:   "Success",
		Name:     gpustring,
	})

	if compat == 0 {
		return fmt.Errorf("No compatible GPUs detected\n\nGPUs Found:\n%s - Check FAQ for supported hardware and ensure your GPU drivers are up to date.", gpustring)
	}
	return nil
}

func (m *Backend) CreateMinerBinaries() ([]*miners.BinaryRunner, error) {
	binaries := miners.GetMinerBinaries()
	gpus := util.GetGPUs()
	closedSource := m.GetClosedSource()
	testnet := m.GetTestnet()
	brs := []*miners.BinaryRunner{}
	for _, b := range binaries {
		match := false
		if b.Platform == runtime.GOOS {
			for _, g := range gpus {
				if g.Type == b.GPUType {
					if b.ClosedSource == closedSource {
						if b.Testnet == testnet {
							match = true
						}
					}
				}
			}
		}

		if match {
			if b.MultiGPUMiner {
				alreadyPresent := false
				for _, br := range brs {
					if br.MinerBinary.MainExecutableName == b.MainExecutableName {
						alreadyPresent = true
						break
					}
				}
				if alreadyPresent {
					logging.Debugf("Not adding already present multi-gpu binary [%s] again\n", b.MainExecutableName)
					continue
				}
			}
			logging.Debugf("Found compatible binary [%s] for [%s/%d] (Closed source: %t)\n", b.MainExecutableName, b.Platform, b.GPUType, b.ClosedSource)
			br, err := miners.NewBinaryRunner(b, m.prerequisiteInstall)
			if err != nil {
				return nil, err
			}
			br.Debug = m.GetDebugging()
			brs = append(brs, br)
		} else {
			logging.Debugf("Found incompatible binary [%s] for [%s/%d] (Closed source: %t)\n", b.MainExecutableName, b.Platform, b.GPUType, b.ClosedSource)
		}
	}

	if len(brs) == 0 {
		return nil, fmt.Errorf("Could not find compatible miner binaries - Check FAQ for supported hardware and ensure your GPU drivers are up to date.")
	}

	return brs, nil
}

func (m *Backend) InstallMinerBinaries() error {
	var err error
	m.minerBinaries, err = m.CreateMinerBinaries()
	if err != nil {
		return err
	}

	for _, br := range m.minerBinaries {
		err := br.Install()
		if err != nil {
			return err
		}
	}
	return nil
}

// Will be run at startup
// Additionally it can be run if the backend returns an error after startup
func (m *Backend) BackendServerSelector() {
	// Pick a random backend off the list
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(len(networks.Active.BackendServers))

	// Run a simple check to see if the backend is up and returned data isn't nonsense
	// If the backend is bad, go through the list until a suitable one is found
	for range networks.Active.BackendServers {
		b := util.CheckBackendStatus(networks.Active.BackendServers[n])
		if b {
			// If backend is up and return data other than 0, save it in networks.Active
			networks.Active.OCMBackend = networks.Active.BackendServers[n]
			logging.Infof("Using backend: %s\n", networks.Active.OCMBackend)
			return
		}
		n += 1
		if n == len(networks.Active.BackendServers) {
			n = 0
		}
	}
	// We'll only ever get here if all backends are unreachable..
	networks.Active.OCMBackend = networks.Active.BackendServers[0]
	logging.Errorf("No working backend could be found..\n")
}
