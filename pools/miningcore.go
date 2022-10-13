package pools

import (
	"fmt"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &MiningcorePool{}

type MiningcorePool struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewMiningcorePool(addr string) *MiningcorePool {
	return &MiningcorePool{Address: addr}
}

func (p *MiningcorePool) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("https://miningcore.pro/api/pools/vtc/account/%s", p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	result, ok := jsonPayload["result"].(map[string]interface{})
	if !ok {
		return 0
	}
	vtc, ok := result["pendingBalance"].(float64)
	if !ok {
		return 0
	}
	return uint64(vtc)
}

func (p *MiningcorePool) GetStratumUrl() string {
	return "stratum+tcp://pool.miningcore.pro:3096"
}

func (p *MiningcorePool) GetUsername() string {
	return p.Address
}

func (p *MiningcorePool) GetPassword() string {
	return "x"
}

func (p *MiningcorePool) GetID() int {
	return 11
}

func (p *MiningcorePool) GetName() string {
	return "Miningcore Pro"
}

func (p *MiningcorePool) GetFee() float64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson("https://miningcore.pro/api/pools/vtc", &jsonPayload)
	if err != nil {
		return 0
	}
	fee, ok := jsonPayload["fee"].(float64)
	if !ok {
		return 0
	}
	return fee
}

func (p *MiningcorePool) OpenBrowserPayoutInfo(addr string) {
	util.OpenBrowser(fmt.Sprintf("https://miningcore.pro/pool/vtc/account/%s", addr))
}
