package networks

type Network struct {
	Base58P2PKHVersion byte
	Base58P2SHVersion  byte
	InsightURL         string
	Bech32Prefix       string
	P2ProxyStratum     string
	P2ProxyURL         string
	WalletDB           string
	OCMBackend         string
	BackendServers     []string
}

var Active Network

// DefaultBackendServers is the mainnet backend list used when neither
// backends.json nor its cached copy is available.
var DefaultBackendServers = []string{
	"https://ocmbackend.javerity.com/",
	"https://ocmbackend.mindcraftblocks.com/"}

var mainnetBackendServers = DefaultBackendServers
var testnetActive bool

// SetBackendServers replaces the mainnet backend list. It also applies to
// Active when the mainnet is selected.
func SetBackendServers(servers []string) {
	mainnetBackendServers = servers
	if !testnetActive {
		Active.BackendServers = servers
	}
}

func SetNetwork(testnet bool) {
	testnetActive = testnet
	if testnet {
		Active = Network{
			Base58P2PKHVersion: 74,
			Base58P2SHVersion:  196,
			InsightURL:         "https://vtc-insight-testnet.gertjaap.org/",
			Bech32Prefix:       "tvtc",
			P2ProxyStratum:     "stratum+tcp://p2proxy-testnet.gertjaap.org:9171",
			P2ProxyURL:         "https://p2proxy-testnet.gertjaap.org/",
			WalletDB:           "wallet-testnet.db",
		}
	} else {
		Active = Network{
			Base58P2PKHVersion: 71,
			Base58P2SHVersion:  5,
			InsightURL:         "https://insight.vertcoin.org/",
			Bech32Prefix:       "vtc",
			P2ProxyStratum:     "stratum+tcp://p2proxy.vertcoin.org:9171",
			P2ProxyURL:         "https://p2proxy.vertcoin.org/",
			WalletDB:           "wallet-testnet.db",
			BackendServers:     mainnetBackendServers,
		}
	}
}
