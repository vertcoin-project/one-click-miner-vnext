package backend

import (
	"errors"
	"testing"
)

func TestSweepResult(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{nil, "ok"},
		{errors.New("insufficient_funds"), "insufficient_funds"},
		{errors.New("invalid_address"), "invalid_address"},
		{errors.New(`Get "https://insight.vertcoin.org/api": dial tcp: timeout`), "other"},
	}
	for _, c := range cases {
		if got := sweepResult(c.err); got != c.want {
			t.Errorf("sweepResult(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}
