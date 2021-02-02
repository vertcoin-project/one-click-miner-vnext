package pools

import (
	"time"
)

var _ Pool = &Suprnova{}

type Suprnova struct {
	Address           string
	LastFetchedPayout time.Time
	LastPayout        uint64
}

func NewSuprnova(addr string) *Suprnova {
	return &Suprnova{Address: addr}
}

func (p *Suprnova) GetPendingPayout() uint64 {
	return 0 // TODO
}

func (p *Suprnova) GetStratumUrl() string {
	return "stratum+tcp://vtc.suprnova.cc:1777"
}

func (p *Suprnova) GetUsername() string {
	return p.Address
}

func (p *Suprnova) GetPassword() string {
	return "x"
}

func (p *Suprnova) GetID() int {
	return 4
}

func (p *Suprnova) GetName() string {
	return "Suprnova"
}
