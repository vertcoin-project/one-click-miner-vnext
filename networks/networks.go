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
}

var Active Network

func SetNetwork(testnet bool) {
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
			OCMBackend:         "https://ocm-backend.blkidx.org/",
			Bech32Prefix:       "vtc",
			P2ProxyStratum:     getClosestNodeStratum(),
			P2ProxyURL:         getClosestNodeURL(),
			WalletDB:           "wallet-testnet.db",
		}
	}
}
