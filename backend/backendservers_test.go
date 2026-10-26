package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

func serveBackendList(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/backends.json"
}

// offlineURL returns the URL of a server that has already shut down.
func offlineURL() string {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	return srv.URL + "/backends.json"
}

func defaultManifest() backendManifest {
	return backendManifest{networks.DefaultBackendServers, networks.DefaultExplorerURL}
}

func TestLoadBackendManifestUsesFetchedList(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "sub", "backends.json")
	listURL := serveBackendList(t, `{"backendServers": ["https://a.example/", "https://b.example/"],
		"explorerURL": "https://blockbook.example/"}`)

	got := loadBackendManifest(listURL, cachePath, util.GetJson)
	want := backendManifest{[]string{"https://a.example/", "https://b.example/"}, "https://blockbook.example/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("the fetched list was not cached: %v", err)
	}
	cached := backendManifest{}
	if err := json.Unmarshal(data, &cached); err != nil || !reflect.DeepEqual(cached, want) {
		t.Fatalf("cached manifest = %v (%v), want %v", cached, err, want)
	}
	if _, err := os.Stat(cachePath + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("the temporary cache file was left behind")
	}
}

func TestLoadBackendManifestDefaultsMissingExplorer(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "backends.json")
	for name, explorer := range map[string]string{
		"missing":  "",
		"insecure": `, "explorerURL": "http://blockbook.example/"`,
		"no slash": `, "explorerURL": "https://blockbook.example"`,
	} {
		listURL := serveBackendList(t, `{"backendServers": ["https://a.example/"]`+explorer+`}`)
		got := loadBackendManifest(listURL, cachePath, util.GetJson)
		want := backendManifest{[]string{"https://a.example/"}, networks.DefaultExplorerURL}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
}

func TestLoadBackendManifestUsesCacheWhenOffline(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "backends.json")
	want := backendManifest{[]string{"https://cached.example/"}, "https://cached-explorer.example/"}
	if err := writeBackendCache(cachePath, want); err != nil {
		t.Fatal(err)
	}

	if got := loadBackendManifest(offlineURL(), cachePath, util.GetJson); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestLoadBackendManifestUsesCacheWhenFetchedListIsEmpty(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "backends.json")
	want := backendManifest{[]string{"https://cached.example/"}, networks.DefaultExplorerURL}
	if err := writeBackendCache(cachePath, want); err != nil {
		t.Fatal(err)
	}
	listURL := serveBackendList(t, `{"backendServers": [], "explorerURL": "https://blockbook.example/"}`)

	if got := loadBackendManifest(listURL, cachePath, util.GetJson); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestLoadBackendManifestFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	for name, listURL := range map[string]string{
		"offline":   offlineURL(),
		"not json":  serveBackendList(t, "404: Not Found"),
		"all bogus": serveBackendList(t, `{"backendServers": ["http://insecure.example/"]}`),
	} {
		cachePath := filepath.Join(dir, "missing.json")
		if got := loadBackendManifest(listURL, cachePath, util.GetJson); !reflect.DeepEqual(got, defaultManifest()) {
			t.Errorf("%s: got %v, want %v", name, got, defaultManifest())
		}
	}

	// A corrupt cache falls back to the defaults as well.
	cachePath := filepath.Join(dir, "corrupt.json")
	if err := os.WriteFile(cachePath, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := loadBackendManifest(offlineURL(), cachePath, util.GetJson); !reflect.DeepEqual(got, defaultManifest()) {
		t.Errorf("corrupt cache: got %v, want %v", got, defaultManifest())
	}
}

func TestLoadBackendManifestDropsInvalidEntries(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "backends.json")
	listURL := serveBackendList(t, `{"backendServers": [
		"https://good.example/",
		"http://insecure.example/",
		"https://no-slash.example",
		"https:///",
		"https://user:pass@creds.example/",
		"https://query.example/?x=/",
		"https://frag.example/#/",
		"ftp://other.example/",
		"",
		"https://good.example:8443/api/"
	]}`)

	got := loadBackendManifest(listURL, cachePath, util.GetJson).BackendServers
	want := []string{"https://good.example/", "https://good.example:8443/api/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExampleBackendManifestIsValid(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "backends.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := backendManifest{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	sanitized, ok := sanitizeBackendManifest(m)
	if !ok || !reflect.DeepEqual(sanitized, m) {
		t.Fatalf("backends.json has invalid entries: %v", m)
	}
}

func TestBackendServerSelectorWithEmptyList(t *testing.T) {
	saved := networks.Active
	t.Cleanup(func() { networks.Active = saved })

	networks.Active.BackendServers = nil
	networks.Active.OCMBackend = "https://stale.example/"
	(&Backend{}).BackendServerSelector()
	if networks.Active.OCMBackend != "" {
		t.Fatalf("OCMBackend = %q, want empty", networks.Active.OCMBackend)
	}
}

func TestSetNetworkKeepsResolvedEndpoints(t *testing.T) {
	saved := networks.Active
	t.Cleanup(func() {
		networks.SetMainnetEndpoints(networks.DefaultBackendServers, networks.DefaultExplorerURL)
		networks.Active = saved
	})

	resolved := []string{"https://resolved.example/"}
	explorer := "https://blockbook.example/"
	networks.SetNetwork()
	networks.SetMainnetEndpoints(resolved, explorer)
	if !reflect.DeepEqual(networks.Active.BackendServers, resolved) || networks.Active.ExplorerURL != explorer {
		t.Fatalf("mainnet = %v %q, want %v %q", networks.Active.BackendServers, networks.Active.ExplorerURL, resolved, explorer)
	}

	networks.SetNetwork()
	if !reflect.DeepEqual(networks.Active.BackendServers, resolved) || networks.Active.ExplorerURL != explorer {
		t.Fatalf("after SetNetwork = %v %q, want %v %q", networks.Active.BackendServers, networks.Active.ExplorerURL, resolved, explorer)
	}
}
