package ping

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestPickResponsiveNode(t *testing.T) {
	nodes := []Nodes{
		{Hostname: "dead1", URL: "http://dead1/"},
		{Hostname: "alive", URL: "http://alive/"},
		{Hostname: "dead2", URL: "http://dead2/"},
	}
	up := func(url string) bool { return url == "http://alive/" }

	// The pick is random, so check it never lands on a dead node.
	for n := 0; n < 50; n++ {
		node, ok := pickResponsiveNode(nodes, up)
		if !ok || node.Hostname != "alive" {
			t.Fatalf("got %q, %v; want alive, true", node.Hostname, ok)
		}
	}

	if _, ok := pickResponsiveNode(nodes, func(string) bool { return false }); ok {
		t.Fatal("expected no node when none respond")
	}
}

// Legacy nodes report miner_hash_rates as an object of address to hashrate.
// The miner count is the number of entries.
func TestCheckCurrentMiners(t *testing.T) {
	rates := func(n int) map[string]interface{} {
		m := map[string]interface{}{}
		for i := 0; i < n; i++ {
			m[fmt.Sprintf("V%d", i)] = 1000.0
		}
		return map[string]interface{}{"miner_hash_rates": m}
	}
	if !CheckCurrentMiners(rates(Set.MaxMiners)) {
		t.Errorf("%d miners should pass", Set.MaxMiners)
	}
	if CheckCurrentMiners(rates(Set.MaxMiners + 1)) {
		t.Errorf("%d miners should fail", Set.MaxMiners+1)
	}
	if !CheckCurrentMiners(map[string]interface{}{}) {
		t.Error("a node that doesn't report miners should pass")
	}
}

// pnode builds a node the selection cascade can judge: no fee, the given
// number of miners and ping time.
func pnode(host string, ping time.Duration, minerCount int) Nodes {
	rates := map[string]interface{}{}
	for i := 0; i < minerCount; i++ {
		rates[fmt.Sprintf("M%d", i)] = 1000.0
	}
	return Nodes{
		Hostname: host, URL: "http://" + host + "/", Stratum: "stratum+tcp://" + host,
		PingTime: ping,
		info:     map[string]interface{}{"fee": 0.0, "donation_proportion": 0.0, "miner_hash_rates": rates},
	}
}

func picked(t *testing.T, nodes []Nodes, pingErr error) string {
	t.Helper()
	sel, ok := pick(nodes, pingErr)
	if !ok {
		return ""
	}
	return sel.P2PoolURL
}

// A busy node is only skipped for a usable node close by in ping.
func TestPickBusyNode(t *testing.T) {
	busy, free := Set.MaxMiners+5, 1
	ms := time.Millisecond
	dead := pnode("dead", 12*ms, free)
	dead.info = nil
	pricey := pnode("pricey", 12*ms, free)
	pricey.info["fee"] = Set.MaxFee + 1
	cases := []struct {
		name  string
		nodes []Nodes
		want  string
	}{
		{"skips to a close free node", []Nodes{pnode("n1", 10*ms, busy), pnode("n2", 20*ms, free)}, "http://n2/"},
		{"keeps it when the free node is far", []Nodes{pnode("n1", 10*ms, busy), pnode("n2", 10*ms+Set.MaxNextPing+ms, free)}, "http://n1/"},
		{"keeps it when the next node blocks ping", []Nodes{pnode("n1", 10*ms, busy), pnode("n2", 0, free)}, "http://n1/"},
		{"keeps it when the close node is down", []Nodes{pnode("n1", 10*ms, busy), dead, pnode("far", 10*ms+Set.MaxNextPing+ms, free)}, "http://n1/"},
		{"keeps it when the close node's fee is too high", []Nodes{pnode("n1", 10*ms, busy), pricey}, "http://n1/"},
		{"looks past an unusable node", []Nodes{pnode("n1", 10*ms, busy), dead, pnode("n2", 20*ms, free)}, "http://n2/"},
		{"keeps the nearest when all are busy", []Nodes{pnode("n1", 10*ms, busy), pnode("n2", 20*ms, busy), pnode("n3", 30*ms, busy)}, "http://n1/"},
	}
	for _, c := range cases {
		if got := picked(t, c.nodes, nil); got != c.want {
			t.Errorf("%s: picked %s, want %s", c.name, got, c.want)
		}
	}
}

func TestPickPingFailed(t *testing.T) {
	dead := pnode("dead", 0, 1)
	dead.info = nil
	if got := picked(t, []Nodes{dead, pnode("alive", 0, 1)}, errors.New("ping failed")); got != "http://alive/" {
		t.Fatalf("picked %s, want the responsive node", got)
	}
	if got := picked(t, []Nodes{dead}, errors.New("ping failed")); got != "" {
		t.Fatalf("picked %s, want nothing when no node responds", got)
	}
}
