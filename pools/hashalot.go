package pools

import (
	"fmt"
	"time"

	"github.com/vertiond/verthash-one-click-miner/util"
)

var _ Pool = &Hashalot{}

type Hashalot struct {
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewHashalot() *Hashalot {
	return &Hashalot{}
}

func (p *Hashalot) GetPendingPayout(addr string) uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("http://api.hashalot.net/pools/vtc/miners/%s", addr), &jsonPayload)
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

func (p *Hashalot) GetStratumUrl() string {
	return "stratum+tcp://vertcoin.hashalot.net:3950"
}

func (p *Hashalot) GetPassword() string {
	return "x"
}

func (p *Hashalot) GetID() int {
	return 3
}

func (p *Hashalot) GetName() string {
	return "Hashalot.net"
}

func (p *Hashalot) GetFee() float64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson("http://api.hashalot.net/pools", &jsonPayload)
	if err != nil {
		return 2.0
	}
	
	pools, ok := jsonPayload["pools"].([]interface{})
	if !ok {
		return 2.0
	}
	
	pool, ok := pools[0].(map[string]interface{})
	if !ok {
		return 2.0
	}
	
	fee, ok := pool["poolFeePercent"].(float64)
	if !ok {
		return 2.0
	}

	return fee
}

func (p *Hashalot) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://hashalot.net/vtc/miners/%s", addr))
}
