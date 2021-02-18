package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/vertiond/verthash-one-click-miner/keyfile"
	"github.com/vertiond/verthash-one-click-miner/networks"
	"github.com/vertiond/verthash-one-click-miner/util"
)

// SignMyInputs finds the inputs in a transaction that came from our own wallet, and signs them with our private keys.
// Will modify the transaction in place, but will ignore inputs that we can't sign and leave them unsigned.
func (w *Wallet) SignMyInputs(tx *wire.MsgTx, password string) error {

	// For now using only P2PKH signing - since we generate
	// a legacy address. Will have to use segwit stuff at some point

	// generate tx-wide hashCache for segwit stuff
	// might not be needed (non-witness) but make it anyway
	// hCache := txscript.NewTxSigHashes(tx)

	// make the stashes for signatures / witnesses
	sigStash := make([][]byte, len(tx.TxIn))
	witStash := make([][][]byte, len(tx.TxIn))

	// get key
	privBytes, err := keyfile.LoadPrivateKey(password)
	if err != nil {
		return err
	}

	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), privBytes)

	for i := range tx.TxIn {
		var found bool
		var utxo Utxo
		for _, u := range w.Utxos {
			if u.TxID == (tx.TxIn[i].PreviousOutPoint.Hash.String()) && u.Vout == uint(tx.TxIn[i].PreviousOutPoint.Index) {
				utxo = u
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("Cannot sign input %s/%d - Not present in known UTXOs", tx.TxIn[i].PreviousOutPoint.Hash.String(), tx.TxIn[i].PreviousOutPoint.Index)
		}
		pkScript, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			return err
		}

		sigStash[i], err = txscript.SignatureScript(tx, i, pkScript, txscript.SigHashAll, priv, true)
		if err != nil {
			return err
		}
	}
	// swap sigs into sigScripts in txins
	for i, txin := range tx.TxIn {
		if sigStash[i] != nil {
			txin.SignatureScript = sigStash[i]
		}
		if witStash[i] != nil {
			txin.Witness = witStash[i]
			txin.SignatureScript = nil
		}
	}

	return nil
}

type txSend struct {
	RawTx string `json:"rawtx"`
}

type txSendReply struct {
	TxId string `json:"txid"`
}

func (w *Wallet) Send(tx *wire.MsgTx) (string, error) {
	var b bytes.Buffer
	tx.Serialize(&b)
	s := txSend{
		RawTx: hex.EncodeToString(b.Bytes()),
	}

	r := txSendReply{}

	err := util.PostJson(fmt.Sprintf("%sinsight-vtc-api/tx/send", networks.Active.InsightURL), s, &r)
	if err != nil {
		return "", err
	}

	w.MarkInputsAsInternallySpent(tx)

	return r.TxId, err
}
