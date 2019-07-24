package backend

import (
	"path/filepath"

	"github.com/tidwall/buntdb"

	"github.com/btcsuite/btcd/wire"
	"github.com/vertcoin-project/one-click-miner-vnext/miners"
	"github.com/vertcoin-project/one-click-miner-vnext/pools"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/vertcoin-project/one-click-miner-vnext/wallet"
	"github.com/wailsapp/wails"
)

type Backend struct {
	runtime             *wails.Runtime
	wal                 *wallet.Wallet
	settings            *buntdb.DB
	pendingSweep        []*wire.MsgTx
	minerBinaries       []*miners.BinaryRunner
	rapidFailures       []*miners.BinaryRunner
	pool                pools.Pool
	refreshBalanceChan  chan bool
	refreshHashChan     chan bool
	refreshRunningState chan bool
	stopMonitoring      chan bool
	stopHash            chan bool
	stopBalance         chan bool
	stopRunningState    chan bool
	prerequisiteInstall chan bool
}

func NewBackend() (*Backend, error) {
	db, err := buntdb.Open(filepath.Join(util.DataDirectory(), "settings.db"))
	if err != nil {
		return nil, err
	}

	return &Backend{
		settings:            db,
		refreshBalanceChan:  make(chan bool),
		refreshHashChan:     make(chan bool),
		refreshRunningState: make(chan bool),
		stopHash:            make(chan bool),
		stopBalance:         make(chan bool),
		stopRunningState:    make(chan bool),
		stopMonitoring:      make(chan bool),
		prerequisiteInstall: make(chan bool),
		minerBinaries:       []*miners.BinaryRunner{},
		rapidFailures:       []*miners.BinaryRunner{},
	}, nil
}

func (m *Backend) WailsInit(runtime *wails.Runtime) error {
	// Save runtime
	m.runtime = runtime

	go func() {
		for pi := range m.prerequisiteInstall {
			send := "0"
			if pi {
				send = "1"
			}
			m.runtime.Events.Emit("prerequisiteInstall", send)
		}
	}()

	return nil
}

func (m *Backend) OpenDownloadUrl(url string) {
	util.OpenBrowser(url)
}
