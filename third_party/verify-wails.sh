#!/usr/bin/env bash
# Checks that third_party/wails is Wails v1.16.9 as published, changed only by
# third_party/wails.patch, plus the PATCHED.md that describes the changes.
#
# Go downloads v1.16.9 into an empty module cache and checks it against the
# public checksum database (sum.golang.org), so the comparison doesn't depend
# on anything already on this machine. The files the patch names must equal
# the download's with the patch applied, and every other file in
# third_party/wails must match the download byte for byte.
#
# Usage: third_party/verify-wails.sh   (needs go, cmp and patch)
set -euo pipefail

version=v1.16.9
# The hash upstream's go.sum recorded for v1.16.9, the same as sum.golang.org's
expected_sum='h1:H0jWILvWU1I3y2MgqwIPjMfD95/nQP9qAuqAsadYaOs='

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

# The files the patch changes, from its "+++ b/<path>" lines
patched=$(sed -n 's|^+++ b/||p' "$patch_file" | tr -d '')
is_patched() { printf '%s
' "$patched" | grep -qxF "$1"; }

fail=0
checked=0
while IFS= read -r -d '' file; do
	rel=${file#"$vendored"/}
	if [ "$rel" = PATCHED.md ] || is_patched "$rel"; then
		continue
	fi
	if [ ! -f "$upstream/$rel" ]; then
		echo "FAIL: $rel is not in Wails $version" >&2
		fail=1
	elif ! cmp -s "$file" "$upstream/$rel"; then
		echo "FAIL: $rel differs from Wails $version" >&2
		fail=1
	fi
	checked=$((checked + 1))
done < <(find "$vendored" -type f -print0)

# The changed files must be upstream's with exactly the stored patch applied
mkdir "$work/expected"
for rel in $patched; do
	mkdir -p "$work/expected/$(dirname "$rel")"
	cp "$upstream/$rel" "$work/expected/$rel"
done
chmod -R u+w "$work/expected"
if ! patch --quiet --batch --fuzz=0 -p1 -d "$work/expected" <"$patch_file" >/dev/null 2>&1; then
	echo "FAIL: third_party/wails.patch doesn't apply to Wails $version" >&2
	fail=1
else
	for rel in $patched; do
		if ! cmp -s "$work/expected/$rel" "$vendored/$rel"; then
			echo "FAIL: $rel is not Wails $version with third_party/wails.patch applied" >&2
			fail=1
		fi
	done
fi

if [ "$fail" -ne 0 ]; then
	exit 1
fi
echo "OK: $checked files match Wails $version exactly, and $(printf '%s
' "$patched" | wc -l | tr -d ' ') files are $version plus third_party/wails.patch:"
printf '  %s
' $patched
