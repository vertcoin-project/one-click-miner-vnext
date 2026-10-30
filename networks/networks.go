package networks

type Network struct {
	Base58P2PKHVersion byte
	Base58P2SHVersion  byte
	ExplorerURL        string // Insight or Blockbook; transactions open at <ExplorerURL>tx/<txid>
	Bech32Prefix       string
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

// SetMainnetEndpoints replaces the mainnet backend list and explorer URL, and
// applies them to Active.
func SetMainnetEndpoints(servers []string, explorerURL string) {
	mainnetBackendServers = servers
	mainnetExplorerURL = explorerURL
	Active.BackendServers = servers
	Active.ExplorerURL = explorerURL
}

func SetNetwork() {
	Active = Network{
		Base58P2PKHVersion: 71,
		Base58P2SHVersion:  5,
		ExplorerURL:        mainnetExplorerURL,
		Bech32Prefix:       "vtc",
		BackendServers:     mainnetBackendServers,
	}
}
