package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &Suprnova{}

type Suprnova struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewSuprnova(addr string) *Suprnova {
	return &Suprnova{Address: addr}
}

func (p *Suprnova) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://vtc.suprnova.cc/index.php?page=api&action=getuserbalance&api_key=%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	el, ok := jsonPayload["getuserbalance"].(map[string]interface{})
	if !ok {
		return 0
	}
	el, ok = el["data"].(map[string]interface{})
	if !ok {
		return 0
	}

	confirmed, ok := el["confirmed"].(float64)
	if !ok {
		return 0
	}

	unconfirmed, ok := el["unconfirmed"].(float64)
	if !ok {
		return 0
	}

	vtc := confirmed + unconfirmed
	vtc *= 100000000
	return uint64(vtc)
}

func (p *Suprnova) GetStratumUrl() string {
	return "stratum+tcp://vtc.suprnova.cc:1776"
}

func (p *Suprnova) GetUsername() string {
	return p.Address
}

func (p *Suprnova) GetPassword() string {
	return "x"
}

func (p *Suprnova) GetID() int {
	return 4
}

func (p *Suprnova) GetName() string {
	return "Suprnova"
}

func (p *Suprnova) GetFee() float64 {
	return 1.00
}

func (p *Suprnova) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://vtc.suprnova.cc/index.php?page=anondashboard&user=%s", addr))
}
