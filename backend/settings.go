package backend

import (
	"github.com/tidwall/buntdb"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
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

func (m *Backend) GetTestnet() bool {
	return m.getSetting("testnet")
}

func (m *Backend) SetTestnet(newTestnet bool) {
	if m.GetTestnet() != newTestnet {
		logging.Infof("Setting testnet to [%b]\n", newTestnet)
		m.setSetting("testnet", newTestnet)

		// Intentionally twice since blockheight is fetched from insight that differs per network
		logging.Infof("Setting network to testnet=%b\n", newTestnet)
		networks.SetNetwork(0, newTestnet)
		blockHeight := util.GetBlockHeight()
		logging.Infof("Setting network again with blockheight %d\n", blockHeight)
		networks.SetNetwork(blockHeight, newTestnet)

		logging.Infof("Calling WalletInitialized\n")
		m.WalletInitialized()
		logging.Infof("Done!")
	}
}

func (m *Backend) GetSkipVerthashExtendedVerify() bool {
	return m.getSetting("skipverthashverify")
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
