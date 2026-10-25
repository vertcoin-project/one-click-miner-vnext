#!/usr/bin/env bash
# Runs inside the Guix container, with network access: downloads every Go
# module and npm package the build needs into /cache. Go checks each module
# against go.sum and the checksum database, npm checks each package against the
# integrity hash in frontend/package-lock.json, so the cache can only hold what
# the commit names.
export LC_ALL=C
set -o errexit -o pipefail -o nounset
umask 0022

export HOME=/tmp/home
mkdir -p "$HOME" /tmp/work
tar -xf "${SRC_TAR:?}" -C /tmp/work
cd /tmp/work/src

export SSL_CERT_DIR="${GUIX_ENVIRONMENT}/etc/ssl/certs"
export SSL_CERT_FILE="${SSL_CERT_DIR}/ca-certificates.crt"
export GOTOOLCHAIN=local
export GOMODCACHE=/cache/gomod
export GOCACHE=/tmp/gocache
export GOFLAGS="-modcacherw -mod=readonly"
export GOPROXY=https://proxy.golang.org
export GOSUMDB=sum.golang.org

go mod download
go mod verify

cd frontend
# --ignore-scripts: the preinstall script runs npx npm-force-resolutions, which
# downloads a package the lockfile doesn't pin and rewrites the lockfile
npm ci --ignore-scripts --no-audit --no-fund --cache /cache/npm
