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

func TestLoadBackendServersUsesFetchedList(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "sub", "backends.json")
	listURL := serveBackendList(t, `{"backendServers": ["https://a.example/", "https://b.example/"]}`)

	got := loadBackendServers(listURL, cachePath, util.GetJson)
	want := []string{"https://a.example/", "https://b.example/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("the fetched list was not cached: %v", err)
	}
	cached := backendList{}
	if err := json.Unmarshal(data, &cached); err != nil || !reflect.DeepEqual(cached.BackendServers, want) {
		t.Fatalf("cached list = %v (%v), want %v", cached.BackendServers, err, want)
	}
	if _, err := os.Stat(cachePath + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("the temporary cache file was left behind")
	}
}

func TestLoadBackendServersUsesCacheWhenOffline(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "backends.json")
	if err := writeBackendCache(cachePath, backendList{[]string{"https://cached.example/"}}); err != nil {
		t.Fatal(err)
	}

	got := loadBackendServers(offlineURL(), cachePath, util.GetJson)
	if want := []string{"https://cached.example/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestLoadBackendServersUsesCacheWhenFetchedListIsEmpty(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "backends.json")
	if err := writeBackendCache(cachePath, backendList{[]string{"https://cached.example/"}}); err != nil {
		t.Fatal(err)
	}
	listURL := serveBackendList(t, `{"backendServers": []}`)

	got := loadBackendServers(listURL, cachePath, util.GetJson)
	if want := []string{"https://cached.example/"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestLoadBackendServersFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	for name, listURL := range map[string]string{
		"offline":   offlineURL(),
		"not json":  serveBackendList(t, "404: Not Found"),
		"all bogus": serveBackendList(t, `{"backendServers": ["http://insecure.example/"]}`),
	} {
		cachePath := filepath.Join(dir, "missing.json")
		got := loadBackendServers(listURL, cachePath, util.GetJson)
		if !reflect.DeepEqual(got, networks.DefaultBackendServers) {
			t.Errorf("%s: got %v, want %v", name, got, networks.DefaultBackendServers)
		}
	}

	// A corrupt cache falls back to the defaults as well.
	cachePath := filepath.Join(dir, "corrupt.json")
	if err := os.WriteFile(cachePath, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := loadBackendServers(offlineURL(), cachePath, util.GetJson); !reflect.DeepEqual(got, networks.DefaultBackendServers) {
		t.Errorf("corrupt cache: got %v, want %v", got, networks.DefaultBackendServers)
	}
}

func TestLoadBackendServersDropsInvalidEntries(t *testing.T) {
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

	got := loadBackendServers(listURL, cachePath, util.GetJson)
	want := []string{"https://good.example/", "https://good.example:8443/api/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExampleBackendListIsValid(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "backends.json"))
	if err != nil {
		t.Fatal(err)
	}
	list := backendList{}
	if err := json.Unmarshal(data, &list); err != nil {
		t.Fatal(err)
	}
	if len(list.BackendServers) == 0 || len(validBackendServers(list.BackendServers)) != len(list.BackendServers) {
		t.Fatalf("backends.json has invalid or no servers: %v", list.BackendServers)
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

func TestSetNetworkKeepsResolvedBackendList(t *testing.T) {
	saved := networks.Active
	t.Cleanup(func() {
		networks.SetBackendServers(networks.DefaultBackendServers)
		networks.Active = saved
	})

	resolved := []string{"https://resolved.example/"}
	networks.SetNetwork(false)
	networks.SetBackendServers(resolved)
	if !reflect.DeepEqual(networks.Active.BackendServers, resolved) {
		t.Fatalf("mainnet list = %v, want %v", networks.Active.BackendServers, resolved)
	}

	networks.SetNetwork(true)
	if len(networks.Active.BackendServers) != 0 {
		t.Fatalf("testnet list = %v, want empty", networks.Active.BackendServers)
	}
	networks.SetNetwork(false)
	if !reflect.DeepEqual(networks.Active.BackendServers, resolved) {
		t.Fatalf("mainnet list after switching back = %v, want %v", networks.Active.BackendServers, resolved)
	}
}
