#!/usr/bin/env bash
# go build -toolexec wrapper for the Linux build. Guix's Go appends
# "-r <gcc's lib directory>" to every link (cmd/go/internal/work/gc.go), which
# becomes a RUNPATH into /gnu/store. Drop that one argument pair when the tool
# is the linker; run every other tool unchanged.
tool="$1"
shift
if [ "$(basename "$tool")" != link ]; then
    exec "$tool" "$@"
fi
args=()
while [ $# -gt 0 ]; do
    if [ "$1" = -r ] && [ $# -gt 1 ] && [[ "$2" == /gnu/store/*-gcc-*-lib/lib ]]; then
        shift 2
        continue
    fi
    args+=("$1")
    shift
done
exec "$tool" "${args[@]}"
