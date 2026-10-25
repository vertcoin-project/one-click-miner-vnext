#!/usr/bin/env bash
# Checks that third_party/wails is Wails v1.16.9 as published, changed only by
# third_party/wails.patch, plus the PATCHED.md that describes the change.
#
# Go downloads v1.16.9 into an empty module cache and checks it against the
# public checksum database (sum.golang.org), so the comparison doesn't depend
# on anything already on this machine. Every other file in third_party/wails
# must match the download byte for byte.
#
# Usage: third_party/verify-wails.sh   (needs go, cmp and patch)
set -euo pipefail

version=v1.16.9
# The hash upstream's go.sum recorded for v1.16.9, the same as sum.golang.org's
expected_sum='h1:H0jWILvWU1I3y2MgqwIPjMfD95/nQP9qAuqAsadYaOs='
patched=lib/renderer/webview.go

here=$(cd "$(dirname "$0")" && pwd)
vendored=$here/wails
patch_file=$here/wails.patch

work=$(mktemp -d)
trap 'chmod -R u+w "$work" 2>/dev/null; rm -rf "$work"' EXIT

# Go on Windows needs a Windows path, Git Bash's mktemp gives a POSIX one
native() {
	if command -v cygpath >/dev/null 2>&1; then cygpath -m "$1"; else printf '%s' "$1"; fi
}

echo "Downloading github.com/wailsapp/wails@$version and checking it against sum.golang.org"
json=$(cd "$work" && GOMODCACHE="$(native "$work")/modcache" GOFLAGS=-modcacherw \
	GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org GONOSUMDB= GONOSUMCHECK= \
	GOPRIVATE= GOINSECURE= go mod download -json "github.com/wailsapp/wails@$version")
sum=$(printf '%s\n' "$json" | sed -n 's/^[[:space:]]*"Sum": "\(.*\)",*$/\1/p')
if [ "$sum" != "$expected_sum" ]; then
	echo "FAIL: downloaded Wails $version has hash '$sum', expected $expected_sum" >&2
	exit 1
fi
upstream=$work/modcache/github.com/wailsapp/wails@$version

fail=0
checked=0
while IFS= read -r -d '' file; do
	rel=${file#"$vendored"/}
	case $rel in
	PATCHED.md | "$patched") continue ;;
	esac
	if [ ! -f "$upstream/$rel" ]; then
		echo "FAIL: $rel is not in Wails $version" >&2
		fail=1
	elif ! cmp -s "$file" "$upstream/$rel"; then
		echo "FAIL: $rel differs from Wails $version" >&2
		fail=1
	fi
	checked=$((checked + 1))
done < <(find "$vendored" -type f -print0)

# The one changed file must be upstream's with exactly the stored patch applied
if ! patch --quiet --batch -o "$work/expected" "$upstream/$patched" <"$patch_file" >/dev/null 2>&1; then
	echo "FAIL: third_party/wails.patch doesn't apply to Wails $version $patched" >&2
	fail=1
elif ! cmp -s "$work/expected" "$vendored/$patched"; then
	echo "FAIL: $patched is not Wails $version with third_party/wails.patch applied" >&2
	fail=1
fi

if [ "$fail" -ne 0 ]; then
	exit 1
fi
echo "OK: $checked files match Wails $version exactly, and $patched is $version plus third_party/wails.patch"
