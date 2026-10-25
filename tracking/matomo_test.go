package tracking

import (
	"path/filepath"
	"testing"
)

func TestRedactHome(t *testing.T) {
	cases := []struct {
		text, home string
		windows    bool
		want       string
	}{
		{`open C:\Users\Jane Doe\AppData\Roaming\vertcoin-ocm\miners: access denied`, `C:\Users\Jane Doe`, true,
			`open ~\AppData\Roaming\vertcoin-ocm\miners: access denied`},
		{`c:/users/jane doe/x and C:\USERS\JANE DOE`, `C:\Users\Jane Doe`, true, `~/x and ~`},
		{`C:\Users\Janet\x`, `C:\Users\Jan`, true, `C:\Users\Janet\x`},
		{`stat /home/jane/.vertcoin-ocm: no such file`, `/home/jane`, false, `stat ~/.vertcoin-ocm: no such file`},
		{`/home/Jane/x`, `/home/jane`, false, `/home/Jane/x`},
		{`no paths here`, `/`, false, `no paths here`},
	}
	for _, c := range cases {
		if got := redactHome(c.text, filepath.FromSlash(c.home), c.windows); got != c.want {
			t.Errorf("redactHome(%q, %q) = %q, want %q", c.text, c.home, got, c.want)
		}
	}
}
