package backend

import (
	"fmt"
	"runtime"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/miners"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func (m *Backend) PerformChecks() string {
	m.runtime.Events.Emit("checkStatus", "rapidfail")
	if len(m.rapidFailures) > 0 {
		m.runtime.Events.Emit("checkStatus", "Failed")
		m.rapidFailures = make([]*miners.BinaryRunner, 0) // Clear the failures
		return "One or more of your miner binaries are showing rapid failures (immediately stop after starting). Please enable debugging under the Settings tab and then Save & Restart. Use the debug.log to learn more about what might be going on."
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
		gpustring += g.OSName
	}

	tracking.Track(tracking.TrackingRequest{
		Category: "EnumerateGPUs",
		Action:   "Success",
		Name:     gpustring,
	})

	if compat == 0 {
		return fmt.Errorf("No compatible GPUs detected")
	}
	return nil
}

func (m *Backend) CreateMinerBinaries() ([]*miners.BinaryRunner, error) {
	binaries := miners.GetMinerBinaries()
	gpus := util.GetGPUs()
	closedSource := m.GetClosedSource()
	brs := []*miners.BinaryRunner{}
	for _, b := range binaries {
		match := false
		if b.Platform == runtime.GOOS {
			for _, g := range gpus {
				if g.Type == b.GPUType {
					if b.ClosedSource == closedSource {
						match = true
						break
					}
				}
			}
		}

		if match {
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
		return nil, fmt.Errorf("Could not find compatible miner binaries")
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
