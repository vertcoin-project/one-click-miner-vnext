package pools

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

func GetPools(addr string, testnet bool) []Pool {
	if testnet {
		return []Pool{
			NewP2Proxy(addr),
		}
	}
	return []Pool{
		NewHashalot(addr),
		NewZergpool(addr),
		NewSuprnova(addr),
		NewP2Pool(addr),
		Newzpool(addr),
		NewBBQDroid(addr),
		NewAcidpool(addr),
		NewMiningpoolSweden(addr),
		NewCoinMinerz(addr),
	}
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
