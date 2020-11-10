@ECHO OFF
SET ver=%1
git describe --always --long --dirty > %TEMP%\git-version
SET /p gitver=<%TEMP%\git-version
DEL %TEMP%\git-version
CD tracking
REN version.go version.go.build
ECHO package tracking >> version.go
ECHO var version="%ver%-%gitver%" >> version.go
CD ..
DEL vertcoin-ocm.exe
wails build
CD build 
7z a ../vertcoin-ocm-%ver%-windows-x64.zip vertcoin-ocm.exe 
CD ../tracking
DEL version.go
REN version.go.build version.go
CD ..