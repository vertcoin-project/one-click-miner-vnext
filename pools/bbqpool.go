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

func (p *BBQPool) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://bbqpool.org/api/vertcoin/miners?method=%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	generate, ok := jsonPayload["body"]["primary"]["payments"]["generate"].(float64)
	if !ok {
		return 0
	}
	immature, ok := jsonPayload["body"]["primary"]["payments"]["immature"].(float64)
	if !ok {
		return 0
	}
	vtc = (generate + immature) * 100000000
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

func (p *BBQPool) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://bbqpool.org/", addr))
}
