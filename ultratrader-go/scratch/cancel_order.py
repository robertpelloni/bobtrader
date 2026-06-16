import json
import hmac
import hashlib
import time
import requests
import sys

def cancel_order(symbol, order_id):
    # Load secrets
    with open("config/secrets/binance-production.json", "r") as f:
        secrets = json.load(f)
    
    api_key = secrets["api_key"]
    secret_key = secrets["secret_key"]
    base_url = "https://api.binance.us"

    params = {
        "symbol": symbol,
        "orderId": order_id,
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
    
    url = f"{base_url}/api/v3/order?{query_string}"
    resp = requests.delete(url, headers=headers)
    print("Status Code:", resp.status_code)
    print("Response:")
    print(json.dumps(resp.json(), indent=2))

if __name__ == "__main__":
    if len(sys.argv) < 3:
        # Default to the ETHUSDT order we found
        cancel_order("ETHUSDT", 1458324154)
    else:
        cancel_order(sys.argv[1], int(sys.argv[2]))
