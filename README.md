# One-Click Miner

The One-Click Miner allows cryptocurrency enthusiasts to get into mining with minimal effort. When you download the One-Click Miner, you will be asked to provide a password for your (built in) wallet. It will then immediately commence mining.

This is a redevelopment of Vertcoin's [One Click Miner](https://github.com/vertcoin-project/one-click-miner).

This software is available for Windows and Linux.

## FAQ

###Which GPUs are supported?

Please refer to this list of [supported hardware.](https://github.com/CryptoGraphics/VerthashMiner#supported-hardware)

###I have an error message that reads 'Failure to configure'

You may need to add an exclusion to your antivirus / Windows Defender.

###My GPU is supported but an error messages reads 'no compatible GPUs'

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

If you want to support the further development of the One Click Miner, feel free to donate Vertcoin to [Vmnbtn5nnNbs1otuYa2LGBtEyFuarFY1f8](https://insight.vertcoin.org/address/Vmnbtn5nnNbs1otuYa2LGBtEyFuarFY1f8).
