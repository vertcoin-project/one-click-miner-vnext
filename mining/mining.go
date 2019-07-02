package mining

import (
	"fmt"
	"runtime"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/vertcoin-project/one-click-miner-vnext/keyfile"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/miners"
	"github.com/vertcoin-project/one-click-miner-vnext/pools"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/vertcoin-project/one-click-miner-vnext/wallet"
	"github.com/wailsapp/wails"
)

type MinerCore struct {
	runtime            *wails.Runtime
	wal                *wallet.Wallet
	pendingSweep       *wire.MsgTx
	minerBinaries      []*miners.BinaryRunner
	pool               pools.Pool
	refreshBalanceChan chan bool
	refreshHashChan    chan bool
	stopHash           chan bool
	stopBalance        chan bool
}

func NewMinerCore() *MinerCore {
	return &MinerCore{
		refreshBalanceChan: make(chan bool),
		refreshHashChan:    make(chan bool),
		stopHash:           make(chan bool),
		stopBalance:        make(chan bool),
		minerBinaries:      []*miners.BinaryRunner{},
	}
}

func (m *MinerCore) WailsInit(runtime *wails.Runtime) error {
	// Save runtime
	m.runtime = runtime

	return nil
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
	m.runtime.Events.Emit("checkStatus", "Checking GPU compatibility...")
	err := m.CheckGPUCompatibility()
	if err != nil {
		m.runtime.Events.Emit("checkStatus", "Failed")
		return err.Error()
	}

	m.runtime.Events.Emit("checkStatus", "Installing miners...")
	err = m.InstallMinerBinaries()
	if err != nil {
		m.runtime.Events.Emit("checkStatus", "Failed")
		return err.Error()
	}

	return "ok"
}

func (m *MinerCore) CheckGPUCompatibility() error {
	gpus := util.GetGPUs()
	compat := 0
	for _, g := range gpus {
		if g.Type != util.GPUTypeOther {
			compat++
		}
	}
	if compat == 0 {
		return fmt.Errorf("No compatible GPUs detected")
	}
	return nil
}

func (m *MinerCore) CreateMinerBinaries() ([]*miners.BinaryRunner, error) {
	binaries := miners.GetMinerBinaries()
	gpus := util.GetGPUs()
	brs := []*miners.BinaryRunner{}
	for _, b := range binaries {
		match := false
		if b.Platform == runtime.GOOS {
			for _, g := range gpus {
				if g.Type == b.GPUType {
					match = true
					break
				}
			}
		}

		if match {
			logging.Debugf("Found compatible binary [%s] for [%s/%d]\n", b.MainExecutableName, b.Platform, b.GPUType)
			br, err := miners.NewBinaryRunner(b)
			if err != nil {
				return nil, err
			}
			brs = append(brs, br)
		} else {
			logging.Debugf("Found incompatible binary [%s] for [%s/%d]\n", b.MainExecutableName, b.Platform, b.GPUType)
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

	// Default to P2Proxy for now
	m.pool = pools.NewP2Proxy(m.wal.Address)
	args := miners.BinaryArguments{
		StratumUrl:      m.pool.GetStratumUrl(),
		StratumUsername: m.pool.GetUsername(),
		StratumPassword: m.pool.GetPassword(),
	}

	go func() {
		cycles := 0
		nhr := util.GetNetHash()
		for {
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
				break
			case <-m.refreshHashChan:
			case <-time.After(time.Second):
			}
		}
	}()

	go func() {
		for {
			logging.Infof("Updating balance...")
			m.wal.Update()
			b, bi := m.wal.GetBalance()
			pb := m.pool.GetPendingPayout()
			m.runtime.Events.Emit("balance", fmt.Sprintf("%0.8f", float64(b)/float64(100000000)))
			m.runtime.Events.Emit("balanceImmature", fmt.Sprintf("%0.8f", float64(bi)/float64(100000000)))
			m.runtime.Events.Emit("balancePendingPool", fmt.Sprintf("%0.8f", float64(pb)/float64(100000000)))
			select {
			case <-m.stopBalance:
				break
			case <-m.refreshBalanceChan:
			case <-time.After(time.Minute * 5):
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

	return true
}

func (m *MinerCore) RefreshBalance() {

	m.refreshBalanceChan <- true
}

func (m *MinerCore) RefreshHashrate() {

	m.refreshHashChan <- true
}

func (m *MinerCore) SendSweep(password string) string {
	err := m.wal.SignMyInputs(m.pendingSweep, password)
	if err != nil {
		return err.Error()
	}

	txHash, err := m.wal.Send(m.pendingSweep)
	if err != nil {
		return err.Error()
	}
	m.pendingSweep = nil

	logging.Debugf("Transaction sent! TXID: %s\n", txHash)

	return txHash

}

func (m *MinerCore) ShowTx(txid string) {
	util.OpenBrowser(fmt.Sprintf("https://insight.vertcoin.org/tx/%s", txid))
}

func (m *MinerCore) PrepareSweep(addr string) string {
	tx, err := m.wal.PrepareSweep(addr)
	if err != nil {
		return err.Error()
	}

	m.pendingSweep = tx

	val := float64(tx.TxOut[0].Value) / float64(100000000)

	m.runtime.Events.Emit("createTransactionResult", fmt.Sprintf("%0.8f VTC", val))
	return ""
}

func (m *MinerCore) StopMining() bool {
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
	return true
}

func (m *MinerCore) Address() string {
	return keyfile.GetAddress()
}

func (m *MinerCore) InitWallet(password string) bool {
	err := keyfile.CreateKeyFile(password)
	if err == nil {
		m.WalletInitialized()
		return true
	}
	logging.Errorf("Error: %s", err.Error())
	return false
}
