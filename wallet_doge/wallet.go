package wallet

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/tidwall/buntdb"
	"github.com/vertiond/verthash-one-click-miner/logging"
	"github.com/vertiond/verthash-one-click-miner/networks"
	"github.com/vertiond/verthash-one-click-miner/util"
	"github.com/vertiond/verthash-one-click-miner/util/bech32"
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
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("%sapi/v2/get_tx_unspent/DOGETEST/%s", networks.Active.InsightURL, w.Address), &jsonPayload)
	json_parse_success := false
	if err == nil {
		jsonData, ok := jsonPayload["data"].(map[string]interface{})
		if ok {
			jsonDataTxArr, ok := jsonData["txs"].([]interface{})
			if ok {
				json_parse_success = true
				for _, jsonDataTxInfo := range jsonDataTxArr {
					jsonDataTxInfoMap := jsonDataTxInfo.(map[string]interface{})
					utxo_txid, ok1 := jsonDataTxInfoMap["txid"].(string)
					utxo_vout, ok2 := jsonDataTxInfoMap["output_no"].(float64)
					tx_value_in_dogecoin_str, ok3 := jsonDataTxInfoMap["value"].(string)
					if !ok1 || !ok2 || !ok3 {
						json_parse_success = false
						break
					}
					tx_value_in_dogecoin_float, _ := strconv.ParseFloat(tx_value_in_dogecoin_str, 64)
					utxo_amount := uint64(math.Round(tx_value_in_dogecoin_float * float64(100000000)))
					u := Utxo{utxo_txid, uint(utxo_vout), utxo_amount}
					utxos = append(utxos, u)
				}
			}
		}
	}
	if !json_parse_success {
		if err != nil {
			logging.Errorf("Error fetching UTXOs from DOGE Backend: %s", err.Error())
		} else {
			logging.Errorf("Error fetching UTXOs from DOGE Backend")
		}
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

		// Vertcoin fee calculation //
		// fee := uint64(vSizeInt * 100)

		// Dogecoin fee calculation //
		// Base fee is 1 DOGE
		fee := uint64(1)
		// Each additional 1000 bytes incurs 1 DOGE added fee
		fee += uint64(math.Floor(float64(vSizeInt) / float64(1000)))
		// 1 DOGE added fee if total transaction amount is dust
		if (totalIn - fee*100000000) < 100000000 { // UTXO Amount is in Satoshis
			fee += 1
		}
		fee *= 100000000 // Convert fee from DOGE to Satoshis

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
	jsonPayload := map[string]interface{}{}
	err := util.GetJson(fmt.Sprintf("%sapi/v2/get_address_balance/DOGETEST/%s", networks.Active.InsightURL, w.Address), &jsonPayload)
	json_parse_success := false
	if err == nil {
		jsonData, ok := jsonPayload["data"].(map[string]interface{})
		if ok {
			balance_confirmed_in_doge_str, ok1 := jsonData["confirmed_balance"].(string)
			balance_unconfirmed_in_doge_str, ok2 := jsonData["unconfirmed_balance"].(string)
			if ok1 && ok2 {
				balance_confirmed_in_doge_float, _ := strconv.ParseFloat(balance_confirmed_in_doge_str, 64)
				balance_unconfirmed_in_doge_float, _ := strconv.ParseFloat(balance_unconfirmed_in_doge_str, 64)
				balance_spendable := uint64(math.Round((balance_confirmed_in_doge_float + balance_unconfirmed_in_doge_float) * float64(100000000)))
				balance_maturing := uint64(0)
				bal = BalanceResponse{balance_spendable, balance_maturing}
				json_parse_success = true
			}
		}
	}
	if !json_parse_success {
		if err != nil {
			logging.Errorf("Error fetching balance from backend: %s", err.Error())
		} else {
			logging.Errorf("Error fetching balance from backend")
		}
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
