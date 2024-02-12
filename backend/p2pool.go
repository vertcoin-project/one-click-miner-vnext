package backend

import (
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/ping"
)

func (m *Backend) SelectP2PoolNode() {
	logging.Infof("Finding best P2Pool node...")
	ping.GetSelectedNode(m.getSetting("testnet"))
	logging.Infof("Found best P2Pool node: %v", ping.Selected.P2PoolURL)
	m.p2poolNodeSelected = true
}
