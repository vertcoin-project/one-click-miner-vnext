package mining

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tidwall/buntdb"

	"github.com/btcsuite/btcd/wire"
	"github.com/vertcoin-project/one-click-miner-vnext/keyfile"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/miners"
	"github.com/vertcoin-project/one-click-miner-vnext/pools"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/vertcoin-project/one-click-miner-vnext/wallet"
	"github.com/wailsapp/wails"
)

type MinerCore struct {
	runtime             *wails.Runtime
	wal                 *wallet.Wallet
	settings            *buntdb.DB
	pendingSweep        []*wire.MsgTx
	minerBinaries       []*miners.BinaryRunner
	rapidFailures       []*miners.BinaryRunner
	pool                pools.Pool
	refreshBalanceChan  chan bool
	refreshHashChan     chan bool
	refreshRunningState chan bool
	stopMonitoring      chan bool
	stopHash            chan bool
	stopBalance         chan bool
	stopRunningState    chan bool
	prerequisiteInstall chan bool
}

func NewMinerCore() (*MinerCore, error) {
	db, err := buntdb.Open(filepath.Join(util.DataDirectory(), "settings.db"))
	if err != nil {
		return nil, err
	}

	return &MinerCore{
		settings:            db,
		refreshBalanceChan:  make(chan bool),
		refreshHashChan:     make(chan bool),
		refreshRunningState: make(chan bool),
		stopHash:            make(chan bool),
		stopBalance:         make(chan bool),
		stopRunningState:    make(chan bool),
		stopMonitoring:      make(chan bool),
		prerequisiteInstall: make(chan bool),
		minerBinaries:       []*miners.BinaryRunner{},
		rapidFailures:       []*miners.BinaryRunner{},
	}, nil
}

func (m *MinerCore) WailsInit(runtime *wails.Runtime) error {
	// Save runtime
	m.runtime = runtime

	go func() {
		for pi := range m.prerequisiteInstall {
			send := "0"
			if pi {
				send = "1"
			}
			m.runtime.Events.Emit("prerequisiteInstall", send)
		}
	}()

	return nil
}

func (m *MinerCore) getSetting(name string) bool {
	setting := "0"
	m.settings.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(name)
		setting = v
		return err
	})
	return setting == "1"
}

func (m *MinerCore) setSetting(name string, value bool) {
	setting := "0"
	if value {
		setting = "1"
	}
	m.settings.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(name, setting, nil)
		return err
	})
}

func (m *MinerCore) GetClosedSource() bool {
	return m.getSetting("closedsource")
}

func (m *MinerCore) SetClosedSource(newClosedSource bool) {
	logging.Infof("Setting closed source to [%b]\n", newClosedSource)
	m.setSetting("closedsource", newClosedSource)
}

func (m *MinerCore) GetDebugging() bool {
	return m.getSetting("debugging")
}

func (m *MinerCore) SetDebugging(newDebugging bool) {
	logging.Infof("Setting debugging to [%b]\n", newDebugging)
	m.setSetting("debugging", newDebugging)
}

func (m *MinerCore) GetVersion() string {
	return tracking.GetVersion()
}

func (m *MinerCore) WalletInitialized() int {
	logging.Infof("Checking wallet..")
	checkWallet := 0
	if keyfile.KeyFileValid() {
		checkWallet = 1
	}
	wal, err := wallet.NewWallet(keyfile.GetAddress()) // TODO: Replace with actual address!
	if err != nil {
		logging.Errorf("Error initializing wallet: %s", err.Error())
	}
	m.wal = wal
	logging.Infof("Wallet initialized: %d", checkWallet)
	return checkWallet
}

var succeed = false

