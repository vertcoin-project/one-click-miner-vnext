package payouts

var _ Payout = &DASHPayout{}

type DASHPayout struct{}

func NewDASHPayout() *DASHPayout {
	return &DASHPayout{}
}

func (p *DASHPayout) GetID() int {
	return 6
}

func (p *DASHPayout) GetName() string {
	return "Dash"
}

func (p *DASHPayout) GetTicker() string {
	return "DASH"
}

func (p *DASHPayout) GetPassword() string {
	return "c=DASH,mc=VTC"
}
