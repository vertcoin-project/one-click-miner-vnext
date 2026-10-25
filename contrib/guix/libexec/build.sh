#!/usr/bin/env bash
# Runs inside the Guix container, with no network access: builds the frontend
# and the binary for $HOST from the source archive and the packages fetch.sh
# downloaded, and writes the zip and its SHA256SUMS.part to /outdir.
export LC_ALL=C
export TZ=UTC
set -o errexit -o pipefail -o nounset
umask 0022

: "${HOST:?}" "${NAME:?}" "${SOURCE_DATE_EPOCH:?}" "${SRC_TAR:?}" "${JOBS:?}"
VERSION="${VERSION:-}"
export SOURCE_DATE_EPOCH

# Given a package name and an output name, return the path of that output in
# the container's profile (from Bitcoin Core's contrib/guix/libexec/prelude.bash)
store_path() {
    grep --extended-regexp "/[^-]{32}-${1}-[^-]+${2:+-${2}}" "${GUIX_ENVIRONMENT}/manifest" \
        | head --lines=1 \
        | sed --expression='s|\x29*$||' \
              --expression='s|^[[:space:]]*"||' \
              --expression='s|"[[:space:]]*$||'
}

# Point the cross compiler at its own libc and nothing else
cross_paths() {
    local libc="$1" kernel="${2:-}" gcc_lib_store gcc_libs gcc_lib
    gcc_lib_store="$(store_path "gcc-cross-${HOST}" lib)"
    gcc_libs=( "${gcc_lib_store}/lib/gcc/${HOST}"/* )
    gcc_lib="${gcc_libs[0]}"
    export CROSS_C_INCLUDE_PATH="${gcc_lib}/include:${gcc_lib}/include-fixed:${libc}/include${kernel:+:${kernel}/include}"
    export CROSS_LIBRARY_PATH="${gcc_lib_store}/lib:${gcc_lib}:${libc}/lib"
    unset LIBRARY_PATH CPATH C_INCLUDE_PATH CPLUS_INCLUDE_PATH CROSS_CPLUS_INCLUDE_PATH
    # Guix's ld wrapper otherwise adds an RPATH entry for every -L, pointing
    # the binary at /gnu/store
    export GUIX_LD_WRAPPER_DISABLE_RPATH=yes
}

export HOME=/tmp/home
mkdir -p "$HOME"
tar -xf "$SRC_TAR" -C /distsrc
cd /distsrc/src

export GOTOOLCHAIN=local
export GOMODCACHE=/cache/gomod
export GOCACHE=/distsrc/gocache
export GOPROXY=off
export GOSUMDB=off
export GOFLAGS="-mod=readonly -modcacherw -trimpath -buildvcs=false"
export GOMAXPROCS="$JOBS"
go mod verify

echo "== frontend"
(
    cd frontend
    npm ci --ignore-scripts --offline --no-audit --no-fund --cache /cache/npm
    # `npm run build`, called directly: the container has no /usr/bin/env for
    # the #!/usr/bin/env node line of node_modules/.bin/vue-cli-service.
    # Webpack 4 hashes with md4, which Node 17+ only has in the legacy provider.
    NODE_ENV=production NODE_OPTIONS=--openssl-legacy-provider \
        node node_modules/@vue/cli-service/bin/vue-cli-service.js build
)
sha256sum frontend/dist/app.js frontend/dist/app.css

ldflags="-s -w -buildid= -X github.com/wailsapp/wails.BuildMode=prod"
if [ -n "$VERSION" ]; then
    ldflags+=" -X github.com/vertcoin-project/one-click-miner-vnext/tracking.version=${VERSION}-g$(basename "$SRC_TAR" .tar | sed 's/^vertcoin-ocm-//')"
fi

mkdir -p out
case "$HOST" in
    x86_64-linux-gnu)
        cross_paths "$(store_path "glibc-cross-${HOST}")" "$(store_path "linux-libre-headers-cross-${HOST}")"
        # gdk-3.0.pc names -lz without a -L for it. Add zlib's own directory
        # rather than the profile's lib/, which would put its libraries ahead
        # of the cross glibc's
        export CGO_LDFLAGS="-L$(dirname "$(readlink -f "${GUIX_ENVIRONMENT}/lib/libz.so")")"
        echo "== go build ($HOST)"
        # --allow-shlib-undefined: WebKitGTK from the Guix store is built against
        # a newer glibc than the one this binary targets. The binary loads the
        # distribution's WebKitGTK at run time, so only its sonames matter here.
        CC="${HOST}-gcc" CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
            go build -toolexec "bash $(dirname "${BASH_SOURCE[0]}")/strip-rpath.sh" -o out/vertcoin-ocm -ldflags "$ldflags -linkmode=external -extldflags '-Wl,--dynamic-linker=/lib64/ld-linux-x86-64.so.2 -Wl,--allow-shlib-undefined -Wl,--as-needed -Wl,-O2'" .

        # The binary has to run outside Guix: the system loader, no store paths
        # to search, and nothing newer than glibc 2.31
        interp="$("${HOST}-readelf" -l out/vertcoin-ocm | sed -n 's/.*Requesting program interpreter: \(.*\)]/\1/p')"
        [ "$interp" = /lib64/ld-linux-x86-64.so.2 ] || { echo "ERR: interpreter is '$interp'" >&2; exit 1; }
        if "${HOST}-readelf" -d out/vertcoin-ocm | grep -E 'RPATH|RUNPATH'; then
            echo "ERR: the binary has a search path" >&2; exit 1
        fi
        newest="$("${HOST}-readelf" -V --wide out/vertcoin-ocm | grep -o 'GLIBC_[0-9.]*' | sort -uV | tail -1)"
        printf '%s\n' "$newest" GLIBC_2.31 | sort -CV || { echo "ERR: needs $newest" >&2; exit 1; }
        echo "interpreter $interp, newest glibc symbol $newest, needs:"
        "${HOST}-readelf" -d out/vertcoin-ocm | sed -n 's/.*(NEEDED).*\[\(.*\)\]/  \1/p'

        binary=vertcoin-ocm
        zipname="vertcoin-ocm-${NAME}-linux-x64.zip"
        ;;
    x86_64-w64-mingw32)
        echo "== windows resources"
        # What `wails build` generates: an icon from appicon.png, Wails' manifest,
        # and an .rc naming both, compiled to a .syso that go build links in.
        # third_party/wails leaves out cmd/packages, so contrib/guix/windows
        # holds copies of Wails v1.16.9's wails.exe.manifest and wails.rc
        CGO_ENABLED=0 go run contrib/guix/winicon.go appicon.png vertcoin-ocm.ico
        cp contrib/guix/windows/wails.exe.manifest vertcoin-ocm.exe.manifest
        sed 's/\$NAME\$/vertcoin-ocm/g' contrib/guix/windows/wails.rc > vertcoin-ocm.rc
        cross_paths "$(store_path mingw-w64-x86_64-winpthreads)"
        "${HOST}-windres" -o vertcoin-ocm-res.syso vertcoin-ocm.rc
        sha256sum vertcoin-ocm.ico vertcoin-ocm-res.syso

        echo "== go build ($HOST)"
        CC="${HOST}-gcc" CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
            go build -o out/vertcoin-ocm.exe -ldflags "$ldflags -H windowsgui -extldflags '-Wl,--no-insert-timestamp'" .
        binary=vertcoin-ocm.exe
        zipname="vertcoin-ocm-${NAME}-windows-x64.zip"
        ;;
    *)
        echo "ERR: unsupported host $HOST" >&2
        exit 1
        ;;
esac

# A zip records each file's time, permissions and owner; fix all three
cd out
chmod 0755 "$binary"
touch --no-dereference --date="@${SOURCE_DATE_EPOCH}" "$binary"
zip -X -D -9 "/outdir/${zipname}" "$binary"
cp "$binary" "/outdir/${binary}"
cd /outdir
sha256sum "$zipname" "$binary" > SHA256SUMS.part
cat SHA256SUMS.part
