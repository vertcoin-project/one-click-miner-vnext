package backend

import (
	"fmt"
	"strconv"

	"github.com/tidwall/buntdb"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/payouts"
	"github.com/vertcoin-project/one-click-miner-vnext/pools"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func (m *Backend) getSetting(name string) bool {
	setting := "0"
	m.settings.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(name)
		setting = v
		return err
	})
	return setting == "1"
}

func (m *Backend) setSetting(name string, value bool) {
	setting := "0"
	if value {
		setting = "1"
	}
	m.settings.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(name, setting, nil)
		return err
	})
}

func (m *Backend) setIntSetting(name string, value int) {
	setting := fmt.Sprintf("%d", value)
	m.settings.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(name, setting, nil)
		return err
	})
}

func (m *Backend) getIntSetting(name string) int {
	setting := "0"
	m.settings.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(name)
		setting = v
		return err
	})
	i, _ := strconv.Atoi(setting)
	return i
}

func (m *Backend) setStringSetting(name string, value string) {
	setting := fmt.Sprintf("%s", value)
	m.settings.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(name, setting, nil)
		return err
	})
}

func (m *Backend) getStringSetting(name string) string {
	setting := ""
	m.settings.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(name)
		setting = v
		return err
	})
	return setting
}

func (m *Backend) GetPool() int {
	pool := m.getIntSetting("pool")
	if pool == 0 {
		if m.GetTestnet() {
			return 2 // Default P2Pool on testnet
		}
		return 3 // Default Hashalot on mainnet (for now...)
	}
	return pool
}

func (m *Backend) SetPool(pool int) {
	if m.GetPool() != pool {
		m.setIntSetting("pool", pool)
		m.ResetPool()
		logging.Infof("Calling WalletInitialized\n")
		m.WalletInitialized()
		logging.Infof("Done!")
	}
}

type PoolChoice struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (m *Backend) GetPools() []PoolChoice {
	pc := make([]PoolChoice, 0)
	for _, p := range pools.GetPools(m.Address(), m.GetTestnet()) {
		pc = append(pc, PoolChoice{
			ID:   p.GetID(),
			Name: fmt.Sprintf("%s (%0.1f%% fee)", p.GetName(), p.GetFee()),
		})
	}
	return pc
}

func (m *Backend) GetPayout() int {
	payout := m.getIntSetting("payout")
	if payout == 0 {
		if m.GetTestnet() {
			return 1 // Default Vertcoin on testnet
		}
		return 1 // Default Vertcoin on mainnet
	}
	return payout
}

func (m *Backend) SetPayout(payout int) {
	if m.GetPayout() != payout {
		m.setIntSetting("payout", payout)
		m.ResetPayout()
	}
}

type PayoutChoice struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (m *Backend) GetPayouts() []PayoutChoice {
	pc := make([]PayoutChoice, 0)
	for _, p := range payouts.GetPayouts(m.GetTestnet()) {
		pc = append(pc, PayoutChoice{
			ID:   p.GetID(),
			Name: p.GetName(),
		})
	}
	return pc
}

func (m *Backend) GetZergpoolAddress() string {
	return m.getStringSetting("zergpoolAddress")
}

func (m *Backend) SetZergpoolAddress(newZergpoolAddress string) {
	logging.Infof("Setting Zergpool address to [%s]\n", newZergpoolAddress)
	m.zergpoolAddress = newZergpoolAddress
	m.setStringSetting("zergpoolAddress", newZergpoolAddress)
}

// TODO: Improve address validation
func (m *Backend) ValidZergpoolAddress() bool {
	zergpoolAddress := m.zergpoolAddress
	if zergpoolAddress != "" {
		return true
	}
	return false
}

func (m *Backend) GetTestnet() bool {
	return false // Testnet is not necessary - return false
	//return m.getSetting("testnet")
}

func (m *Backend) SetTestnet(newTestnet bool) {
	if m.GetTestnet() != newTestnet {
		logging.Infof("Setting testnet to [%b]\n", newTestnet)
		m.setSetting("testnet", newTestnet)

		logging.Infof("Setting network to testnet=%b\n", newTestnet)
		networks.SetNetwork(newTestnet)

		logging.Infof("Calling WalletInitialized\n")
		m.WalletInitialized()
		logging.Infof("Done!")
	}
}

func (m *Backend) GetSkipVerthashExtendedVerify() bool {
	return false // Verification is fast and chain split is possible if datafile is wrong - return false
	//return m.getSetting("skipverthashverify")
}

func (m *Backend) SetSkipVerthashExtendedVerify(newVerthashVerify bool) {
	logging.Infof("Setting skip verthash verify to [%b]\n", newVerthashVerify)
	m.setSetting("skipverthashverify", newVerthashVerify)
}

func (m *Backend) GetClosedSource() bool {
	return false // No closed source Verthash miners - return false
	//return m.getSetting("closedsource")
}

func (m *Backend) SetClosedSource(newClosedSource bool) {
	logging.Infof("Setting closed source to [%b]\n", newClosedSource)
	m.setSetting("closedsource", newClosedSource)
}

func (m *Backend) GetDebugging() bool {
	return m.getSetting("debugging")
}

func (m *Backend) SetDebugging(newDebugging bool) {
	logging.Infof("Setting debugging to [%b]\n", newDebugging)
	m.setSetting("debugging", newDebugging)
}

func (m *Backend) GetAutoStart() bool {
	return util.GetAutoStart()
}

func (m *Backend) SetAutoStart(newAutoStart bool) {
	util.SetAutoStart(newAutoStart)
}

func (m *Backend) GetVersion() string {
	return tracking.GetVersion()
}

func (m *Backend) PrerequisiteProxyLoop() {
	for pi := range m.prerequisiteInstall {
		send := "0"
		if pi {
			send = "1"
		}
		m.runtime.Events.Emit("prerequisiteInstall", send)
	}
}
