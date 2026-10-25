# Reproducible builds with Guix

`contrib/guix/guix-build` builds the One Click Miner for Linux and Windows
(x86-64) so that anyone building the same commit gets byte-identical zips.
Every compiler, Go, Node and library comes from one pinned Guix commit, and the
build runs in a container with no network and nothing from the host.

## Requirements

- Linux (WSL 2 works) with [Guix](https://guix.gnu.org/manual/en/html_node/Binary-Installation.html)
  installed and `guix-daemon` running
- git, and a clean checkout of the commit to build

## Building

```
./contrib/guix/guix-build
```

Outputs land in `guix-build-<commit>/output/<host>/`: the zip, the bare binary
and a `SHA256SUMS.part`. Compare those hashes with someone else's build of the
same commit.

| variable | meaning |
|---|---|
| `HOSTS` | `x86_64-linux-gnu`, `x86_64-w64-mingw32`, or both (the default) |
| `VERSION` | e.g. `2.3.0`: stamped into the binary as `2.3.0-g<commit>` and used in the zip names. Unset, the binary keeps the source's dev version and the zips are named after the commit |
| `JOBS` | parallel jobs, default `nproc` |
| `CACHE_PATH` | where Go modules and npm packages are kept, default `./guix-cache` |

The first run takes a long time: it builds the GCC cross toolchains and may
build WebKitGTK if no substitute is available. Later runs reuse them.

## What it does

1. `git archive HEAD` makes the source tarball, so only committed files are
   built. The script refuses to run with uncommitted changes.
2. **Fetch** (`libexec/fetch.sh`, the only step with network): `go mod
   download` and `npm ci --ignore-scripts` into `CACHE_PATH`. Go checks each
   module against `go.sum`, npm checks each package against the integrity hash
   in `frontend/package-lock.json`.
3. **Build** (`libexec/build.sh`, one container per host, no network):
   `npm ci --offline` and `npm run build` for the frontend, which `main.go`
   embeds, then `go build -trimpath -buildvcs=false` with the flags `wails
   build` would use. Windows also gets the icon, manifest and `.syso` resource
   `wails build` generates (`winicon.go` repeats Wails' icon code, and
   `windows/` holds Wails v1.16.9's manifest and `.rc`, which
   `third_party/wails` leaves out). The zip's timestamp is the commit's time.

The Linux binary is linked with GCC against glibc 2.31 and the system loader,
with no Guix store paths. At run time it needs glibc 2.31 or newer and
WebKitGTK 4.1 (`libwebkit2gtk-4.1-0` on Ubuntu 22.04 and later and Debian 12
and later), which the patched Wails in `third_party/wails` links in place of
4.0. The build fails if the binary needs a newer glibc or has a search path
into `/gnu/store`.

Guix's own Go adds `-r <gcc lib dir>` to every link, which would become such a
search path; `libexec/strip-rpath.sh` removes that argument from the Linux link.

`manifest.scm` and `patches/` reuse Bitcoin Core v32.0's cross toolchains
unchanged, and `guix-build` pins the same Guix commit Bitcoin Core v32.0 does
(`c5eee333`), so a machine that has built Bitcoin Core v32 or Vertcoin Core's
32.x branch already has them.
