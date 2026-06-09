@echo off
setlocal
title BobTrader - UltraTrader Go
cd /d "%~dp0"

echo [BobTrader] Building UltraTrader Go...
cd ultratrader-go
go build -o ultratrader.exe ./cmd/ultratrader
if errorlevel 1 (
    echo [BobTrader] Build failed.
    pause
    exit /b 1
)
echo [BobTrader] Build complete.
endlocal
