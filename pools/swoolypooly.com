package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &SWoolyPooly{}

type SWoolyPooly struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewSWoolyPooly(addr string) *SWoolyPooly {
	return &SWoolyPooly{Address: addr}
}

func (p *SWoolyPooly) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://api.woolypooly.com/api/vtc-1/accounts/%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	vtc, ok := jsonPayload["stats"]["immature_balance"].(float64)
	if !ok {
		return 0
	}
	vtc *= 100000000
	return uint64(vtc)
}

func (p *SWoolyPooly) GetStratumUrl() string {
	return "stratum+tcp://pool.woolypooly.com:3103"
}

func (p *SWoolyPooly) GetUsername() string {
	return p.Address
}

func (p *SWoolyPooly) GetPassword() string {
	return "x"
}

func (p *SWoolyPooly) GetID() int {
	return 10
}

func (p *SWoolyPooly) GetName() string {
	return "[SOLO] WoolyPooly"
}

func (p *SWoolyPooly) GetFee() float64 {
	return 0.9
}

func (p *SWoolyPooly) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://woolypooly.com/ru/coin/vtc/wallet/%s", addr))
}
