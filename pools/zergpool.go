package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &Zergpool{}

type Zergpool struct {
	Address           string
	Payout            string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewZergpool(addr string) *Zergpool {
	return &Zergpool{Address: addr}
}

func (p *Zergpool) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("http://api.zergpool.com:8080/api/wallet?address=%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	doge, ok := jsonPayload["unpaid"].(float64)
	if !ok {
		return 0
	}
	doge *= 100000000
	return uint64(doge)
}

func (p *Zergpool) GetStratumUrl() string {
	return "stratum+tcp://verthash.na.mine.zergpool.com:4534"
}

func (p *Zergpool) GetUsername() string {
	return p.Address
}

func (p *Zergpool) GetPassword() string {
	return "x"
}

func (p *Zergpool) GetID() int {
	return 5
}

func (p *Zergpool) GetName() string {
	return "Zergpool"
}

func (p *Zergpool) GetFee() float64 {
	return 0.50
}
