import requests
import json

def check():
    url = "https://api.binance.us/api/v3/exchangeInfo"
    resp = requests.get(url)
    data = resp.json()
    
    symbols = [s["symbol"] for s in data["symbols"]]
    
    targets = ["SOLUSDT", "DOGEUSDT", "PEPEUSDT", "WIFUSDT", "XRPUSDT"]
    for t in targets:
        if t in symbols:
            print(f"✅ {t} is supported on Binance.US")
        else:
            print(f"❌ {t} is NOT supported on Binance.US")

if __name__ == "__main__":
    check()
