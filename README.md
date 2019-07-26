# One-Click Miner

This is a redevelopment of Vertcoin's [One Click Miner](https://github.com/vertcoin-project/one-click-miner) that adds:

* Built-in Wallet
* Linux Support
* True One-Click design

Want to donate to the development of this software? Send VTC to [VjEBWk2jJBbesrVUhneVCEZ3Lf1t8gEqk6](https://insight.vertcoin.org/address/VjEBWk2jJBbesrVUhneVCEZ3Lf1t8gEqk6).

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

## Milestone #1: MVP (Minimal Viable Product)

* [X] (Password protected) built-in wallet
* [X] Allows sweeping your entire balance to another address
* [X] Uses Vertcoin Insight to retrieve balances
* [X] Uses P2Proxy as mining pool
* [X] Supports lyclMiner (AMD) on Linux
* [X] Supports lyclMiner (AMD) on Windows
* [X] Supports ccminer (Nvidia) on Windows
* [X] Supports ccminer (Nvidia) on Linux

