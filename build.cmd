@echo off
setlocal
cd /d "%~dp0"
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go test -count=1 -v ./...
if errorlevel 1 exit /b 1
go vet ./...
if errorlevel 1 exit /b 1
go build -trimpath -ldflags="-H=windowsgui -s -w" -o vpn-monitor.exe .
if errorlevel 1 exit /b 1
echo Built vpn-monitor.exe
