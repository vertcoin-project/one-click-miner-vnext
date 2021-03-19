package backend

import (
	"fmt"
	"time"

	"github.com/vertiond/verthash-one-click-miner/logging"
	"github.com/vertiond/verthash-one-click-miner/miners"
	"github.com/vertiond/verthash-one-click-miner/payouts"
	"github.com/vertiond/verthash-one-click-miner/tracking"
	"github.com/vertiond/verthash-one-click-miner/util"
)

func (m *Backend) GetArgs() miners.BinaryArguments {
	tracking.Track(tracking.TrackingRequest{
		Category: "Mining",
		Action:   "Switch Pool",
		Name:     fmt.Sprintf("%v", m.pool.GetName()),
	})

	var username string
	var password string
	if m.UseCustomPayout() {
		username = m.customAddress
		password = m.payout.GetPassword()
	} else {
		// Use wallet address (Dogecoin) for payout
		walletPayout := payouts.NewDOGEPayout()
		username = m.walletAddress
		password = walletPayout.GetPassword()
	}

	return miners.BinaryArguments{
		StratumUrl:       m.pool.GetStratumUrl(),
		StratumUsername:  username,
		StratumPassword:  password,
		EnableIntegrated: m.getSetting("enableIntegrated"),
	}
}

func (m *Backend) GetPoolFee() string {
	return fmt.Sprintf("%0.2f%%", m.pool.GetFee())
}

func (m *Backend) GetPoolName() string {
	return m.pool.GetName()
}

func (m *Backend) GetPayoutTicker() string {
	if m.UseCustomPayout() {
		return m.payout.GetTicker()
	}
	return "DOGE"
}

func (m *Backend) PayoutInformation() {
	m.pool.OpenBrowserPayoutInfo(m.GetCurrentMiningAddress())
}

