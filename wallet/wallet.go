package wallet

import (
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/btcsuite/btcutil/base58"
	"github.com/mit-dci/lit/bech32"
	"github.com/narula/btcd/txscript"

	"github.com/btcsuite/btcd/wire"
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
	Spent        bool
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

func (w *Wallet) PrepareSweep(addr string) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(2)
	totalIn := uint64(0)
	for _, u := range w.Utxos {
		if !(u.IsCoinbase && u.Height+101 > w.TipHeight) {
			totalIn += u.Amount
			pkScript, _ := hex.DecodeString(u.ScriptPubKey)
			h, _ := chainhash.NewHashFromStr(u.TxID)
			tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, uint32(u.Vout)), pkScript, nil))
		}
	}

	if strings.HasPrefix(addr, "V") {
		pubKeyHash, _, err := base58.CheckDecode(addr)
		if err != nil {
			return nil, fmt.Errorf("Invalid address")
		}
		if len(pubKeyHash) != 20 {
			return nil, fmt.Errorf("Invalid address")
		}
		p2pkhScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).
			AddOp(txscript.OP_HASH160).AddData(pubKeyHash).
			AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).Script()
		if err != nil {
			return nil, fmt.Errorf("Script failure")
		}
		tx.AddTxOut(wire.NewTxOut(0, p2pkhScript))
	} else if strings.HasPrefix(addr, "3") {
		scriptHash, _, err := base58.CheckDecode(addr)
		if err != nil {
			return nil, fmt.Errorf("Invalid address")
		}
		if len(scriptHash) != 20 {
			return nil, fmt.Errorf("Invalid address")
		}
		p2shScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(scriptHash).AddOp(txscript.OP_EQUAL).Script()
		if err != nil {
			return nil, fmt.Errorf("Script failure")
		}
		tx.AddTxOut(wire.NewTxOut(0, p2shScript))
	} else if strings.HasPrefix(addr, "vtc1") {
		script, err := bech32.SegWitAddressDecode(addr)
		if err != nil {
			return nil, fmt.Errorf("Invalid address: %s", err.Error())
		}
		tx.AddTxOut(wire.NewTxOut(int64(totalIn), script))
	} else {
		return nil, fmt.Errorf("Invalid address")
	}

	// Core sends transactions with less than min relay fee, find out what the correct
	// formula is for this.
	size := tx.SerializeSizeStripped()
	logging.Debugf("Transaction size is %d bytes\n", size)
	fee := uint64(size * 100)
	logging.Debugf("Setting fee to %d\n", fee)

	if fee < 100000 { // min relay fee
		fee = 100000
		logging.Debugf("Setting fee to %d\n", fee)
	}

	tx.TxOut[0].Value = int64(totalIn - fee)
	if tx.TxOut[0].Value < 50000 {
		return nil, fmt.Errorf("Insufficient funds")
	}
	return tx, nil
}

func DirectWPKHScriptFromPKH(pkh [20]byte) []byte {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0).AddData(pkh[:])
	b, _ := builder.Script()
	return b
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
