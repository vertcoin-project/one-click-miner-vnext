#!/bin/bash
GITVER=$(git describe --always --long --dirty)
mv tracking/version.go tracking/version.go.dev
echo "package tracking" > tracking/version.go
echo "var version=\"$1-$GITVER\"" >> tracking/version.go
wails build
cd build
zip ../verthash-ocm-$1-linux-x64.zip ./verthash-ocm
cd ..
wails build -d 
cd build
zip ../verthash-ocm-$1-linux-x64-debug.zip ./verthash-ocm
cd ..
rm tracking/version.go
mv tracking/version.go.dev tracking/version.go