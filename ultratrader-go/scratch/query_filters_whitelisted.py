import urllib.request
import json
import sys

sys.stdout.reconfigure(encoding='utf-8')

targets = ["SOLUSDT", "DOGEUSDT", "PEPEUSDT", "WIFUSDT", "XRPUSDT"]
# Format JSON without spaces: ["SOLUSDT","DOGEUSDT",...]
symbols_param = json.dumps(targets, separators=(',', ':'))
# URL encode the query parameters
import urllib.parse
url = 'https://api.binance.us/api/v3/exchangeInfo?symbols=' + urllib.parse.quote(symbols_param)

print(f"Querying URL: {url}")
try:
    req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
    resp = urllib.request.urlopen(req)
    data = json.loads(resp.read())
    
    with open("scratch/exchange_info_output.json", "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2)
    print("Saved raw JSON to scratch/exchange_info_output.json")
    
    found_symbols = []
    for s in data.get('symbols', []):
        symbol = s['symbol']
        found_symbols.append(symbol)
        print(f"Symbol: {symbol} - Status: {s['status']}")
        for filt in s.get('filters', []):
            ftype = filt['filterType']
            if ftype == 'LOT_SIZE':
                print(f"  LOT_SIZE: minQty={filt['minQty']}, maxQty={filt['maxQty']}, stepSize={filt['stepSize']}")
            elif ftype == 'NOTIONAL':
                print(f"  NOTIONAL: minNotional={filt.get('minNotional')}, applyToMarket={filt.get('applyToMarket')}, avgPriceMins={filt.get('avgPriceMins')}")
            elif ftype == 'PRICE_FILTER':
                print(f"  PRICE_FILTER: minPrice={filt['minPrice']}, maxPrice={filt['maxPrice']}, tickSize={filt['tickSize']}")
            elif ftype == 'MIN_NOTIONAL':
                print(f"  MIN_NOTIONAL: minNotional={filt.get('minNotional')}")
                
    for t in targets:
        if t not in found_symbols:
            print(f"Symbol {t} was NOT returned by Binance.US!")
            
except Exception as e:
    print('Error:', e)
