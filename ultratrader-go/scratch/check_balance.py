import json
import hmac
import hashlib
import time
import requests

def check_balance():
    with open("config/secrets/binance-production.json", "r") as f:
        secrets = json.load(f)
    
    api_key = secrets["api_key"]
    secret_key = secrets["secret_key"]
    base_url = "https://api.binance.us"

    params = {
        "timestamp": int(time.time() * 1000),
        "recvWindow": 60000
    }
    
    query_string = "&".join([f"{k}={v}" for k, v in params.items()])
    signature = hmac.new(secret_key.encode('utf-8'), query_string.encode('utf-8'), hashlib.sha256).hexdigest()
    query_string += f"&signature={signature}"
    
    headers = {
        "X-MBX-APIKEY": api_key
    }
    
    url = f"{base_url}/api/v3/account?{query_string}"
    resp = requests.get(url, headers=headers)
    if resp.status_code != 200:
        print("Error checking balance:", resp.text)
        return
        
    data = resp.json()
    balances = data.get("balances", [])
    print("Non-zero Balances:")
    for bal in balances:
        free = float(bal["free"])
        locked = float(bal["locked"])
        if free > 0 or locked > 0:
            print(f"Asset: {bal['asset']} | Free: {free:.8f} | Locked: {locked:.8f} | Total: {free+locked:.8f}")

if __name__ == "__main__":
    check_balance()
