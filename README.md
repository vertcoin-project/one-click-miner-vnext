# One-Click Miner

The One-Click Miner allows cryptocurrency enthusiasts to get into mining with minimal effort. When you download the One-Click Miner, you will be asked to provide a password for your (built in) wallet. It will then immediately commence mining.

This is a redevelopment of Vertcoin's [One Click Miner](https://github.com/vertcoin-project/one-click-miner).

This software is available for Windows and Linux. Linux needs WebKitGTK 4.1 (`libwebkit2gtk-4.1-0`, available on Ubuntu 22.04+ and Debian 12+).

## FAQ

### Which GPUs are supported?

Please refer to this list of [supported hardware.](https://github.com/CryptoGraphics/VerthashMiner#supported-hardware)

### I have an error message that reads 'Failure to configure' or 'Checks failed'

You may need to add an exclusion to your antivirus / Windows Defender.  The data directory is located at `%AppData%\vertcoin-ocm` on Windows or `~/.vertcoin-ocm` on Linux.

### My GPU is supported but an error messages reads 'no compatible GPUs'

Update your GPU drivers to the latest version.

### My GPU is not being utilized

The OCM is using all of your GPU resources.  To verify, check CUDA usage for Nvidia or Compute 0/1 usage for AMD.

## Building

The GUI is based on [Wails](https://wails.app) v1 and [Go](https://golang.org/). This repository builds against a patched copy of Wails v1.16.9 in [third_party/wails](third_party/wails) (see [PATCHED.md](third_party/wails/PATCHED.md)), which `go.mod` uses in place of the published module.

You need:

  - Go 1.22+
  - Node.js. On Node 17+, set `NODE_OPTIONS=--openssl-legacy-provider` for the Vue CLI 4 frontend build
  - C compiler for cgo: MinGW-w64 on Windows, or on Linux `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` (Ubuntu 22.04+, Debian 12+)
  - Wails v1.16.9 CLI:

```bash
go install github.com/wailsapp/wails/cmd/wails@v1.16.9
```

Then clone this repository, and inside its main folder, execute:

```bash
wails build
```

Use `wails build -d` for a debug build, which also prints what the app is doing to the console and enables the web inspector. Every build logs to `debug.log` in the data directory.

For reproducible Linux and Windows builds, which give byte-identical zips for the same commit, see [contrib/guix](contrib/guix/README.md).

## Donations

If you want to support the further development of the One Click Miner, feel free to donate Vertcoin to [Vmnbtn5nnNbs1otuYa2LGBtEyFuarFY1f8](https://insight.vertcoin.org/address/Vmnbtn5nnNbs1otuYa2LGBtEyFuarFY1f8).
