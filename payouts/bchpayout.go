package payouts

var _ Payout = &BCHPayout{}

type BCHPayout struct{}

func NewBCHPayout() *BCHPayout {
	return &BCHPayout{}
}

func (p *BCHPayout) GetID() int {
	return 5
}

func (p *BCHPayout) GetName() string {
	return "Bitcoin Cash"
}

func (p *BCHPayout) GetTicker() string {
	return "BCH"
}

func (p *BCHPayout) GetPassword() string {
	return "c=BCH,mc=VTC"
}
