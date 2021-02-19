package wallet

import (
	"errors"
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
	"github.com/vertcoin-project/one-click-miner-vnext/networks"
	"github.com/vertcoin-project/one-click-miner-vnext/util"
	"github.com/vertcoin-project/one-click-miner-vnext/util/bech32"
)

type Wallet struct {
	Address   string
	Script    []byte
	Spendable uint64
	Maturing  uint64
	db        *buntdb.DB
}

type Utxo struct {
	TxID   string `json:"txid"`
	Vout   uint   `json:"vout"`
	Amount uint64 `json:"satoshis"`
}

// Insight's limit is 100kB HEX (so 50kB raw bytes) - limiting this to 45kB. Once we have
// better backends (an insight that performs better and allows higher limits, an integrated
// full node or just using Vertcoin Core) we can scale this up.
var maxTxSize = 45000

func NewWallet(addr string, script []byte) (*Wallet, error) {
	logging.Infof("Initializing wallet %s", addr)
	db, err := buntdb.Open(filepath.Join(util.DataDirectory(), networks.Active.WalletDB))
	if err != nil {
		return nil, err
	}
	return &Wallet{Address: addr, Script: script, db: db}, nil
}

func (w *Wallet) Utxos() ([]Utxo, error) {
	utxos := []Utxo{}
	err := util.GetJson(fmt.Sprintf("%sutxos/%x", networks.Active.OCMBackend, w.Script), &utxos)
	if err != nil {
		logging.Errorf("Error fetching UTXOs from OCM Backend: %s", err.Error())
		return utxos, err
	}
	return utxos, nil
}

func (w *Wallet) PrepareSweep(addr string) ([]*wire.MsgTx, error) {
	utxos, err := w.Utxos()
	if err != nil {
		return nil, errors.New("backend_failure")
	}
	retArr := make([]*wire.MsgTx, 0)
	for {
		tx := wire.NewMsgTx(2)
		totalIn := uint64(0)
		for _, u := range utxos {
			alreadyIncluded := false
			for _, t := range retArr {
				for _, i := range t.TxIn {
					if i.PreviousOutPoint.Hash.String() == u.TxID && i.PreviousOutPoint.Index == uint32(u.Vout) {
						alreadyIncluded = true
						break
					}
				}
			}
			if alreadyIncluded {
				logging.Debugf("UTXO Already Included: %v", u)
				continue
			}
			totalIn += u.Amount
			h, _ := chainhash.NewHashFromStr(u.TxID)
			tx.AddTxIn(wire.NewTxIn(wire.NewOutPoint(h, uint32(u.Vout)), w.Script, nil))
		}

		if len(tx.TxIn) == 0 {
			logging.Warnf("Trying to sweep with zero UTXOs")
			return nil, errors.New("insufficient_funds")
		}

		hash, version, err := base58.CheckDecode(addr)
		if err == nil && version == networks.Active.Base58P2PKHVersion {
			pubKeyHash := hash
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
		} else if err == nil && version == networks.Active.Base58P2SHVersion {
			scriptHash := hash
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
		} else if strings.HasPrefix(addr, fmt.Sprintf("%s1", networks.Active.Bech32Prefix)) {
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
				for _, u := range utxos {
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

		sigOpCost, err := w.GetSigOpCost(btcTx, w.Script, false, true, true)
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

func DirectWPKHScriptFromPKH(pkh [20]byte) []byte {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_0).AddData(pkh[:])
	b, _ := builder.Script()
	return b
}

type BalanceResponse struct {
	Spendable uint64 `json:"confirmed"`
	Maturing  uint64 `json:"maturing"`
}

// Update will reload balance from the backend
func (w *Wallet) Update() {
	bal := BalanceResponse{}
	err := util.GetJson(fmt.Sprintf("%sbalance/%x", networks.Active.OCMBackend, w.Script), &bal)
	if err != nil {
		logging.Errorf("Error fetching balance from backend: %s", err.Error())
		return
	}
	w.Spendable = bal.Spendable
	w.Maturing = bal.Maturing
}

// GetBalance will scan the utxos in the wallet and return
// two values: mature and immature balance. Mining outputs
// need to wait for 101 confirmations before being allowed
// to spend
func (w *Wallet) GetBalance() (bal uint64, balImmature uint64) {
	return w.Spendable, w.Maturing
}
