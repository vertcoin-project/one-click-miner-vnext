package backend

import (
	"fmt"

	"github.com/vertiond/verthash-one-click-miner/keyfile"
	"github.com/vertiond/verthash-one-click-miner/logging"
	"github.com/vertiond/verthash-one-click-miner/networks"
	"github.com/vertiond/verthash-one-click-miner/tracking"
	"github.com/vertiond/verthash-one-click-miner/util"
	"github.com/vertiond/verthash-one-click-miner/wallet_doge"
)

func (m *Backend) WalletInitialized() int {
	logging.Infof("Checking wallet..")
	checkWallet := 0
	if keyfile.KeyFileValid() {
		checkWallet = 1
	}
	script, err := keyfile.GetScript()
	if err != nil {
		logging.Errorf("Error initializing wallet: %s", err.Error())
	}
	wal, err := wallet.NewWallet(keyfile.GetAddress(), script)
	if err != nil {
		logging.Errorf("Error initializing wallet: %s", err.Error())
	}
	m.wal = wal
	logging.Infof("Wallet initialized: %d", checkWallet)
	return checkWallet
}

func (m *Backend) SendSweep(password string) []string {
	tracking.Track(tracking.TrackingRequest{
		Category: "Sweep",
		Action:   "Send",
	})

	txids := make([]string, 0)

	if len(m.pendingSweep) == 0 {
		// Somehow user managed to press send without properly
		// preparing the sweep first
		return []string{"send_failed"}
	}

	for _, s := range m.pendingSweep {
		err := m.wal.SignMyInputs(s, password)
		if err != nil {
			logging.Errorf("Error signing transaction: %s", err.Error())
			return []string{"sign_failed"}
		}

		txHash, err := m.wal.Send(s)
		if err != nil {
			logging.Errorf("Error sending transaction: %s", err.Error())
			return []string{"send_failed"}
		}
		txids = append(txids, txHash)
	}

	m.pendingSweep = nil

	logging.Debugf("Transaction(s) sent! TXIDs: %v\n", txids)
	m.refreshBalanceChan <- true
	return txids

}

func (m *Backend) ShowTx(txid string) {
	util.OpenBrowser(fmt.Sprintf("%stx/DOGE/%s", networks.Active.InsightURL, txid))
}

type PrepareResult struct {
	FormattedAmount      string
	NumberOfTransactions int
}

func (m *Backend) PrepareSweep(addr string) string {
	tracking.Track(tracking.TrackingRequest{
		Category: "Sweep",
		Action:   "Prepare",
	})

	logging.Debugf("Preparing sweep")

	txs, err := m.wal.PrepareSweep(addr)
	if err != nil {
		logging.Errorf("Error preparing sweep: %v", err)
		return err.Error()
	}

	m.pendingSweep = txs
	val := float64(0)
	for _, tx := range txs {
		val += (float64(tx.TxOut[0].Value) / float64(100000000))
	}

	result := PrepareResult{fmt.Sprintf("%0.8f DOGE", val), len(txs)}
	logging.Debugf("Prepared sweep: %v", result)

	m.runtime.Events.Emit("createTransactionResult", result)
	return ""
}

func (m *Backend) Address() string {
	return keyfile.GetAddress()
}

func (m *Backend) InitWallet(password string) bool {
	tracking.Track(tracking.TrackingRequest{
		Category: "Wallet",
		Action:   "Initialize",
	})

	err := keyfile.CreateKeyFile(password)
	if err == nil {
		m.WalletInitialized()
		m.ResetPool()
		m.ResetPayout()
		m.ResetWalletAddress()
		return true
	}
	logging.Errorf("Error: %s", err.Error())
	return false
}
