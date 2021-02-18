package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &Hashalot{}

type Hashalot struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewHashalot(addr string) *Hashalot {
	return &Hashalot{Address: addr}
}

func (p *Hashalot) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("http://api.hashalot.net/pools/vtc/miners/%s", p.Address), &jsonPayload)
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

func (p *Hashalot) GetUsername() string {
	return p.Address
}

func (p *Hashalot) GetPassword() string {
	return "x"
}

func (p *Hashalot) GetID() int {
	return 3
}

func (p *Hashalot) GetName() string {
	return "Hashalot"
}

func (p *Hashalot) GetFee() float64 {
	return 2.00
}
