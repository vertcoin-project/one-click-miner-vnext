package payouts

type Payout interface {
	GetID() int
	GetName() string
	GetTicker() string
	GetPassword() string
}

func GetPayouts(testnet bool) []Payout {
	if testnet {
		return []Payout{
			NewVTCPayout(),
		}
	}
	return []Payout{
		NewVTCPayout(),
		NewDOGEPayout(),
		NewBTCPayout(),
		NewLTCPayout(),
		NewBCHPayout(),
		NewDASHPayout(),
	}
}

func GetPayout(payout int, testnet bool) Payout {
	payouts := GetPayouts(testnet)
	for _, p := range payouts {
		if p.GetID() == payout {
			return p
		}
	}
	return payouts[0]
}
