package backend

import (
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func (m *Backend) BlockHeight() int64 {
	return util.GetBlockHeight()
}

func (m *Backend) LaunchForkSite() {
	util.OpenBrowser("https://wenvtcfork.xyz/")
}
