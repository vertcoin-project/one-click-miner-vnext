package networks

type Network struct {
	Base58P2PKHVersion byte
	Base58P2SHVersion  byte
	ExplorerURL        string // Insight or Blockbook; transactions open at <ExplorerURL>tx/<txid>
	Bech32Prefix       string
	P2ProxyStratum     string
	P2ProxyURL         string
	WalletDB           string
	OCMBackend         string
	BackendServers     []string
}

var Active Network

// DefaultBackendServers and DefaultExplorerURL are the mainnet endpoints used
// when neither backends.json nor its cached copy provides them.
var DefaultBackendServers = []string{
	"https://ocmbackend.javerity.com/",
	"https://ocmbackend.mindcraftblocks.com/"}

const DefaultExplorerURL = "https://insight.vertcoin.org/"

var mainnetBackendServers = DefaultBackendServers
var mainnetExplorerURL = DefaultExplorerURL
var testnetActive bool

// SetMainnetEndpoints replaces the mainnet backend list and explorer URL.
// They also apply to Active when the mainnet is selected.
func SetMainnetEndpoints(servers []string, explorerURL string) {
	mainnetBackendServers = servers
	mainnetExplorerURL = explorerURL
	if !testnetActive {
		Active.BackendServers = servers
		Active.ExplorerURL = explorerURL
	}
}

func SetNetwork(testnet bool) {
	testnetActive = testnet
	if testnet {
		Active = Network{
			Base58P2PKHVersion: 74,
			Base58P2SHVersion:  196,
			ExplorerURL:        "https://vtc-insight-testnet.gertjaap.org/",
			Bech32Prefix:       "tvtc",
			P2ProxyStratum:     "stratum+tcp://p2proxy-testnet.gertjaap.org:9171",
			P2ProxyURL:         "https://p2proxy-testnet.gertjaap.org/",
			WalletDB:           "wallet-testnet.db",
		}
	} else {
		Active = Network{
			Base58P2PKHVersion: 71,
			Base58P2SHVersion:  5,
			ExplorerURL:        mainnetExplorerURL,
			Bech32Prefix:       "vtc",
			P2ProxyStratum:     "stratum+tcp://p2proxy.vertcoin.org:9171",
			P2ProxyURL:         "https://p2proxy.vertcoin.org/",
			WalletDB:           "wallet-testnet.db",
			BackendServers:     mainnetBackendServers,
		}
	}
}
