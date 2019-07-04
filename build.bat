@ECHO OFF
SET ver=%1
git describe --always --long --dirty > git-version
SET /p gitver=<git-version
DEL git-version
CD tracking
REN version.go version.go.build
ECHO package tracking >> version.go
ECHO var version="%ver%-%gitver%" >> version.go
CD ..
DEL vertcoin-ocm.exe
wails build
REN vertcoin-ocm vertcoin-ocm.exe
7z a vertcoin-ocm-%ver%-windows-x64.zip vertcoin-ocm.exe 
CD tracking
DEL version.go
REN version.go.build version.go
CD ..