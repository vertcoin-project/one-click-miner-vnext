package pools

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConfiguredPoolPendingPayout(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		w.Write([]byte(`{"getuserbalance": {"data": {"confirmed": 1.25, "unconfirmed": 0.5}}, "unpaid": "1.0"}`))
	}))
	defer srv.Close()

	p := NewConfiguredPool("vtc1qaddr", Config{
		ID:           4,
		PayoutURL:    srv.URL + "/api?key={address}",
		PayoutFields: []string{"getuserbalance.data.confirmed", "getuserbalance.data.unconfirmed"},
	})
	if got := p.GetPendingPayout(); got != 175000000 {
		t.Fatalf("payout = %d, want 175000000", got)
	}
	if gotPath != "/api?key=vtc1qaddr" {
		t.Fatalf("requested %q", gotPath)
	}

	for _, fields := range [][]string{
		{"getuserbalance.data.missing"},
		{"getuserbalance.data.confirmed.deeper"},
		{"unpaid"}, // a string, not a number
	} {
		p.Config.PayoutFields = fields
		if got := p.GetPendingPayout(); got != 0 {
			t.Errorf("fields %v: payout = %d, want 0", fields, got)
		}
	}

	p.Config.PayoutURL = ""
	if got := p.GetPendingPayout(); got != 0 {
		t.Errorf("no payout URL: payout = %d, want 0", got)
	}
}

func TestGetPoolsFollowsConfigs(t *testing.T) {
	t.Cleanup(func() { SetConfigs(DefaultConfigs) })

	SetConfigs([]Config{
		{ID: 7, Name: "Example", StratumURL: "stratum+tcp://pool.example:3333", Password: "x", Fee: fee(1.5)},
		{Type: P2PoolType},
	})
	ps := GetPools("vtc1qaddr")
	if len(ps) != 2 || ps[0].GetID() != 7 || ps[1].GetID() != 2 {
		t.Fatalf("got %d pools", len(ps))
	}
	if _, ok := ps[1].(*P2Pool); !ok {
		t.Fatalf("the p2pool entry is a %T", ps[1])
	}
	if ps[0].GetFee() != 1.5 || ps[0].GetUsername() != "vtc1qaddr" || ps[0].GetStratumUrl() != "stratum+tcp://pool.example:3333" {
		t.Fatalf("pool 7 = %+v", ps[0])
	}
	if !HasPool(7, "") || HasPool(4, "") {
		t.Fatalf("HasPool does not follow the configs")
	}
	if GetPool(4, "").GetID() != 7 {
		t.Fatalf("an unknown pool does not fall back to the first one")
	}
}
