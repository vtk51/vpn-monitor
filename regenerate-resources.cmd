@echo off
setlocal
cd /d "%~dp0"
rem app.ico is already included. Python/Pillow is optional for redesign only.
rem Resource compiler pinned to the verified upstream v0.10.2 release.
go run github.com/akavel/rsrc@v0.10.2 -arch amd64 -ico app.ico -o resource_windows_amd64.syso
if errorlevel 1 exit /b 1
echo Embedded resource object regenerated. Run build.cmd.
