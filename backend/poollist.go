package backend

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/pools"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

const poolListURL = "https://raw.githubusercontent.com/vertcoin-project/one-click-miner-vnext/master/pools.json"

// The pool list is fetched before the window opens, so give up quickly and
// fall back to the cached copy.
var poolListClient = &http.Client{Timeout: 5 * time.Second}

func fetchPoolList(listURL string, target interface{}) error {
	r, err := poolListClient.Get(listURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

// poolManifest is the format of pools.json. The order of Pools is the order
// shown in the settings.
type poolManifest struct {
	Pools []pools.Config `json:"pools"`
}

// ResolvePools loads the mainnet pool list from pools.json (or its cached
// copy) and makes it active. Run it before ResetPool.
func (m *Backend) ResolvePools() {
	cachePath := filepath.Join(util.DataDirectory(), "pools.json")
	pools.SetConfigs(loadPoolConfigs(poolListURL, cachePath, fetchPoolList))
}

// loadPoolConfigs fetches pools.json and caches it at cachePath. If the fetch
// fails or yields no valid pools, the cached copy is used, and failing that
// the built-in default list.
func loadPoolConfigs(listURL, cachePath string, fetch func(string, interface{}) error) []pools.Config {
	fetched := poolManifest{}
	err := fetch(listURL, &fetched)
	if err == nil {
		if configs := validPoolConfigs(fetched.Pools); len(configs) > 0 {
			if err := writePoolCache(cachePath, poolManifest{configs}); err != nil {
				logging.Warnf("Unable to cache the pool list: %v\n", err)
			}
			return configs
		}
		logging.Warnf("The fetched pool list has no valid pools\n")
	} else {
		logging.Warnf("Unable to fetch the pool list: %v\n", err)
	}

	if data, err := os.ReadFile(cachePath); err == nil {
		cached := poolManifest{}
		if err := json.Unmarshal(data, &cached); err == nil {
			if configs := validPoolConfigs(cached.Pools); len(configs) > 0 {
				logging.Infof("Using the cached pool list\n")
				return configs
			}
		}
	}

	logging.Infof("Using the built-in pool list\n")
	return append([]pools.Config(nil), pools.DefaultConfigs...)
}

// ID 1 was the testnet P2Proxy pool and ID 2 is P2Pool.
const firstConfiguredPoolID = 3

// validPoolConfigs drops invalid entries and repeated IDs. A P2Pool entry
// keeps only its type.
func validPoolConfigs(configs []pools.Config) []pools.Config {
	valid := make([]pools.Config, 0, len(configs))
	seenP2Pool := false
	seenIDs := map[int]bool{}
	for _, c := range configs {
		if c.Type == pools.P2PoolType {
			if seenP2Pool {
				logging.Warnf("Ignoring repeated P2Pool entry\n")
				continue
			}
			seenP2Pool = true
			valid = append(valid, pools.Config{Type: pools.P2PoolType})
			continue
		}
		if reason := invalidPoolConfig(c); reason != "" {
			logging.Warnf("Ignoring pool %d [%s]: %s\n", c.ID, c.Name, reason)
			continue
		}
		if seenIDs[c.ID] {
			logging.Warnf("Ignoring pool %d [%s]: repeated ID\n", c.ID, c.Name)
			continue
		}
		seenIDs[c.ID] = true
		valid = append(valid, c)
	}
	return valid
}

// invalidPoolConfig says what is wrong with a non-P2Pool entry, or returns ""
// when it is usable.
func invalidPoolConfig(c pools.Config) string {
	switch {
	case c.Type != "":
		return "unknown type " + c.Type
	case c.ID < firstConfiguredPoolID:
		return "ID must be 3 or higher"
	case strings.TrimSpace(c.Name) == "":
		return "no name"
	case !validStratumURL(c.StratumURL):
		return "invalid stratum URL"
	case c.Password == "":
		return "no password"
	case c.Fee == nil || *c.Fee < 0 || *c.Fee >= 100:
		return "missing or invalid fee"
	case c.PayoutURL != "" && !validPoolWebURL(c.PayoutURL):
		return "invalid payout URL"
	case c.PayoutURL != "" && len(c.PayoutFields) == 0:
		return "payout URL without payout fields"
	case c.DashboardURL != "" && !validPoolWebURL(c.DashboardURL):
		return "invalid dashboard URL"
	}
	for _, f := range c.PayoutFields {
		if f == "" || strings.HasPrefix(f, ".") || strings.HasSuffix(f, ".") || strings.Contains(f, "..") {
			return "invalid payout field " + f
		}
	}
	return ""
}

// validStratumURL accepts stratum+tcp://host:port with nothing after the port.
func validStratumURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme == "stratum+tcp" && u.Hostname() != "" && u.Port() != "" &&
		u.User == nil && (u.Path == "" || u.Path == "/") && u.RawQuery == "" && u.Fragment == ""
}

// validPoolWebURL accepts https:// URLs with a host and no credentials, once
// {address} is filled in.
func validPoolWebURL(s string) bool {
	u, err := url.Parse(strings.ReplaceAll(s, "{address}", "address"))
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil
}

func writePoolCache(cachePath string, m poolManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return err
	}
	tmp := cachePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, cachePath); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
