package wallet

import (
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/tidwall/buntdb"
	"github.com/vertcoin-project/one-click-miner-vnext/logging"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/vertcoin-project/one-click-miner-vnext/util/bech32"
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

// Insight's limit is 100kB HEX (so 50kB raw bytes) - limiting this to 45kB. Once we have
// better backends (an insight that performs better and allows higher limits, an integrated
// full node or just using Vertcoin Core) we can scale this up.
var maxTxSize = 45000

type Status struct {
	Height uint `json:"height"`
}

type Tx struct {
	IsCoinBase bool `json:"isCoinBase"`
}

func NewWallet(addr string) (*Wallet, error) {
	logging.Infof("Initializing wallet %s", addr)
	db, err := buntdb.Open(filepath.Join(util.DataDirectory(), "wallet.db"))
	if err != nil {
		return nil, err
	}
	return &Wallet{Address: addr, db: db}, nil
}

func (w *Wallet) PrepareSweep(addr string) ([]*wire.MsgTx, error) {
	retArr := make([]*wire.MsgTx, 0)
	for {
		tx := wire.NewMsgTx(2)
		totalIn := uint64(0)
		for _, u := range w.Utxos {
			alreadyIncluded := false
			for _, t := range retArr {
				for _, i := range t.TxIn {
					if i.PreviousOutPoint.Hash.String() == u.TxID && i.PreviousOutPoint.Index == uint32(u.Vout) {
						alreadyIncluded = true
						break
					}
				}
			}

			if !alreadyIncluded && !(u.IsCoinbase && u.Height+101 > w.TipHeight) {
				totalIn += u.Amount
				pkScript, _ := hex.DecodeString(u.ScriptPubKey)
				h, _ := chainhash.NewHashFromStr(u.TxID)
				tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, uint32(u.Vout)), pkScript, nil))
			}
		}

		if strings.HasPrefix(addr, "V") {
			pubKeyHash, _, err := base58.CheckDecode(addr)
			if err != nil {
				return nil, fmt.Errorf("invalid_address")
			}
			if len(pubKeyHash) != 20 {
				return nil, fmt.Errorf("invalid_address")
			}
			p2pkhScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_DUP).
				AddOp(txscript.OP_HASH160).AddData(pubKeyHash).
				AddOp(txscript.OP_EQUALVERIFY).AddOp(txscript.OP_CHECKSIG).Script()
			if err != nil {
				return nil, fmt.Errorf("script_failure")
			}
			tx.AddTxOut(wire.NewTxOut(0, p2pkhScript))
		} else if strings.HasPrefix(addr, "3") {
			scriptHash, _, err := base58.CheckDecode(addr)
			if err != nil {
				return nil, fmt.Errorf("invalid_address")
			}
			if len(scriptHash) != 20 {
				return nil, fmt.Errorf("invalid_address")
			}
			p2shScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_HASH160).AddData(scriptHash).AddOp(txscript.OP_EQUAL).Script()
			if err != nil {
				return nil, fmt.Errorf("script_failure")
			}
			tx.AddTxOut(wire.NewTxOut(0, p2shScript))
		} else if strings.HasPrefix(addr, "vtc1") {
			script, err := bech32.SegWitAddressDecode(addr)
			if err != nil {
				return nil, fmt.Errorf("invalid_address")
			}
			tx.AddTxOut(wire.NewTxOut(int64(totalIn), script))
		} else {
			return nil, fmt.Errorf("invalid_address")
		}

		for i := range tx.TxIn {
			tx.TxIn[i].SignatureScript = make([]byte, 107) // add dummy signature to properly calculate size
		}

		// Weight = (stripped_size * 4) + witness_size formula,
		// using only serialization with and without witness data. As witness_size
		// is equal to total_size - stripped_size, this formula is identical to:
		// weight = (stripped_size * 3) + total_size.
		logging.Debugf("Transaction raw serialize size is %d\n", tx.SerializeSize())
		logging.Debugf("Transaction serialize size stripped is %d\n", tx.SerializeSizeStripped())

		chunked := false
		// Chunk if needed
		if tx.SerializeSize() > maxTxSize {
			chunked = true
			// Remove some extra inputs so we have enough for the next TX to remain valid, we
			// want to have enough money to create an output with enough value
			valueRemoved := uint64(0)
			for tx.SerializeSize() > maxTxSize || valueRemoved < 100000 {
				for _, u := range w.Utxos {
					if u.TxID == tx.TxIn[len(tx.TxIn)-1].PreviousOutPoint.Hash.String() &&
						uint32(u.Vout) == tx.TxIn[len(tx.TxIn)-1].PreviousOutPoint.Index {
						totalIn -= u.Amount
						valueRemoved += u.Amount
					}
				}
				tx.TxIn = tx.TxIn[:len(tx.TxIn)-1]
			}
		}

		txWeight := (tx.SerializeSizeStripped() * 3) + tx.SerializeSize()
		logging.Debugf("Transaction weight is %d\n", txWeight)
		btcTx := btcutil.NewTx(tx)

		sigOpCost, err := w.GetSigOpCost(btcTx, false, true, true)
		if err != nil {
			return nil, fmt.Errorf("could_not_calculate_fee")
		}
		logging.Debugf("Transaction sigop cost is %d\n", sigOpCost)

		vSize := (math.Max(float64(txWeight), float64(sigOpCost*20)) + float64(3)) / float64(4)
		logging.Debugf("Transaction vSize is %.4f\n", vSize)
		vSizeInt := uint64(vSize + float64(0.5)) // Round Up
		logging.Debugf("Transaction vSizeInt is %d\n", vSizeInt)

		fee := uint64(vSizeInt * 100)
		logging.Debugf("Setting fee to %d\n", fee)

		// empty out the dummy sigs
		for i := range tx.TxIn {
			tx.TxIn[i].SignatureScript = nil
		}

		tx.TxOut[0].Value = int64(totalIn - fee)
		if tx.TxOut[0].Value < 50000 {
			return nil, fmt.Errorf("insufficient_funds")
		}
		retArr = append(retArr, tx)

		if !chunked {
			break
		}
	}
	return retArr, nil
}

func (w *Wallet) GetUtxo(txid string, pout uint) Utxo {
	for _, u := range w.Utxos {
		if u.TxID == txid && u.Vout == pout {
			return u
		}
	}
	return Utxo{}
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
	w.UpdateSpentStatus()
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

func (w *Wallet) UpdateSpentStatus() {
	for i := range w.Utxos {
		w.Utxos[i].Spent = w.IsSpent(w.Utxos[i].TxID, w.Utxos[i].Vout)
	}
}

func (w *Wallet) IsSpent(txid string, vout uint) bool {
	spent := ""
	w.db.View(func(tx *buntdb.Tx) error {
		v, err := tx.Get(fmt.Sprintf("spent-%s-%09d", txid, vout))
		spent = v
		return err
	})
	return spent == "1"
}

func (w *Wallet) MarkSpent(txid string, vout uint) {
	w.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(fmt.Sprintf("spent-%s-%09d", txid, vout), "1", nil)
		return err
	})
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
		if !u.Spent {
			if u.IsCoinbase && u.Height+101 > w.TipHeight {
				balImmature += u.Amount
			} else {
				bal += u.Amount
			}
		}
	}
	return
}

func (w *Wallet) MarkInputsAsInternallySpent(tx *wire.MsgTx) {
	for _, txi := range tx.TxIn {
		w.MarkSpent(txi.PreviousOutPoint.Hash.String(), uint(txi.PreviousOutPoint.Index))
	}
	w.UpdateSpentStatus()
}
