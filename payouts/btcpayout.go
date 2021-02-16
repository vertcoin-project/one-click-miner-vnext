package payouts

var _ Payout = &BTCPayout{}

type BTCPayout struct {}

func NewBTCPayout() *BTCPayout {
	return &BTCPayout{}
}

func (p *BTCPayout) GetID() int {
	return 2
}

func (p *BTCPayout) GetName() string {
	return "Bitcoin"
}

func (p *BTCPayout) GetTicker() string {
	return "BTC"
}

func (p *BTCPayout) GetPassword() string {
	return "c=BTC,mc=VTC"
}
