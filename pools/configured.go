package pools

import (
	"net/url"
	"strings"

	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

var _ Pool = &ConfiguredPool{}

// ConfiguredPool is a pool described entirely by a pools.json entry.
type ConfiguredPool struct {
	Address string
	Config  Config
}

func NewConfiguredPool(addr string, c Config) *ConfiguredPool {
	return &ConfiguredPool{Address: addr, Config: c}
}

func (p *ConfiguredPool) withAddress(template string) string {
	return strings.ReplaceAll(template, "{address}", url.QueryEscape(p.Address))
}

func (p *ConfiguredPool) GetPendingPayout() uint64 {
	if p.Config.PayoutURL == "" || len(p.Config.PayoutFields) == 0 {
		return 0
	}
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(p.withAddress(p.Config.PayoutURL), &jsonPayload)
	if err != nil {
		return 0
	}
	vtc := 0.0
	for _, field := range p.Config.PayoutFields {
		v, ok := lookupNumber(jsonPayload, field)
		if !ok {
			return 0
		}
		vtc += v
	}
	vtc *= 100000000
	return uint64(vtc)
}

// lookupNumber follows a dot-separated path of object keys and returns the
// number found there.
func lookupNumber(payload map[string]interface{}, path string) (float64, bool) {
	var el interface{} = payload
	for _, key := range strings.Split(path, ".") {
		obj, ok := el.(map[string]interface{})
		if !ok {
			return 0, false
		}
		el = obj[key]
	}
	v, ok := el.(float64)
	return v, ok
}

func (p *ConfiguredPool) GetStratumUrl() string {
	return p.Config.StratumURL
}

func (p *ConfiguredPool) GetUsername() string {
	return p.Address
}

func (p *ConfiguredPool) GetPassword() string {
	return p.Config.Password
}

func (p *ConfiguredPool) GetID() int {
	return p.Config.ID
}

func (p *ConfiguredPool) GetName() string {
	return p.Config.Name
}

func (p *ConfiguredPool) GetFee() float64 {
	if p.Config.Fee == nil {
		return 0
	}
	return *p.Config.Fee
}

func (p *ConfiguredPool) OpenBrowserPayoutInfo(addr string) {
	if p.Config.DashboardURL == "" {
		return
	}
	util.OpenBrowser(strings.ReplaceAll(p.Config.DashboardURL, "{address}", url.QueryEscape(addr)))
}
