package backend

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/tidwall/buntdb"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/pools"
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func (m *Backend) getSetting(name string) bool {
	setting := "0"
	err := m.settings.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(name)
		setting = v
		return err
	})
	if err != nil {
		logging.Errorf("Error in getSetting(%s): %v", name, err)
	}
	return setting == "1"
}

func (m *Backend) setSetting(name string, value bool) {
	setting := "0"
	if value {
		setting = "1"
	}
	err := m.settings.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(name, setting, nil)
		return err
	})
	if err != nil {
		logging.Errorf("Error in setSetting(%s): %v", name, err)
	}
}

func (m *Backend) setIntSetting(name string, value int) {
	setting := fmt.Sprintf("%d", value)
	err := m.settings.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(name, setting, nil)
		return err
	})
	if err != nil {
		logging.Errorf("Error in setSetting(%s): %v", name, err)
	}
}

func (m *Backend) getIntSetting(name string) int {
	setting := "0"
	err := m.settings.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(name)
		setting = v
		return err
	})
	if err != nil {
		logging.Errorf("Error in getSetting(%s): %v", name, err)
	}
	i, _ := strconv.Atoi(setting)
	return i
}

// getStringSetting returns "" for a setting that was never saved
func (m *Backend) getStringSetting(name string) string {
	setting := ""
	err := m.settings.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(name)
		setting = v
		return err
	})
	if err != nil && err != buntdb.ErrNotFound {
		logging.Errorf("Error in getSetting(%s): %v", name, err)
	}
	return setting
}

func (m *Backend) setStringSetting(name string, value string) {
	err := m.settings.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(name, value, nil)
		return err
	})
	if err != nil {
		logging.Errorf("Error in setSetting(%s): %v", name, err)
	}
}

func (m *Backend) GetPool() int {
	pool := m.getIntSetting("pool")
	if pool != 0 && !pools.HasPool(pool, m.Address()) {
		// The saved pool was removed from the pool list, so pick another one.
		logging.Warnf("Pool %d is no longer available, choosing another\n", pool)
		pool = 0
	}
	if pool == 0 {
		// Default to a random pool
		pools := pools.GetPools(m.Address())
		pool := pools[rand.Intn(len(pools))].GetID()
		// Save this setting immediately so that we don't get
		// a different random pool in future calls to GetPool().
		m.setIntSetting("pool", pool)
		return pool
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

func (m *Backend) SetEnableIntegrated(enabled bool) {
	m.setSetting("enableIntegrated", enabled)
}

func (m *Backend) GetEnableIntegrated() bool {
	return m.getSetting("enableIntegrated")
}

type PoolChoice struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (m *Backend) GetPools() []PoolChoice {
	pc := make([]PoolChoice, 0)
	for _, p := range pools.GetPools(m.Address()) {
		pc = append(pc, PoolChoice{
			ID:   p.GetID(),
			Name: fmt.Sprintf("%s (%0.2f%% fee)", p.GetName(), p.GetFee()),
		})
	}
	return pc
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
