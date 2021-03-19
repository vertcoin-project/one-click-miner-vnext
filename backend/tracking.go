package backend

import (
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func (m *Backend) EnableTracking() {
	tracking.Enable()
}

func (m *Backend) DisableTracking() {
	tracking.Disable()
}

func (m *Backend) TrackingEnabled() string {
	if tracking.IsEnabled() {
		return "1"
	}
	return "0"
}

func (m *Backend) ReportIssue() {
	util.OpenBrowser("https://github.com/vertcoin-project/one-click-miner-vnext/issues/new")
}
