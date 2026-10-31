package ping

import (
	"fmt"
	"testing"
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
