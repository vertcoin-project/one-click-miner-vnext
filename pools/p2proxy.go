package pools

import (
	"fmt"

	"github.com/vertcoin-project/one-click-miner-vnext/networks"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &P2Proxy{}

type P2Proxy struct {
	Address string
}

func NewP2Proxy(addr string) *P2Proxy {
	return &P2Proxy{Address: addr}
}

func (p *P2Proxy) GetPendingPayout() uint64 {
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("%sapi/balance?address=%s", networks.Active.P2ProxyURL, p.Address), &jsonPayload)
	if err != nil {
		return 0
	}
	vtc, ok := jsonPayload[p.Address].(float64)
	if !ok {
		return 0
	}
	vtc *= 100000000
	return uint64(vtc)
}

func (p *P2Proxy) GetStratumUrl() string {
	return networks.Active.P2ProxyStratum
}

func (p *P2Proxy) GetUsername() string {
	return p.Address
}

func (p *P2Proxy) GetPassword() string {
	return "x"
}
