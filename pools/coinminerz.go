package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &CoinMinerz{}

type CoinMinerz struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewCoinMinerz(addr string) *CoinMinerz {
	return &CoinMinerz{Address: addr}
}

func (p *CoinMinerz) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://coinminerz.com/api/v1/Vertcoin/statistics?address=%s", p.address), &jsonPayload)
	if err != nil {
		return 0
	}
	vtc, ok := jsonPayload["payments.next"].(float64)
	if !ok {
		return 0
	}
	vtc *= 100000000
	return uint64(vtc)
}

func (p *CoinMinerz) GetStratumUrl() string {
	return "stratum+tcp://stratum.coinminerz.com:3317"
}

func (p *CoinMinerz) GetUsername() string {
	return p.Address
}

func (p *CoinMinerz) GetPassword() string {
	return "x"
}

func (p *CoinMinerz) GetID() int {
	return 10
}

func (p *CoinMinerz) GetName() string {
	return "CoinMinerz.com"
}

func (p *CoinMinerz) GetFee() float64 {
	return 0.50
}

func (p *CoinMinerz) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://coinminerz.com/miner/Vertcoin/%s", addr))
}
