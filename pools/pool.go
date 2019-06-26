package pools

type Pool interface {
	GetPendingPayout() uint64
	GetStratumUrl() string
	GetUsername() string
	GetPassword() string
}
