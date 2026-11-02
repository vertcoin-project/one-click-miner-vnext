package backend

import (
	"testing"

	"github.com/vertcoin-project/one-click-miner-vnext/pools"
)

// Until a P2Pool node is picked its fee is unknown, so none is shown.
func TestGetPoolFeeBeforeP2PoolNode(t *testing.T) {
	m := &Backend{pool: pools.NewP2Pool("Vaddr")}
	if got := m.GetPoolFee(); got != "" {
		t.Errorf("GetPoolFee() before a node is picked = %q, want \"\"", got)
	}
}

func TestFormatHashrate(t *testing.T) {
	cases := map[float64]string{
		0:             "0.00 kH/s",
		500:           "0.50 kH/s",
		1000:          "1.00 kH/s",
		1000000:       "1.00 MH/s",
		2010000:       "2.01 MH/s",
		676736286:     "676.74 MH/s",
		1500000000:    "1.50 GH/s",
		2500000000000: "2.50 TH/s",
	}
	for in, want := range cases {
		if got := formatHashrate(in); got != want {
			t.Errorf("formatHashrate(%v) = %q, want %q", in, got, want)
		}
	}
}
