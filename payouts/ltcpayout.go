package payouts

var _ Payout = &LTCPayout{}

type LTCPayout struct {}

func NewLTCPayout() *LTCPayout {
	return &LTCPayout{}
}

func (p *LTCPayout) GetID() int {
	return 3
}

func (p *LTCPayout) GetName() string {
	return "Litecoin"
}

func (p *LTCPayout) GetTicker() string {
	return "LTC"
}

func (p *LTCPayout) GetPassword() string {
	return "c=LTC,mc=VTC"
}

func (p *LTCPayout) GetCoingeckoExchange() string {
	return "bittrex"
}
