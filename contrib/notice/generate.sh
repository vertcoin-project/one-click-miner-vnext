#!/usr/bin/env bash
# Writes NOTICE, the license and notice files of everything built into
# vertcoin-ocm, which the release zips carry next to the program. Run it from
# the repository root, after `npm ci` in frontend/, whenever a Go module or a
# frontend package changes; CI fails when NOTICE doesn't match what it writes.
# Go's own license is taken from the Go that runs this, so run it with the Go
# version go.mod names, which CI and the Guix build use, such as
# GOTOOLCHAIN=go1.22.12.
#
# A package that ships no license file gets its files from
# contrib/notice/missing/, under its module path or package name.
# contrib/notice/extra/ holds what is neither, such as the font the frontend
# inlines into app.css.
set -euo pipefail
export LC_ALL=C

here=contrib/notice

# Go and npm give Windows paths with backslashes
slashes() { tr '\\' '/'; }

# The files in a directory, not below it, that hold a license or notice
licenses() {
    find "$1" -maxdepth 1 -type f \( -iname 'licen[cs]e*' -o -iname 'copying*' -o -iname 'notice*' \) | sort
}

heading() {
    printf '\n\n%s\n%s\n' "$1" "${1//?/=}"
}

# Each file in a directory, under its path relative to that directory
files() {
    local base=$1 f
    shift
    for f in "$@"; do
        printf '\n--- %s\n\n' "${f#"$base"/}"
        sed '$a\' "$f"
    done
}

# A package's own license files, else those from contrib/notice/missing/
own_or_missing() {
    local name=$1 dir=$2 found
    found="$(licenses "$dir")"
    if [ -n "$found" ]; then
        [ ! -d "$here/missing/$name" ] || { echo "ERR: $name now ships a license; remove $here/missing/$name" >&2; exit 1; }
        # shellcheck disable=SC2086 # the paths have no spaces
        files "$dir" $found
    else
        found="$(licenses "$here/missing/$name" 2>/dev/null || true)"
        [ -n "$found" ] || { echo "ERR: $name has no license file; add it under $here/missing/$name" >&2; exit 1; }
        # shellcheck disable=SC2086
        files "$here/missing/$name" $found
    fi
}

# Every package of every module in the Windows and Linux builds: module, module directory, package directory
pkgs="$(for os in windows linux; do
    CGO_ENABLED=1 GOOS=$os GOARCH=amd64 go list -deps \
        -f '{{with .Module}}{{if not .Main}}{{.Path}}{{"\t"}}{{with .Replace}}{{.Dir}}{{else}}{{.Dir}}{{end}}{{"\t"}}{{$.Dir}}{{end}}{{end}}' .
done | slashes | sort -u)"

{
    cat <<'EOF'
Third-party notices
===================

vertcoin-ocm is built from Go, the Go modules, the frontend packages and the
font below. Each name is followed by the license and notice files it is
published with, as they are. contrib/notice/generate.sh writes this file.
EOF

    goroot="$(go env GOROOT | slashes)"
    heading "Go"
    files "$goroot" "$goroot/LICENSE"

    for mod in $(cut -f1 <<<"$pkgs" | sort -u); do
        root="$(awk -F'\t' -v m="$mod" '$1 == m { print $2; exit }' <<<"$pkgs")"
        heading "$mod"
        own_or_missing "$mod" "$root"
        # Packages below the module's root can carry licenses of their own, such as Wails' webview
        # shellcheck disable=SC2046
        files "$root" $(awk -F'\t' -v m="$mod" '$1 == m && $3 != $2 { print $3 }' <<<"$pkgs" |
            while read -r dir; do licenses "$dir"; done)
    done

    for dir in $(cd frontend && npm ls --omit=dev --all --parseable | tail -n +2 | slashes | sort -u); do
        name="$(node -p "require('$dir/package.json').name")"
        heading "$name (frontend)"
        own_or_missing "$name" "$dir"
    done

    for dir in "$here"/extra/*/; do
        dir=${dir%/}
        heading "$(basename "$dir")"
        # shellcheck disable=SC2046
        files "$dir" $(licenses "$dir")
    done
} > NOTICE
