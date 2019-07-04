#!/bin/bash
GITVER=$(git describe --always --long --dirty)
mv tracking/version.go tracking/version.go.dev
echo "package tracking" > tracking/version.go
echo "var version=\"$1-$GITVER\"" >> tracking/version.go
wails build
rm tracking/version.go
mv tracking/version.go.dev tracking/version.go 
zip vertcoin-ocm-$1-linux-x64.zip ./vertcoin-ocm