@ECHO OFF
IF NOT "%~1"=="" GOTO :BUILD

:USAGE
ECHO Usage: %~nx0 version
GOTO :EOF

:BUILD
SET ver=%1
git describe --always --long --dirty > %TEMP%\git-version
SET /p gitver=<%TEMP%\git-version
DEL %TEMP%\git-version >nul 2>&1
CD tracking
REN version.go version.go.build
ECHO package tracking >> version.go
ECHO var version="%ver%-%gitver%" >> version.go
CD ..
DEL build\vertcoin-ocm.exe >nul 2>&1
wails build
ECHO "Sign the release assembly now on the windows machine if desired, then:"
PAUSE
CD build 
7z -sdel -aou a vertcoin-ocm-%ver%-windows-x64.zip vertcoin-ocm.exe
CD ..
wails build -d
ECHO "Sign the debug assembly now on the windows machine if desired, then:"
PAUSE
CD build
7z -sdel -aou a vertcoin-ocm-%ver%-windows-x64-debug.zip vertcoin-ocm.exe
CD ..
DEL *.syso *.manifest *.ico *.rc *.exe >nul 2>&1
CD tracking
DEL version.go >nul 2>&1
REN version.go.build version.go
CD ..