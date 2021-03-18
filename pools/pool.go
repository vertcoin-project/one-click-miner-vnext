package pools

type Pool interface {
	GetPendingPayout(addr string) uint64
	GetStratumUrl() string
	GetPassword() string
	GetName() string
	GetID() int
	GetFee() float64
	OpenBrowserPayoutInfo(addr string)
}

func GetPools(testnet bool) []Pool {
	if testnet {
		return []Pool{
			NewP2Proxy(),
		}
	}
	return []Pool{
		NewZergpool(),
		//NewHashalot(),
		//NewSuprnova(),
		//NewP2Pool(),
	}
}

func GetPool(pool int, testnet bool) Pool {
	pools := GetPools(testnet)
	for _, p := range pools {
		if p.GetID() == pool {
			return p
		}
	}
	return pools[0]
}
