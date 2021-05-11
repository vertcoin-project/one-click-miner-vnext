package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &zpool{}

type zpool struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func Newzpool(addr string) *zpool {
	return &zpool{Address: addr}
}

func (p *zpool) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://zpool.ca/api/wallet?address=%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	vtc, ok := jsonPayload["unpaid"].(float64)
	if !ok {
		return 0
	}
	vtc *= 100000000
	return uint64(vtc)
}

func (p *zpool) GetStratumUrl() string {
	return "stratum+tcp://verthash.mine.zpool.ca:6144"
}

func (p *zpool) GetUsername() string {
	return p.Address
}

func (p *zpool) GetPassword() string {
	return "c=VTC,zap=VTC"
}

func (p *zpool) GetID() int {
	return 6
}

func (p *zpool) GetName() string {
	return "zpool.ca"
}

func (p *zpool) GetFee() float64 {
	return 0.50
}

func (p *zpool) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://zpool.ca/wallet/%s", addr))
}
