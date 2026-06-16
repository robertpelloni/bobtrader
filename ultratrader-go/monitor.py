import sys
import json
import time
import hmac
import hashlib
import urllib.request
from datetime import datetime

SECRETS_FILE = "config/secrets/binance-production.json"
CHECK_INTERVAL = 600  # 10 minutes


def load_secrets():
    with open(SECRETS_FILE) as f:
        return json.load(f)


def binance_request(endpoint, params="", secret_key=""):
    secrets = load_secrets()
    api_key = secrets["api_key"]
    timestamp = int(time.time() * 1000)
    full_params = (
        f"{params}&timestamp={timestamp}&recvWindow=5000"
        if params
        else f"timestamp={timestamp}&recvWindow=5000"
    )
    signature = hmac.new(
        secret_key.encode(), full_params.encode(), hashlib.sha256
    ).hexdigest()
    url = f"https://api.binance.us/{endpoint}?{full_params}&signature={signature}"
    req = urllib.request.Request(url, headers={"X-MBX-APIKEY": api_key})
    resp = urllib.request.urlopen(req, timeout=10)
    return json.loads(resp.read())


def get_price(symbol):
    url = f"https://api.binance.us/api/v3/ticker/price?symbol={symbol}"
    resp = urllib.request.urlopen(url, timeout=10)
    return float(json.loads(resp.read())["price"])


def get_fear_greed():
    try:
        url = "https://api.alternative.me/fng/?limit=1"
        resp = urllib.request.urlopen(url, timeout=10)
        data = json.loads(resp.read())
        return int(data["data"][0]["value"]), data["data"][0]["value_classification"]
    except:
        return None, None


def check_orders(secret_key):
    data = binance_request("api/v3/openOrders", "symbol=ETHUSDT", secret_key)
    return data


def check_balance(secret_key):
    data = binance_request("api/v3/account", "", secret_key)
    eth = next((b for b in data["balances"] if b["asset"] == "ETH"), None)
    usdt = next((b for b in data["balances"] if b["asset"] == "USDT"), None)
    return eth, usdt


def main():
    secrets = load_secrets()
    secret_key = secrets["secret_key"]

    print("=== ETH TRADING MONITOR ===")
    print(f"Started: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"Check interval: {CHECK_INTERVAL}s")
    print()

    last_price = None
    while True:
        try:
            eth_price = get_price("ETHUSDT")
            fng_value, fng_class = get_fear_greed()
            orders = check_orders(secret_key)
            eth_bal, usdt_bal = check_balance(secret_key)

            print(
                f"\n[{datetime.now().strftime('%H:%M:%S')}] ETH: ${eth_price:.2f}",
                end="",
            )
            if fng_value:
                print(f" | F&G: {fng_value} ({fng_class})", end="")

            # Check triggers
            alerts = []
            if eth_price <= 1630:
                alerts.append("⚠️ DIP BUY ZONE - Order should fill soon")
            if eth_price >= 1720:
                alerts.append("🎯 TAKE PROFIT ZONE - Manual sell recommended")
            if eth_price <= 1620:
                alerts.append("🚨 STOP LOSS BREACH - Sell immediately!")
            if fng_value and fng_value > 50:
                alerts.append("📈 Sentiment shifted to Greed")
            if fng_value and fng_value < 10:
                alerts.append("💀 Extreme Panic - Potential opportunity")

            if alerts:
                print(f"\n{'!' * 60}")
                for alert in alerts:
                    print(f"  {alert}")
                print(f"{'!' * 60}")

            # Status
            if eth_bal:
                eth_free = float(eth_bal["free"])
                eth_locked = float(eth_bal["locked"])
                print(f" | ETH: {eth_free:.4f} (locked: {eth_locked:.4f})", end="")
            if usdt_bal:
                usdt_free = float(usdt_bal["free"])
                usdt_locked = float(usdt_bal["locked"])
                print(f" | USDT: ${usdt_free:.2f} (locked: ${usdt_locked:.2f})", end="")

            if orders:
                print(f"\n  Open orders: {len(orders)}")
                for o in orders:
                    print(
                        f"    {o['side']} {o['origQty']} @ ${o['price']} ({o['type']})"
                    )

            sys.stdout.flush()
            last_price = eth_price
            time.sleep(CHECK_INTERVAL)

        except KeyboardInterrupt:
            print("\n\nMonitor stopped by user")
            break
        except Exception as e:
            print(f"\n[ERROR] {e}")
            time.sleep(60)


if __name__ == "__main__":
    main()
