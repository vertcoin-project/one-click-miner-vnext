# Verthash One-Click Miner

You will be asked to create a password for your built-in Dogecoin wallet. It will create the Verthash datafile and then immediately begin mining to your Dogecoin wallet address.

**[View your Zergpool payout information here](https://zergpool.com/?address=) (Click 'Copy address' to the right of Spendable Balance)**

`DOGE payment threshold is 400 DOGE and 80 DOGE on Sunday late evening(CET)`

`VTC payment threshold is the equivalent of 0.0001 BTC and 0.00002 BTC on Sunday late evening(CET)`

`BTC payment threshold is 0.0025 BTC `

`LTC payment threshold is 0.25 LTC and 0.05 LTC on Sunday late evening(CET)`

`BCH payment threshold is the equivalent of 0.0001 BTC and 0.00002 BTC on Sunday late evening(CET)`

`DASH payment threshold is the equivalent of 0.0001 BTC and 0.00002 BTC on Sunday late evening(CET)`

*Network and exchange fees may apply

The Verthash One-Click Miner mines [Vertcoin](https://vertcoin.org) and is functionally the same as [upstream](https://github.com/vertcoin-project/one-click-miner-vnext) utilizing a new data directory, `verthash-ocm`.  This is essentially a fancy wrapper for [VerthashMiner](https://github.com/CryptoGraphics/VerthashMiner) which is also open source.

This software is available for Windows and Linux.

## FAQ

### Which GPUs are supported?

Please refer to this list of [supported hardware.](https://github.com/CryptoGraphics/VerthashMiner#supported-hardware)

### I have an error message that reads 'Failure to configure'

You may need to add an exclusion to your antivirus / Windows Defender.

### My GPU is supported but an error messages reads 'no compatible GPUs'

Update your GPU drivers to the latest version.


## Building

The GUI of this MVP is based on [Wails](https://wails.app) and [Go](https://golang.org/).

Install the Wails [prerequisites](https://wails.app/home.html#prerequisites) for your platform, and then run:

```bash
go get github.com/wailsapp/wails/cmd/wails
```

Then clone this repository, and inside its main folder, execute:

```bash
wails build
```

## Donations

If you want to support the further development of the upstream One Click Miner, feel free to donate Vertcoin to [Vmnbtn5nnNbs1otuYa2LGBtEyFuarFY1f8](https://insight.vertcoin.org/address/Vmnbtn5nnNbs1otuYa2LGBtEyFuarFY1f8).
