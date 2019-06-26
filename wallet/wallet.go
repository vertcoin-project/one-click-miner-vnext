package wallet

import (
	"fmt"
	"path"

	"github.com/tidwall/buntdb"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
)

type Wallet struct {
	Address   string
	Utxos     []Utxo
	TipHeight uint
	db        *buntdb.DB
}

type Utxo struct {
	TxID         string `json:"txid"`
	Vout         uint   `json:"vout"`
	ScriptPubKey string `json:"scriptPubKey"`
	Amount       uint64 `json:"satoshis"`
	Height       uint   `json:"height"`
	IsCoinbase   bool
}

type Status struct {
	Height uint `json:"height"`
}

type Tx struct {
	IsCoinBase bool `json:"isCoinBase"`
}

func NewWallet(addr string) (*Wallet, error) {
	logging.Infof("Initializing wallet %s", addr)
	db, err := buntdb.Open(path.Join(util.DataDirectory(), "wallet.db"))
	if err != nil {
		return nil, err
	}
	return &Wallet{Address: addr, db: db}, nil
}

// Update will rescan the chain for UTXOs on the wallet's address
// It will fetch the transaction details for each UTXO to determine
// if it is a coinbase, and cache that
func (w *Wallet) Update() {
	utxos := []Utxo{}
	err := util.GetJson(fmt.Sprintf("https://insight.vertcoin.org/insight-vtc-api/addr/%s/utxo", w.Address), &utxos)
	if err != nil {
		logging.Errorf("Error fetching UTXOs from Insight: %s", err.Error())
		return
	}
	w.Utxos = utxos

	w.UpdateCoinbaseStatus()

	status := Status{}
	err = util.GetJson("https://insight.vertcoin.org/insight-vtc-api/sync", &status)
	if err != nil {
		logging.Errorf("Error fetching sync status of Insight: %s", err.Error())
		return
	}
	w.TipHeight = status.Height
}

func (w *Wallet) UpdateCoinbaseStatus() {
	for i := range w.Utxos {
		w.Utxos[i].IsCoinbase = w.IsCoinbase(w.Utxos[i].TxID)
	}
}

func (w *Wallet) IsCoinbase(txid string) bool {
	coinBase := ""
	err := w.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(fmt.Sprintf("coinbase-%s", txid))
		coinBase = v
		return err
	})
	if err == nil {
		return coinBase == "1"
	}

	isTx := Tx{}
	err = util.GetJson(fmt.Sprintf("https://insight.vertcoin.org/insight-vtc-api/tx/%s", txid), &isTx)
	if err != nil {
		logging.Errorf("Error fetching coinbase status of TX from Insight: %s", err.Error())
		return false
	}
	coinBase = "0"
	if isTx.IsCoinBase {
		coinBase = "1"
	}

	err = w.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(fmt.Sprintf("coinbase-%s", txid), coinBase, nil)
		return err
	})
	if err != nil {
		logging.Errorf("Error writing coinbase status to database: %s", err.Error())
		return false
	}
	return coinBase == "1"
}

// GetBalance will scan the utxos in the wallet and return
// two values: mature and immature balance. Mining outputs
// need to wait for 101 confirmations before being allowed
// to spend
func (w *Wallet) GetBalance() (bal uint64, balImmature uint64) {
	for _, u := range w.Utxos {
		if u.IsCoinbase && u.Height+101 > w.TipHeight {
			balImmature += u.Amount
		} else {
			bal += u.Amount
		}
	}
	return
}