func (m *Backend) StartMining() bool {
	logging.Infof("Starting mining process...")

	tracking.Track(tracking.TrackingRequest{
		Category: "Mining",
		Action:   "Start",
	})

	args := m.GetArgs()

	startProcessMonitoring := make(chan bool)

	go func() {
		<-startProcessMonitoring
		continueLoop := true
		for continueLoop {
			newMinerBinaries := make([]*miners.BinaryRunner, 0)
			for _, br := range m.minerBinaries {
				if br.CheckRunning() == miners.RunningStateRapidFail {
					m.rapidFailures = append(m.rapidFailures, br)
					m.runtime.Events.Emit("minerRapidFail", br.MinerBinary.MainExecutableName)

				} else {
					newMinerBinaries = append(newMinerBinaries, br)
				}
			}

			m.minerBinaries = newMinerBinaries

			select {
			case <-m.stopMonitoring:
				continueLoop = false
			case <-time.After(time.Second):
			}
		}
		logging.Infof("Stopped monitoring thread")
	}()

	go func() {
		unitVtcPerBtc := 0.0
		unitPayoutCoinPerBtc := 0.0
		vtcPayout := payouts.NewVTCPayout()
		btcPayout := payouts.NewBTCPayout()
		var myPayout payouts.Payout
		if m.UseCustomPayout() {
			myPayout = m.payout
		} else {
			// Default Dogecoin payout
			myPayout = payouts.NewDOGEPayout()
		}

		cycles := 0
		nhr := uint64(0)
		continueLoop := true
		for continueLoop {
			if cycles >= 600 {
				cycles = 0
			}
			if cycles == 0 {
				// Don't refresh this every time since we refresh it every second
				// and this pulls from Insight. Every 600s is fine (~every 4 blocks)
				nhr = util.GetNetHash()
				if myPayout.GetName() != vtcPayout.GetName() {
					unitVtcPerBtc = payouts.GetBitcoinPerUnitCoin(vtcPayout.GetName(), vtcPayout.GetTicker(), vtcPayout.GetCoingeckoExchange())
					if myPayout.GetName() == btcPayout.GetName() {
						unitPayoutCoinPerBtc = 1
					} else {
						time.Sleep(750 * time.Millisecond) // Put time between API calls
						unitPayoutCoinPerBtc = payouts.GetBitcoinPerUnitCoin(myPayout.GetName(), myPayout.GetTicker(), myPayout.GetCoingeckoExchange())
					}
					logging.Infof(fmt.Sprintf("Payout exchange rate: VTC/BTC=%0.10f, %s/BTC=%0.10f", unitVtcPerBtc, myPayout.GetTicker(), unitPayoutCoinPerBtc))
				}
			}
			cycles++

			hr := uint64(0)
			for _, br := range m.minerBinaries {
				hr += br.HashRate()
			}
			hashrate := float64(hr) / float64(1000)
			hashrateUnit := "kH/s"
			if hashrate > 1000 {
				hashrate /= 1000
				hashrateUnit = "MH/s"
			}
			if hashrate > 1000 {
				hashrate /= 1000
				hashrateUnit = "GH/s"
			}
			if hashrate > 1000 {
				hashrate /= 1000
				hashrateUnit = "TH/s"
			}
			m.runtime.Events.Emit("hashRate", fmt.Sprintf("%0.2f %s", hashrate, hashrateUnit))

			netHash := float64(nhr) / float64(1000000000)

			m.runtime.Events.Emit("networkHashRate", fmt.Sprintf("%0.2f %s", netHash, hashrateUnit))

			avgEarning := float64(hr) / float64(nhr) * float64(14400) // 14400 = Emission per day. Need to adjust for halving

			// Convert average earning from Vertcoin to selected payout coin
			avgEarningTicker := "VTC"
			if myPayout.GetName() != vtcPayout.GetName() {
				if unitVtcPerBtc != 0 && unitPayoutCoinPerBtc != 0 {
					avgEarningTicker = myPayout.GetTicker()
					avgEarning = avgEarning * unitVtcPerBtc / unitPayoutCoinPerBtc
				}
			}

			// Show at least three significant figures of average earning value
			avgEarningDecimals := 2
			if avgEarning > 0.0 && avgEarning < 1.0 {
				avgEarningScaled := avgEarning
				addDecimals := 1
				for addDecimals = 1; addDecimals <= 10; addDecimals++ {
					avgEarningScaled *= 10
					if avgEarningScaled >= 1.0 {
						break
					}
				}
				avgEarningDecimals += addDecimals
			}
			avgEarningStrfmt := "%0." + fmt.Sprint(avgEarningDecimals) + "f %s"

			m.runtime.Events.Emit("avgEarnings", fmt.Sprintf(avgEarningStrfmt, avgEarning, avgEarningTicker))

			select {
			case <-m.stopHash:
				continueLoop = false
			case <-m.refreshHashChan:
			case <-time.After(time.Second):
			}
		}
	}()

	go func() {
		continueLoop := true
		var pb uint64
		for continueLoop {
			m.wal.Update()
			b, bi := m.wal.GetBalance()
			m.runtime.Events.Emit("balance", fmt.Sprintf("%0.8f", float64(b)/float64(100000000)))
			m.runtime.Events.Emit("balanceImmature", fmt.Sprintf("%0.8f", float64(bi)/float64(100000000)))
			logging.Infof("Updating pending pool payout...")
			var payoutAddr string
			if m.UseCustomPayout() {
				payoutAddr = m.customAddress
			} else {
				payoutAddr = m.walletAddress
			}
			newPb := m.pool.GetPendingPayout(payoutAddr)
			pb = newPb
			m.runtime.Events.Emit("balancePendingPool", fmt.Sprintf("%0.8f", float64(pb)/float64(100000000)))
			select {
			case <-m.stopBalance:
				continueLoop = false
			case <-m.refreshBalanceChan:
			case <-time.After(time.Minute * 5):
			}
		}
	}()

	go func() {
		continueLoop := true
		for continueLoop {

			runningProcesses := 0
			for _, br := range m.minerBinaries {
				if br.IsRunning() {
					runningProcesses++
				}
			}

			m.runtime.Events.Emit("runningMiners", runningProcesses)

			timeout := time.Second * 1
			if runningProcesses > 0 {
				timeout = time.Second * 10
			}
			select {
			case <-m.stopRunningState:
				continueLoop = false
			case <-m.refreshRunningState:
			case <-time.After(timeout):
			}
		}
	}()

	for _, br := range m.minerBinaries {
		err := br.Start(args)
		if err != nil {
			m.StopMining()
			logging.Errorf("Failure to start %s: %s\n", br.MinerBinary.MainExecutableName, err.Error())
			return false
		}
	}

	startProcessMonitoring <- true

	return true
}

func (m *Backend) RefreshBalance() {

	m.refreshBalanceChan <- true
}

func (m *Backend) RefreshHashrate() {

	m.refreshHashChan <- true
}

func (m *Backend) RefreshRunningState() {

	m.refreshRunningState <- true
}

func (m *Backend) StopMining() bool {
	tracking.Track(tracking.TrackingRequest{
		Category: "Mining",
		Action:   "Stop",
	})
	select {
	case m.stopMonitoring <- true:
	default:
	}
	logging.Infof("Stopping mining process...")
	for _, br := range m.minerBinaries {
		br.Stop()
	}
	select {
	case m.stopBalance <- true:
	default:
	}
	select {
	case m.stopHash <- true:
	default:
	}
	select {
	case m.stopRunningState <- true:
	default:
	}
	return true
}
