package backend

import (
	"github.com/vertcoin-project/one-click-miner-vnext/tracking"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
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
