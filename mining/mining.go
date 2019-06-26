package mining

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/keyfile"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/pools"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/vertcoin-project/one-click-miner-vnext/wallet"
	"github.com/wailsapp/wails"
)

type MinerCore struct {
	runtime            *wails.Runtime
	wal                *wallet.Wallet
	pool               pools.Pool
	refreshBalanceChan chan bool
}

func NewMinerCore() *MinerCore {
	return &MinerCore{
		refreshBalanceChan: make(chan bool),
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
	m.runtime.Events.Emit("checkStatus", "Checking stuff...")
	time.Sleep(1 * time.Second)
	m.runtime.Events.Emit("checkStatus", "Failed")
	if succeed {
		return "ok"
	} else {
		succeed = true //  succeed after retry
		return "Failure starting the miner:\n\nBlablabla"
	}

}

func (m *MinerCore) GetGPUs() []string {
	return []string{util.GetGPU()}
}

func (m *MinerCore) StartMining() bool {
	logging.Infof("Starting mining process...")

	// Default to P2Proxy for now
	m.pool = pools.NewP2Proxy(m.wal.Address)

	go func() {
		for {
			hashrate := rand.Float64() * float64(100)
			avgEarnings := hashrate / float64(600000) * float64(14400)
			m.runtime.Events.Emit("hashRate", fmt.Sprintf("%0.2f MH/s", hashrate))
			m.runtime.Events.Emit("avgEarnings", fmt.Sprintf("%0.2f VTC", avgEarnings))
			<-time.After(time.Second * 2)
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
			case <-m.refreshBalanceChan:
			case <-time.After(time.Minute * 5):
			}
		}
	}()

	return true
}

func (m *MinerCore) RefreshBalance() {

	m.refreshBalanceChan <- true
}

func (m *MinerCore) StopMining() bool {
	logging.Infof("Stopping mining process...")
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
