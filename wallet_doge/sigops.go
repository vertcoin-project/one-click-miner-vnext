package wallet

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

var WitnessScaleFactor = 4

// CountSigOps returns the number of signature operations for all transaction
// input and output scripts in the provided transaction.  This uses the
// quicker, but imprecise, signature operation counting mechanism from
// txscript.
func CountSigOps(tx *btcutil.Tx) int {
	msgTx := tx.MsgTx()

	// Accumulate the number of signature operations in all transaction
	// inputs.
	totalSigOps := 0
	for _, txIn := range msgTx.TxIn {
		numSigOps := txscript.GetSigOpCount(txIn.SignatureScript)
		totalSigOps += numSigOps
	}

	// Accumulate the number of signature operations in all transaction
	// outputs.
	for _, txOut := range msgTx.TxOut {
		numSigOps := txscript.GetSigOpCount(txOut.PkScript)
		totalSigOps += numSigOps
	}

	return totalSigOps
}

func (w *Wallet) GetSigOpCost(tx *btcutil.Tx, pkScript []byte, isCoinBaseTx bool, bip16, segWit bool) (int, error) {
	numSigOps := CountSigOps(tx) * WitnessScaleFactor
	if bip16 {
		numP2SHSigOps, err := w.CountP2SHSigOps(tx, isCoinBaseTx)
		if err != nil {
			return 0, nil
		}
		numSigOps += (numP2SHSigOps * WitnessScaleFactor)
	}

	if segWit && !isCoinBaseTx {
		msgTx := tx.MsgTx()
		for _, txIn := range msgTx.TxIn {
			witness := txIn.Witness
			sigScript := txIn.SignatureScript
			numSigOps += txscript.GetWitnessSigOpCount(sigScript, pkScript, witness)
		}

	}

	return numSigOps, nil
}

// CountP2SHSigOps returns the number of signature operations for all input
// transactions which are of the pay-to-script-hash type.  This uses the
// precise, signature operation counting mechanism from the script engine which
// requires access to the input transaction scripts.
func (w *Wallet) CountP2SHSigOps(tx *btcutil.Tx, isCoinBaseTx bool) (int, error) {
	// We never spend P2SH in OCM, so this can be disabled
	return 0, nil
	/*
		// Coinbase transactions have no interesting inputs.
		if isCoinBaseTx {
			return 0, nil
		}

		// Accumulate the number of signature operations in all transaction
		// inputs.
		msgTx := tx.MsgTx()
		totalSigOps := 0
		for txInIndex, txIn := range msgTx.TxIn {
			// Ensure the referenced input transaction is available.
			utxo := w.GetUtxo(txIn.PreviousOutPoint.Hash.String(), uint(txIn.PreviousOutPoint.Index))
			if utxo.TxID == "" {
				str := fmt.Sprintf("output %v referenced from "+
					"transaction %s:%d either does not exist or "+
					"has already been spent", txIn.PreviousOutPoint,
					tx.Hash(), txInIndex)
				return 0, fmt.Errorf(str)
			}

			// We're only interested in pay-to-script-hash types, so skip
			// this input if it's not one.
			pkScript, _ := hex.DecodeString(utxo.ScriptPubKey)
			if !txscript.IsPayToScriptHash(pkScript) {
				continue
			}

			// Count the precise number of signature operations in the
			// referenced public key script.
			sigScript := txIn.SignatureScript
			numSigOps := txscript.GetPreciseSigOpCount(sigScript, pkScript,
				true)

			// We could potentially overflow the accumulator so check for
			// overflow.
			lastSigOps := totalSigOps
			totalSigOps += numSigOps
			if totalSigOps < lastSigOps {
				str := fmt.Sprintf("the public key script from output "+
					"%v contains too many signature operations - "+
					"overflow", txIn.PreviousOutPoint)
				return 0, fmt.Errorf(str)
			}
		}

		return totalSigOps, nil*/
}
