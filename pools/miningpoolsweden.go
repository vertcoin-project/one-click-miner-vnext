package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &MiningpoolSweden{}

type MiningpoolSweden struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewMiningpoolSweden(addr string) *MiningpoolSweden {
	return &MiningpoolSweden{Address: addr}
}

func (p *MiningpoolSweden) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://api.miningpoolsweden.eu/api/pools/vert1/miners/%s", p.Address), &jsonPayload)
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

func (p *MiningpoolSweden) GetStratumUrl() string {
	return "stratum+tcp://vtc.miningpoolsweden.eu:3052"
}

func (p *MiningpoolSweden) GetUsername() string {
	return p.Address
}

func (p *MiningpoolSweden) GetPassword() string {
	return "x"
}

func (p *MiningpoolSweden) GetID() int {
	return 9
}

func (p *MiningpoolSweden) GetName() string {
	return "MiningpoolSweden.eu"
}

func (p *MiningpoolSweden) GetFee() float64 {
	return 0.6
}

func (p *MiningpoolSweden) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://miningpoolsweden.eu/?#vert1/dashboard?address=%s", addr))
}
