@echo off
setlocal
title BobTrader - UltraTrader Go
cd /d "%~dp0"

echo [BobTrader] Starting UltraTrader Go...
cd ultratrader-go

if "%1"=="" (
    echo [BobTrader] Usage: start.bat ^<config^>
    echo [BobTrader]   e.g. start.bat config\autonomous-paper.json
    echo [BobTrader] Using default: config\autonomous-paper.json
    go run ./cmd/ultratrader --config config\autonomous-paper.json
) else (
    go run ./cmd/ultratrader --config %1
)

if errorlevel 1 (
    echo [BobTrader] Exited with error code %errorlevel%.
    pause
)
endlocal
