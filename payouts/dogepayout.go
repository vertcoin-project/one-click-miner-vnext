package payouts

var _ Payout = &DOGEPayout{}

type DOGEPayout struct {}

func NewDOGEPayout() *DOGEPayout {
	return &DOGEPayout{}
}

func (p *DOGEPayout) GetID() int {
	return 4
}

func (p *DOGEPayout) GetName() string {
	return "Dogecoin"
}

func (p *DOGEPayout) GetTicker() string {
	return "DOGE"
}

func (p *DOGEPayout) GetPassword() string {
	return "c=DOGE,mc=VTC"
}
