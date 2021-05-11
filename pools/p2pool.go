package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &P2Pool{}

type P2Pool struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
	LastFetchedFee    time.Time
	LastFee           float64
}

func NewP2Pool(addr string) *P2Pool {
	return &P2Pool{Address: addr}
}

func (p *P2Pool) GetPendingPayout() uint64 {
	if time.Now().Sub(p.LastFetchedPayout) > time.Minute*2 {
		jsonPayload := map[string]interface{}{}
		err := util.GetJson(fmt.Sprintf("%scurrent_payouts", networks.Active.P2ProxyURL), &jsonPayload)
		if err != nil {
			logging.Warnf("Unable to fetch p2pool payouts: %s", err.Error())
			p.LastPayout = 0
		}
		address := p.Address
		vtc, ok := jsonPayload[address].(float64)
		if !ok {
			p.LastFetchedPayout = time.Now()
			p.LastPayout = 0
		}
		vtc *= 100000000
		p.LastFetchedPayout = time.Now()
		p.LastPayout = uint64(vtc)
	}
	return p.LastPayout
}

func (p *P2Pool) GetStratumUrl() string {
	return networks.Active.P2ProxyStratum
}

func (p *P2Pool) GetUsername() string {
	return p.Address
}

func (p *P2Pool) GetPassword() string {
	return "x"
}

func (p *P2Pool) GetID() int {
	return 2
}

func (p *P2Pool) GetName() string {
	return "P2Pool"
}

func (p *P2Pool) GetFee() (fee float64) {
	if time.Now().Sub(p.LastFetchedFee) > time.Minute*30 {
		jsonPayload := map[string]interface{}{}
		err := util.GetJson(fmt.Sprintf("%slocal_stats", networks.Active.P2ProxyURL), &jsonPayload)
		if err != nil {
			logging.Warnf("Unable to fetch p2pool fee: %s", err.Error())
			fee = 2.0
		}
		fee, ok := jsonPayload["fee"].(float64)
		if !ok {
			fee = 2.0
			return fee
		}
		donationFee, ok := jsonPayload["donation_proportion"].(float64)
		if !ok {
			fee = 2.0
			return fee
		}
		fee += donationFee
		p.LastFetchedFee = time.Now()
		p.LastFee = float64(fee)
	}
	return p.LastFee
}

func (p *P2Pool) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(networks.Active.P2ProxyURL)
}
