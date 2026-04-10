@echo off
echo Building bobtrader...
pip install -r requirements.txt
python -m py_compile pt_hub.py pt_trader.py
echo Build complete.
pause