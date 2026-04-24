@echo off
setlocal
title BobTrader
cd /d "%~dp0"

echo [BobTrader] Starting...
python --version >nul 2>nul
if errorlevel 1 (
    echo [BobTrader] python not found. Please install it.
    pause
    exit /b 1
)

python -m bobtrader

if errorlevel 1 (
    echo [BobTrader] Exited with error code %errorlevel%.
    pause
)
endlocal
