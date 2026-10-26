package tracking

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
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

func TestTrackBeforeStartAndAfterStopIsDropped(t *testing.T) {
	Stop() // before StartTracker: must not panic
	Track(TrackingRequest{Category: "Lifecycle", Action: "Crash"})

	useTempDataDir(t)
	endpoint = "http://127.0.0.1:1/"
	StartTracker()
	Stop()
	Stop()
	Track(TrackingRequest{Category: "Lifecycle", Action: "Crash"}) // after Stop: must not panic
}

func TestStopSendsQueuedEvents(t *testing.T) {
	var mu sync.Mutex
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = append(got, r.URL.Query().Get("idsite")+"|"+r.URL.Query().Get("e_a")+"|"+r.URL.Query().Get("e_n"))
		mu.Unlock()
	}))
	defer srv.Close()
	useTempDataDir(t)
	old := endpoint
	endpoint = srv.URL
	defer func() { endpoint = old }()

	StartTracker()
	home, _ := os.UserHomeDir()
	Track(TrackingRequest{Category: "Lifecycle", Action: "Crash", Name: "open " + filepath.Join(home, "x")})
	Stop()

	want := siteID + "|Crash|open " + filepath.Join("~", "x")
	if len(got) != 1 || got[0] != want {
		t.Errorf("server got %q, want [%q]", got, want)
	}
}

// useTempDataDir points util.DataDirectory at a temp folder, so the tests
// don't touch the real tracking and unique_id files.
func useTempDataDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, "vertcoin-ocm"), 0700); err != nil {
		t.Fatal(err)
	}
	vId = ""
}
