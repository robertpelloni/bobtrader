import json
import hmac
import hashlib
import time
import requests

def get_open_orders():
    # Load secrets
    with open("config/secrets/binance-production.json", "r") as f:
        secrets = json.load(f)
    
    api_key = secrets["api_key"]
    secret_key = secrets["secret_key"]
    base_url = "https://api.binance.us" # or production

    params = {
        "timestamp": int(time.time() * 1000),
        "recvWindow": 60000
    }
    
    # Sign request
    query_string = "&".join([f"{k}={v}" for k, v in params.items()])
    signature = hmac.new(secret_key.encode('utf-8'), query_string.encode('utf-8'), hashlib.sha256).hexdigest()
    query_string += f"&signature={signature}"
    
    headers = {
        "X-MBX-APIKEY": api_key
    }
    
    url = f"{base_url}/api/v3/openOrders?{query_string}"
    resp = requests.get(url, headers=headers)
    print("Status Code:", resp.status_code)
    print("Response:")
    print(json.dumps(resp.json(), indent=2))

if __name__ == "__main__":
    get_open_orders()
