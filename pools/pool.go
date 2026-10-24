package pools

import (
	"sync"
)

type Pool interface {
	GetPendingPayout() uint64
	GetStratumUrl() string
	GetUsername() string
	GetPassword() string
	GetName() string
	GetID() int
	GetFee() float64
	OpenBrowserPayoutInfo(addr string)
}

// P2PoolType marks the manifest entry for the built-in P2Pool, whose stratum,
// fee and payouts come from the selected P2Pool node.
const P2PoolType = "p2pool"

// Config describes one mainnet pool in pools.json. An entry with Type
// P2PoolType needs no other fields. In PayoutURL and DashboardURL, {address}
// is replaced by the wallet address. The pending payout is the sum of the
// PayoutFields, each a dot-separated path into the PayoutURL's JSON response,
// in VTC.
type Config struct {
	Type         string   `json:"type,omitempty"`
	ID           int      `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	StratumURL   string   `json:"stratumURL,omitempty"`
	Password     string   `json:"password,omitempty"`
	Fee          *float64 `json:"fee,omitempty"`
	PayoutURL    string   `json:"payoutURL,omitempty"`
	PayoutFields []string `json:"payoutFields,omitempty"`
	DashboardURL string   `json:"dashboardURL,omitempty"`
}

func fee(f float64) *float64 { return &f }

// DefaultConfigs is the mainnet pool list used when neither pools.json nor its
// cached copy is available.
var DefaultConfigs = []Config{
	{
		ID:           4,
		Name:         "Suprnova.cc",
		StratumURL:   "stratum+tcp://vtc.suprnova.cc:1776",
		Password:     "x",
		Fee:          fee(1.00),
		PayoutURL:    "https://vtc.suprnova.cc/index.php?page=api&action=getuserbalance&api_key={address}",
		PayoutFields: []string{"getuserbalance.data.confirmed", "getuserbalance.data.unconfirmed"},
		DashboardURL: "https://vtc.suprnova.cc/index.php?page=anondashboard&user={address}",
	},
	{Type: P2PoolType},
	{
		ID:           6,
		Name:         "Zpool.ca",
		StratumURL:   "stratum+tcp://verthash.mine.zpool.ca:6144",
		Password:     "c=VTC,zap=VTC",
		Fee:          fee(0.50),
		PayoutURL:    "https://zpool.ca/api/wallet?address={address}",
		PayoutFields: []string{"unpaid"},
		DashboardURL: "https://zpool.ca/wallet/{address}",
	},
	{
		ID:           9,
		Name:         "MiningpoolSweden.eu",
		StratumURL:   "stratum+tcp://vtc.miningpoolsweden.eu:3052",
		Password:     "x",
		Fee:          fee(0.6),
		PayoutURL:    "https://api.miningpoolsweden.eu/api/pools/vert1/miners/{address}",
		PayoutFields: []string{"pendingBalance"},
		DashboardURL: "https://miningpoolsweden.eu/?#vert1/dashboard?address={address}",
	},
}

var configsMu sync.RWMutex
var mainnetConfigs = DefaultConfigs

// SetConfigs replaces the mainnet pool list.
func SetConfigs(configs []Config) {
	configsMu.Lock()
	defer configsMu.Unlock()
	mainnetConfigs = configs
}

func GetPools(addr string, testnet bool) []Pool {
	if testnet {
		return []Pool{
			NewP2Proxy(addr),
		}
	}
	configsMu.RLock()
	defer configsMu.RUnlock()
	pools := make([]Pool, 0, len(mainnetConfigs))
	for _, c := range mainnetConfigs {
		if c.Type == P2PoolType {
			pools = append(pools, NewP2Pool(addr))
		} else {
			pools = append(pools, NewConfiguredPool(addr, c))
		}
	}
	return pools
}

func GetPool(pool int, addr string, testnet bool) Pool {
	pools := GetPools(addr, testnet)
	for _, p := range pools {
		if p.GetID() == pool {
			return p
		}
	}
	return pools[0]
}

// HasPool reports whether the pool with this ID is on the current list.
func HasPool(pool int, addr string, testnet bool) bool {
	for _, p := range GetPools(addr, testnet) {
		if p.GetID() == pool {
			return true
		}
	}
	return false
}
