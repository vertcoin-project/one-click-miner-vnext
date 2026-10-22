package ping

import "testing"

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
