package backend

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

const backendListURL = "https://raw.githubusercontent.com/vertcoin-project/one-click-miner-vnext/master/backends.json"

type backendList struct {
	BackendServers []string `json:"backendServers"`
}

// ResolveBackendServers loads the mainnet backend list from backends.json
// (or its cached copy) and makes it the active list. Run it before
// BackendServerSelector.
func (m *Backend) ResolveBackendServers() {
	cachePath := filepath.Join(util.DataDirectory(), "backends.json")
	networks.SetBackendServers(loadBackendServers(backendListURL, cachePath, util.GetJson))
}

// loadBackendServers fetches the backend list and caches it at cachePath.
// If the fetch fails or yields no valid servers, the cached copy is used,
// and failing that the built-in default list.
func loadBackendServers(listURL, cachePath string, fetch func(string, interface{}) error) []string {
	fetched := backendList{}
	err := fetch(listURL, &fetched)
	if err == nil {
		if servers := validBackendServers(fetched.BackendServers); len(servers) > 0 {
			if err := writeBackendCache(cachePath, backendList{servers}); err != nil {
				logging.Warnf("Unable to cache the backend list: %v\n", err)
			}
			return servers
		}
		logging.Warnf("The fetched backend list has no valid servers\n")
	} else {
		logging.Warnf("Unable to fetch the backend list: %v\n", err)
	}

	if data, err := os.ReadFile(cachePath); err == nil {
		cached := backendList{}
		if err := json.Unmarshal(data, &cached); err == nil {
			if servers := validBackendServers(cached.BackendServers); len(servers) > 0 {
				logging.Infof("Using the cached backend list\n")
				return servers
			}
		}
	}

	logging.Infof("Using the built-in backend list\n")
	return append([]string(nil), networks.DefaultBackendServers...)
}

// validBackendServers keeps only https:// URLs with a host that end in "/"
// and carry no credentials, query or fragment.
func validBackendServers(servers []string) []string {
	valid := make([]string, 0, len(servers))
	for _, s := range servers {
		u, err := url.Parse(s)
		if err != nil || !strings.HasPrefix(s, "https://") || !strings.HasSuffix(s, "/") ||
			u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			logging.Warnf("Ignoring invalid backend server [%s]\n", s)
			continue
		}
		valid = append(valid, s)
	}
	return valid
}

func writeBackendCache(cachePath string, list backendList) error {
	data, err := json.MarshalIndent(list, "", "  ")
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
