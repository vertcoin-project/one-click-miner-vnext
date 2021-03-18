package backend

import (
	"github.com/vertiond/verthash-one-click-miner/tracking"
	"github.com/vertiond/verthash-one-click-miner/util"
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
	util.OpenBrowser("https://github.com/vertiond/verthash-one-click-miner/issues/new")
}

func (m *Backend) PayoutInformation() {
	m.pool.OpenBrowserPayoutInfo(m.GetCurrentMiningAddress())
}
