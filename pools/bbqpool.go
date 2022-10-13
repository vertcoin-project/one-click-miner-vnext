package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &BBQPool{}

type BBQPool struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewBBQPool(addr string) *BBQPool {
	return &BBQPool{Address: addr}
}

func (p *BBQPool) GetJsonMember(val interface{}, keys []string) (interface{}, bool) {
	for _, k := range keys {
		valMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, false
		}

		val, ok = valMap[k]
		if !ok {
			return nil, false
		}
	}
	return val, true
}

func (p *BBQPool) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://bbqpool.org/api/vertcoin/miners?method=%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}

	generateRaw, ok := p.GetJsonMember(jsonPayload, []string{"body","primary","payments","generate"})
	if !ok {
		return 0
	}

	immatureRaw, ok := p.GetJsonMember(jsonPayload, []string{"body","primary","payments","immature"})
    if !ok {
		return 0
	}

	generate, ok := generateRaw.(float64)
    if !ok {
		return 0
	}

	immature, ok := immatureRaw.(float64)
    if !ok {
		return 0
	}
	
	vtc := (generate + immature) * 100000000
	return uint64(vtc)
}

func (p *BBQPool) GetStratumUrl() string {
	return "stratum+tcp://bbqpool.org:10001"
}

func (p *BBQPool) GetUsername() string {
	return p.Address
}

func (p *BBQPool) GetPassword() string {
	return "x"
}

func (p *BBQPool) GetID() int {
	return 7
}

func (p *BBQPool) GetName() string {
	return "bbqpool.org"
}

func (p *BBQPool) GetFee() float64 {
	return 1.0
}

func (p *BBQPool) OpenBrowserPayoutInfo(_addr string) {
	util.OpenBrowser("https://bbqpool.org/")
}
