#!/bin/bash
GITVER=$(git describe --always --long --dirty)
mv tracking/version.go tracking/version.go.dev
echo "package tracking" > tracking/version.go
echo "var version=\"$1-$GITVER\"" >> tracking/version.go

# v2 build - production
~/go/bin/wails build -clean

# Create production zip
cd build/bin
zip ../../vertcoin-ocm-$1-linux-x64.zip ./vertcoin-ocm
cd ../..

# v2 build - debug
~/go/bin/wails build -clean -debug

# Create debug zip
cd build/bin
zip ../../vertcoin-ocm-$1-linux-x64-debug.zip ./vertcoin-ocm
cd ../..

rm tracking/version.go
mv tracking/version.go.dev tracking/version.go