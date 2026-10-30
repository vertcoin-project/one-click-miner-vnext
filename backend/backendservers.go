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

// backendManifest is the format of backends.json. ExplorerURL is optional and
// may point at an Insight or a Blockbook server.
type backendManifest struct {
	BackendServers []string `json:"backendServers"`
	ExplorerURL    string   `json:"explorerURL,omitempty"`
}

// ResolveBackendServers loads the mainnet backend list and explorer URL from
// backends.json (or its cached copy) and makes them active. Run it before
// BackendServerSelector.
func (m *Backend) ResolveBackendServers() {
	cachePath := filepath.Join(util.DataDirectory(), "backends.json")
	manifest := loadBackendManifest(backendListURL, cachePath, util.GetJson)
	networks.SetMainnetEndpoints(manifest.BackendServers, manifest.ExplorerURL)
	logging.Infof("Using explorer: %s\n", manifest.ExplorerURL)
}

// loadBackendManifest fetches backends.json and caches it at cachePath. If the
// fetch fails or yields no valid servers, the cached copy is used, and failing
// that the built-in defaults. A missing or invalid explorer URL is replaced by
// the default one.
func loadBackendManifest(listURL, cachePath string, fetch func(string, interface{}) error) backendManifest {
	fetched := backendManifest{}
	err := fetch(listURL, &fetched)
	if err == nil {
		if m, ok := sanitizeBackendManifest(fetched); ok {
			if err := writeBackendCache(cachePath, m); err != nil {
				logging.Warnf("Unable to cache the backend list: %v\n", err)
			}
			return m
		}
		logging.Warnf("The fetched backend list has no valid servers\n")
	} else {
		logging.Warnf("Unable to fetch the backend list: %v\n", err)
	}

	if data, err := os.ReadFile(cachePath); err == nil {
		cached := backendManifest{}
		if err := json.Unmarshal(data, &cached); err == nil {
			if m, ok := sanitizeBackendManifest(cached); ok {
				logging.Infof("Using the cached backend list\n")
				return m
			}
		}
	}

	logging.Infof("Using the built-in backend list\n")
	return backendManifest{
		BackendServers: append([]string(nil), networks.DefaultBackendServers...),
		ExplorerURL:    networks.DefaultExplorerURL,
	}
}

// sanitizeBackendManifest drops invalid servers and replaces an invalid
// explorer URL with the default. It reports false when no server is left.
func sanitizeBackendManifest(m backendManifest) (backendManifest, bool) {
	servers := make([]string, 0, len(m.BackendServers))
	for _, s := range m.BackendServers {
		if !validEndpointURL(s) {
			logging.Warnf("Ignoring invalid backend server [%s]\n", s)
			continue
		}
		servers = append(servers, s)
	}
	explorer := m.ExplorerURL
	if !validEndpointURL(explorer) {
		if explorer != "" {
			logging.Warnf("Ignoring invalid explorer URL [%s]\n", explorer)
		}
		explorer = networks.DefaultExplorerURL
	}
	return backendManifest{servers, explorer}, len(servers) > 0
}

// validEndpointURL accepts only https:// URLs with a host that end in "/" and
// carry no credentials, query or fragment.
func validEndpointURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && strings.HasPrefix(s, "https://") && strings.HasSuffix(s, "/") &&
		u.Host != "" && u.User == nil && u.RawQuery == "" && u.Fragment == ""
}

func writeBackendCache(cachePath string, m backendManifest) error {
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
