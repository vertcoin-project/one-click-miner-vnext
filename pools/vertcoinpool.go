package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &Vertcoinpool{}

type Vertcoinpool struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewVertcoinpool(addr string) *Vertcoinpool {
	return &Vertcoinpool{Address: addr}
}

func (p *Vertcoinpool) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("http://vertcoinpool.com:4000/api/pools/verthash1/miners/%s", p.Address), &jsonPayload)
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

func (p *Vertcoinpool) GetStratumUrl() string {
	return "stratum+tcp://vtc.vertcoinpool.com:3052"
}

func (p *Vertcoinpool) GetUsername() string {
	return p.Address
}

func (p *Vertcoinpool) GetPassword() string {
	return "x"
}

func (p *Vertcoinpool) GetID() int {
	return 6
}

func (p *Vertcoinpool) GetName() string {
	return "Vertcoinpool.com"
}

func (p *Vertcoinpool) GetFee() float64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson("http://vertcoinpool.com:4000/api/pools", &jsonPayload)
	if err != nil {
		return 0.42
	}
	
	pools, ok := jsonPayload["pools"].([]interface{})
	if !ok {
		return 0.42
	}
	
	pool, ok := pools[0].(map[string]interface{})
	if !ok {
		return 0.42
	}
	
	fee, ok := pool["poolFeePercent"].(float64)
	if !ok {
		return 0.42
	}

	return fee
}

func (p *Vertcoinpool) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("http://vertcoinpool.com/?#verthash1/dashboard?address=%s", addr))
}