func (m *MinerCore) PerformChecks() string {
	m.runtime.Events.Emit("checkStatus", "Checking for rapid failure occurrences")
	if len(m.rapidFailures) > 0 {
		m.runtime.Events.Emit("checkStatus", "Failed")
		m.rapidFailures = make([]*miners.BinaryRunner, 0) // Clear the failures
		return "One or more of your miner binaries are showing rapid failures (immediately stop after starting). Please enable debugging under the Settings tab and then Save & Restart. Use the debug.log to learn more about what might be going on."
	}

	m.runtime.Events.Emit("checkStatus", "Checking GPU compatibility...")
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

	m.runtime.Events.Emit("checkStatus", "Installing miners...")
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

func (m *MinerCore) EnableTracking() {
	tracking.Enable()
}

func (m *MinerCore) DisableTracking() {
	tracking.Disable()
}

func (m *MinerCore) TrackingEnabled() string {
	if tracking.IsEnabled() {
		return "1"
	}
	return "0"
}

func (m *MinerCore) CheckGPUCompatibility() error {
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

func (m *MinerCore) CreateMinerBinaries() ([]*miners.BinaryRunner, error) {
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

func (m *MinerCore) InstallMinerBinaries() error {
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

func (m *MinerCore) StartMining() bool {
	logging.Infof("Starting mining process...")

	tracking.Track(tracking.TrackingRequest{
		Category: "Mining",
		Action:   "Start",
	})

	// Default to P2Proxy for now
	m.pool = pools.NewP2Proxy(m.wal.Address)
	args := miners.BinaryArguments{
		StratumUrl:      m.pool.GetStratumUrl(),
		StratumUsername: m.pool.GetUsername(),
		StratumPassword: m.pool.GetPassword(),
	}

	startProcessMonitoring := make(chan bool)

	go func() {
		<-startProcessMonitoring
		continueLoop := true
		for continueLoop {
			newMinerBinaries := make([]*miners.BinaryRunner, 0)
			for _, br := range m.minerBinaries {
				if br.CheckRunning() == miners.RunningStateRapidFail {
					m.rapidFailures = append(m.rapidFailures, br)
					m.runtime.Events.Emit("minerRapidFail", br.MinerBinary.MainExecutableName)

				} else {
					newMinerBinaries = append(newMinerBinaries, br)
				}
			}

			m.minerBinaries = newMinerBinaries

			select {
			case <-m.stopMonitoring:
				continueLoop = false
			case <-time.After(time.Second):
			}
		}
		logging.Infof("Stopped monitoring thread")
	}()

	go func() {
		cycles := 0
		nhr := util.GetNetHash()
		continueLoop := true
		for continueLoop {
			cycles++
			if cycles > 150 {
				// Don't refresh this every time since we refresh it every second
				// and this pulls from Insight. Every 150s is fine (every block)
				nhr = util.GetNetHash()
				cycles = 0
			}
			hr := uint64(0)
			for _, br := range m.minerBinaries {
				hr += br.HashRate()
			}
			hashrate := float64(hr) / float64(1000000)
			hashrateUnit := "MH/s"
			if hashrate > 1000 {
				hashrate /= 1000
				hashrateUnit = "GH/s"
			}
			m.runtime.Events.Emit("hashRate", fmt.Sprintf("%0.2f %s", hashrate, hashrateUnit))
			hashrateUnit = "GH/s"
			if hashrate > 1000 {
				hashrate /= 1000
				hashrateUnit = "TH/s"
			}

			netHash := float64(nhr) / float64(1000000000)

			m.runtime.Events.Emit("networkHashRate", fmt.Sprintf("%0.2f %s", netHash, hashrateUnit))

			avgEarning := float64(hr) / float64(nhr) * float64(14400) // 14400 = Emission per day. Need to adjust for halving

			m.runtime.Events.Emit("avgEarnings", fmt.Sprintf("%0.2f VTC", avgEarning))
			select {
			case <-m.stopHash:
				continueLoop = false
			case <-m.refreshHashChan:
			case <-time.After(time.Second):
			}
		}
	}()

	go func() {
		continueLoop := true
		for continueLoop {

			logging.Infof("Updating balance...")
			m.wal.Update()
			b, bi := m.wal.GetBalance()
			pb := m.pool.GetPendingPayout()
			m.runtime.Events.Emit("balance", fmt.Sprintf("%0.8f", float64(b)/float64(100000000)))
			m.runtime.Events.Emit("balanceImmature", fmt.Sprintf("%0.8f", float64(bi)/float64(100000000)))
			m.runtime.Events.Emit("balancePendingPool", fmt.Sprintf("%0.8f", float64(pb)/float64(100000000)))
			select {
			case <-m.stopBalance:
				continueLoop = false
			case <-m.refreshBalanceChan:
			case <-time.After(time.Minute * 5):
			}
		}
	}()

	go func() {
		continueLoop := true
		for continueLoop {

			runningProcesses := 0
			for _, br := range m.minerBinaries {
				if br.IsRunning() {
					runningProcesses++
				}
			}

			m.runtime.Events.Emit("runningMiners", runningProcesses)

			timeout := time.Second * 1
			if runningProcesses > 0 {
				timeout = time.Second * 10
			}
			select {
			case <-m.stopRunningState:
				continueLoop = false
			case <-m.refreshRunningState:
			case <-time.After(timeout):
			}
		}
	}()

	for _, br := range m.minerBinaries {
		err := br.MinerImpl.Configure(args)
		if err != nil {
			logging.Errorf("Failure to configure %s: %s\n", br.MinerBinary.MainExecutableName, err.Error())
			return false
		}
		err = br.Start(args)
		if err != nil {
			logging.Errorf("Failure to start %s: %s\n", br.MinerBinary.MainExecutableName, err.Error())
			return false
		}
	}

	startProcessMonitoring <- true

	return true
}

func (m *MinerCore) RefreshBalance() {

	m.refreshBalanceChan <- true
}

func (m *MinerCore) RefreshHashrate() {

	m.refreshHashChan <- true
}

func (m *MinerCore) RefreshRunningState() {

	m.refreshRunningState <- true
}

func (m *MinerCore) SendSweep(password string) []string {
	tracking.Track(tracking.TrackingRequest{
		Category: "Sweep",
		Action:   "Send",
	})

	txids := make([]string, 0)

	for _, s := range m.pendingSweep {
		err := m.wal.SignMyInputs(s, password)
		if err != nil {
			return []string{err.Error()}
		}

		txHash, err := m.wal.Send(s)
		if err != nil {
			return []string{err.Error()}
		}
		txids = append(txids, txHash)
	}

	m.pendingSweep = nil

	logging.Debugf("Transaction(s) sent! TXIDs: %v\n", txids)

	return txids

}

func (m *MinerCore) ShowTx(txid string) {
	util.OpenBrowser(fmt.Sprintf("https://insight.vertcoin.org/tx/%s", txid))
}

func (m *MinerCore) ReportIssue() {
	util.OpenBrowser("https://github.com/vertcoin-project/one-click-miner-vnext/issues/new")
}

func (m *MinerCore) PrepareSweep(addr string) string {
	tracking.Track(tracking.TrackingRequest{
		Category: "Sweep",
		Action:   "Prepare",
	})

	txs, err := m.wal.PrepareSweep(addr)
	if err != nil {
		return err.Error()
	}

	m.pendingSweep = txs
	val := float64(0)
	for _, tx := range txs {
		val += (float64(tx.TxOut[0].Value) / float64(100000000))
	}
	m.runtime.Events.Emit("createTransactionResult", fmt.Sprintf("%0.8f VTC in %d transaction(s)", val, len(txs)))
	return ""
}

func (m *MinerCore) StopMining() bool {
	tracking.Track(tracking.TrackingRequest{
		Category: "Mining",
		Action:   "Stop",
	})
	select {
	case m.stopMonitoring <- true:
	default:
	}
	logging.Infof("Stopping mining process...")
	for _, br := range m.minerBinaries {
		br.Stop()
	}
	select {
	case m.stopBalance <- true:
	default:
	}
	select {
	case m.stopHash <- true:
	default:
	}
	select {
	case m.stopRunningState <- true:
	default:
	}
	return true
}

func (m *MinerCore) Address() string {
	return keyfile.GetAddress()
}

func (m *MinerCore) InitWallet(password string) bool {
	tracking.Track(tracking.TrackingRequest{
		Category: "Wallet",
		Action:   "Initialize",
	})

	err := keyfile.CreateKeyFile(password)
	if err == nil {
		m.WalletInitialized()
		return true
	}
	logging.Errorf("Error: %s", err.Error())
	return false
}
