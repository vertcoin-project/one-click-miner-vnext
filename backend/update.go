package backend

import (
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (m *Backend) UpdateAvailable() bool {
	r, _ := util.GetLatestRelease()

	lastVersion := util.VersionStringToNumeric(r.Tag)
	myVersion := util.VersionStringToNumeric(tracking.GetVersion())

	return lastVersion > myVersion
}

func (m *Backend) VersionDetails() []string {
	r, _ := util.GetLatestRelease()
	return []string{r.Tag, r.Body, r.URL}
}

func (m *Backend) UpdateLoop() {
	for {
		stopUpdate := false
		select {
		case stopUpdate = <-m.stopUpdate:
		case <-time.After(time.Second * 15):
		}

		if stopUpdate {
			break
		}

		runtime.EventsEmit(m.ctx,"updateAvailable", m.UpdateAvailable())
	}
}
