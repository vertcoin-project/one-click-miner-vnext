package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &BBQDroid{}

type BBQDroid struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewBBQDroid(addr string) *BBQDroid {
	return &BBQDroid{Address: addr}
}

func (p *BBQDroid) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://miningapi.bbqdroid.org/api/pools/vertcoin/miners/%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	vtc, ok := jsonPayload["pendingBalance"].(float64)
	if !ok {
		return 0
	}
	vtc *= 100000000
	return uint64(vtc)
}

func (p *BBQDroid) GetStratumUrl() string {
	return "stratum+tcp://bbqdroid.org:10001"
}

func (p *BBQDroid) GetUsername() string {
	return p.Address
}

func (p *BBQDroid) GetPassword() string {
	return "x"
}

func (p *BBQDroid) GetID() int {
	return 7
}

func (p *BBQDroid) GetName() string {
	return "BBQDroid.org"
}

func (p *BBQDroid) GetFee() float64 {
	return 0.5
}

func (p *BBQDroid) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://bbqdroid.org/?#vertcoin/dashboard?address=%s", addr))
}
