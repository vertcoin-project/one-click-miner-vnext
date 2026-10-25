# Reproducible builds with Guix

These scripts build the One Click Miner for Linux and Windows (x86-64) the
same way on every machine: anyone who builds the same commit gets
byte-for-byte the same zips. Then they sign a list of the zips' hashes, so
people can check that a release really came from this source.

- `guix-build` builds the zips.
- `guix-sign` checks them, lists their hashes in `manifest-<version>.txt`
  and signs that file.

## Before you start (once per machine)

1. **Linux, or WSL 2 on Windows**, with
   [Guix](https://guix.gnu.org/manual/en/html_node/Binary-Installation.html)
   installed and `guix-daemon` running.
2. **A clone of this repository** in Linux (in WSL, in your Linux home folder,
   not under `/mnt/c`).
3. **Your GPG key**, to sign releases. On Windows, keep it in
   [GnuPG](https://gnupg.org/download/) or [Gpg4win](https://gpg4win.org/);
   `guix-sign` uses that from WSL. Tell git which key it is, in the Linux
   clone:

   ```
   git config user.signingkey <your key ID>
   ```

   Find your key ID with `gpg --list-secret-keys --keyid-format long`. It is the
   part after `rsa3072/` or `ed25519/`, like `A7BEB2621678D37D`.

The first build takes a long time (an hour or more): Guix builds the
compilers first. Later builds reuse them and take minutes.

## Making a release, step by step

The example is version `2.2.2`; use your own.

**1. Tag the commit** (on Windows or Linux, wherever you commit):

```
git tag -s 2.2.2 -m 2.2.2
git push origin 2.2.2
```

**2. Check out the tag in the Linux clone:**

```
cd ~/one-click-miner-vnext
git fetch origin --tags --force
git checkout 2.2.2
```

**3. Build:**

```
./contrib/guix/guix-build
```

It names everything after the tag on the commit, so you don't need to give
the version. When it finishes, the zips are in `guix-build-2.2.2/output/`.

**4. Sign, and copy the release files to Windows:**

```
RELEASE_DIR=/mnt/c/Users/<you>/Downloads/2.2.2 ./contrib/guix/guix-sign
```

It checks the zips, writes `manifest-2.2.2.txt`, and asks for your GPG
passphrase to sign it. Then it copies four files to `Downloads\2.2.2`:

| file | what it is |
|---|---|
| `vertcoin-ocm-2.2.2-linux-x64.zip` | the Linux release |
| `vertcoin-ocm-2.2.2-windows-x64.zip` | the Windows release |
| `manifest-2.2.2.txt` | the SHA-256 of each zip and of the program inside it |
| `manifest-2.2.2.txt.sig` | your signature of the manifest |

**5. Upload** those four files to the GitHub release.

### If you move a tag

If you delete a tag and create it again on a newer commit, build again (step
3). `guix-sign` refuses a build of the old commit even though it has the same
name. Delete the old folder first:

```
rm -rf guix-build-2.2.2
```

## Checking a release

Anyone can check a release by downloading the four files into one folder and
running:

```
gpg --verify manifest-2.2.2.txt.sig manifest-2.2.2.txt
sha256sum --check --ignore-missing manifest-2.2.2.txt
```

The first command should say `Good signature` from the release key. The second
should say `OK` for each zip. On Windows, `Get-FileHash <zip>` in PowerShell
prints the hash to compare with the manifest by eye.

To be sure the zips came from the source, build the same tag yourself (steps 2
and 3) and compare: your `guix-build-<version>/output/*/SHA256SUMS.part` files
should have the same hashes as the manifest.

## Options

Options are environment variables and go **before** the command:

```
VERSION=2.3.0 ./contrib/guix/guix-build     # right
./contrib/guix/guix-build VERSION=2.3.0     # wrong: ignored
```

Usually you don't need any.

**`guix-build`**

| option | what it does |
|---|---|
| `VERSION` | the version to name the zips after and to build into the program (it shows as `2.3.0-g<commit>`). By default, the tag on the commit. With no tag, or with `VERSION=` (empty), the program keeps the source's dev version and everything is named after the commit |
| `HOSTS` | `x86_64-linux-gnu` or `x86_64-w64-mingw32` to build only one. By default, both |
| `JOBS` | how many jobs to run at once. By default, one per CPU |
| `CACHE_PATH` | where downloaded Go and npm packages are kept between builds. By default, `./guix-cache` |

**`guix-sign`**

| option | what it does |
|---|---|
| `RELEASE_DIR` | a folder to copy the four release files into |
| `SIGNER` | the GPG key ID to sign with. By default, git's `user.signingkey`. Unset, gpg's default key, which with several keys may not be the one you want |
| `GPG` | the gpg program to sign with. By default, the Linux `gpg` if it holds a secret key; otherwise, from WSL, Windows' GnuPG in `C:\Program Files\GnuPG` |
| `VERSION` | the version the build was named after, if you set one for `guix-build` |

## Troubleshooting

- **The zips are named after a commit, not the version.** The commit had no
  tag, or the version was written after the command. Tag it, or put
  `VERSION=...` before the command, delete the `guix-build-...` folder and
  build again.
- **"the working tree has uncommitted changes".** Only committed files are
  built. Commit or stash your changes.
- **"already exists; remove it to rebuild".** Delete that build's folder.
- **"not as 2.2.2" or "was not built from HEAD"** from `guix-sign`. The
  build doesn't match the checkout: build again.
- **gpg says "waiting for lock".** Another gpg process is stuck. In Windows
  PowerShell, run `gpgconf --kill all`, then sign again.

## How it works

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
4. **Sign** (`guix-sign`): checks each output against its `SHA256SUMS.part`,
   joins them into `manifest-<version>.txt` (Linux, then Windows, as for
   2.1.5) and writes a binary detached signature of it.

Every compiler, Go, Node and library comes from one pinned Guix commit, and
the build runs in a container with no network and nothing from the host.

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
