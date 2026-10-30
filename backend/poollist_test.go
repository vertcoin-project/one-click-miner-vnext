package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vertcoin-project/one-click-miner-vnext/pools"
)

func servePoolList(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/pools.json"
}

func offlinePoolListURL() string {
	srv := httptest.NewServer(http.NotFoundHandler())
	srv.Close()
	return srv.URL + "/pools.json"
}

func poolIDs(configs []pools.Config) []int {
	ids := []int{}
	for _, c := range configs {
		if c.Type == pools.P2PoolType {
			ids = append(ids, 2)
		} else {
			ids = append(ids, c.ID)
		}
	}
	return ids
}

const twoPools = `{"pools": [
	{"type": "p2pool"},
	{"id": 7, "name": "Example", "stratumURL": "stratum+tcp://pool.example:3333", "password": "x", "fee": 1.5,
	 "payoutURL": "https://pool.example/api/{address}", "payoutFields": ["balance"],
	 "dashboardURL": "https://pool.example/miner/{address}"}
]}`

func TestLoadPoolConfigsUsesFetchedList(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "sub", "pools.json")

	got := loadPoolConfigs(servePoolList(t, twoPools), cachePath, fetchPoolList)
	if ids := poolIDs(got); !reflect.DeepEqual(ids, []int{2, 7}) {
		t.Fatalf("got pools %v, want [2 7]", ids)
	}
	if got[1].Name != "Example" || *got[1].Fee != 1.5 || got[1].StratumURL != "stratum+tcp://pool.example:3333" {
		t.Fatalf("pool 7 = %+v", got[1])
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("the fetched list was not cached: %v", err)
	}
	cached := poolManifest{}
	if err := json.Unmarshal(data, &cached); err != nil || !reflect.DeepEqual(cached.Pools, got) {
		t.Fatalf("cached list = %+v (%v), want %+v", cached.Pools, err, got)
	}
	if _, err := os.Stat(cachePath + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("the temporary cache file was left behind")
	}
}

func TestLoadPoolConfigsUsesCacheWhenOffline(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "pools.json")
	loadPoolConfigs(servePoolList(t, twoPools), cachePath, fetchPoolList)

	got := loadPoolConfigs(offlinePoolListURL(), cachePath, fetchPoolList)
	if ids := poolIDs(got); !reflect.DeepEqual(ids, []int{2, 7}) {
		t.Fatalf("got pools %v, want [2 7]", ids)
	}
}

func TestLoadPoolConfigsUsesCacheWhenFetchedListIsEmpty(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "pools.json")
	loadPoolConfigs(servePoolList(t, twoPools), cachePath, fetchPoolList)

	got := loadPoolConfigs(servePoolList(t, `{"pools": []}`), cachePath, fetchPoolList)
	if ids := poolIDs(got); !reflect.DeepEqual(ids, []int{2, 7}) {
		t.Fatalf("got pools %v, want [2 7]", ids)
	}
}

func TestLoadPoolConfigsFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	want := poolIDs(pools.DefaultConfigs)
	for name, listURL := range map[string]string{
		"offline":   offlinePoolListURL(),
		"not json":  servePoolList(t, "404: Not Found"),
		"all bogus": servePoolList(t, `{"pools": [{"id": 7, "name": "No stratum", "password": "x", "fee": 1}]}`),
	} {
		got := loadPoolConfigs(listURL, filepath.Join(dir, "missing.json"), fetchPoolList)
		if ids := poolIDs(got); !reflect.DeepEqual(ids, want) {
			t.Errorf("%s: got pools %v, want %v", name, ids, want)
		}
	}

	cachePath := filepath.Join(dir, "corrupt.json")
	if err := os.WriteFile(cachePath, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}
	if ids := poolIDs(loadPoolConfigs(offlinePoolListURL(), cachePath, fetchPoolList)); !reflect.DeepEqual(ids, want) {
		t.Errorf("corrupt cache: got pools %v, want %v", ids, want)
	}
}

func TestLoadPoolConfigsDropsInvalidEntries(t *testing.T) {
	listURL := servePoolList(t, `{"pools": [
		{"id": 10, "name": "Good", "stratumURL": "stratum+tcp://good.example:3333", "password": "x", "fee": 0},
		{"type": "p2pool", "name": "ignored", "stratumURL": "stratum+tcp://ignored.example:1"},
		{"type": "p2pool"},
		{"type": "solo", "id": 11, "name": "Unknown type", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 1},
		{"id": 2, "name": "Takes P2Pool's ID", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 1},
		{"id": 10, "name": "Repeated ID", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 1},
		{"id": 12, "name": " ", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 1},
		{"id": 13, "name": "No port", "stratumURL": "stratum+tcp://a.example", "password": "x", "fee": 1},
		{"id": 14, "name": "Wrong scheme", "stratumURL": "https://a.example:1", "password": "x", "fee": 1},
		{"id": 15, "name": "Credentials", "stratumURL": "stratum+tcp://u:p@a.example:1", "password": "x", "fee": 1},
		{"id": 16, "name": "No password", "stratumURL": "stratum+tcp://a.example:1", "fee": 1},
		{"id": 17, "name": "No fee", "stratumURL": "stratum+tcp://a.example:1", "password": "x"},
		{"id": 18, "name": "Negative fee", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": -1},
		{"id": 19, "name": "Huge fee", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 100},
		{"id": 20, "name": "Insecure payout", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 1,
		 "payoutURL": "http://a.example/{address}", "payoutFields": ["x"]},
		{"id": 21, "name": "Payout without fields", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 1,
		 "payoutURL": "https://a.example/{address}"},
		{"id": 22, "name": "Bad field", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 1,
		 "payoutURL": "https://a.example/{address}", "payoutFields": ["a..b"]},
		{"id": 23, "name": "Insecure dashboard", "stratumURL": "stratum+tcp://a.example:1", "password": "x", "fee": 1,
		 "dashboardURL": "javascript:alert(1)"},
		{"id": 24, "name": "Also good", "stratumURL": "stratum+tcp://also.example:3052/", "password": "c=VTC", "fee": 0.6,
		 "payoutURL": "https://also.example/api?a={address}", "payoutFields": ["a.b", "c"],
		 "dashboardURL": "https://also.example/#dash?address={address}"}
	]}`)

	got := loadPoolConfigs(listURL, filepath.Join(t.TempDir(), "pools.json"), fetchPoolList)
	if ids := poolIDs(got); !reflect.DeepEqual(ids, []int{10, 2, 24}) {
		t.Fatalf("got pools %v, want [10 2 24]", ids)
	}
	if !reflect.DeepEqual(got[1], pools.Config{Type: pools.P2PoolType}) {
		t.Fatalf("P2Pool entry kept extra fields: %+v", got[1])
	}
}

func TestExamplePoolListIsValid(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "pools.json"))
	if err != nil {
		t.Fatal(err)
	}
	m := poolManifest{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Pools) == 0 || len(validPoolConfigs(m.Pools)) != len(m.Pools) {
		t.Fatalf("pools.json has invalid or no pools: %+v", m.Pools)
	}
	if !reflect.DeepEqual(m.Pools, pools.DefaultConfigs) {
		t.Fatalf("pools.json and the built-in pool list differ")
	}
}
