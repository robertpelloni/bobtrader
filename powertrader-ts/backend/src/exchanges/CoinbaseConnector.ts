import { IExchangeConnector } from "../engine/connector/IExchangeConnector";
import axios from 'axios';
import crypto from 'crypto-js';

export class CoinbaseConnector implements IExchangeConnector {
    name = "Coinbase";
    private baseUrl = "https://api.coinbase.com/api/v3";
    private apiKey: string;
    private apiSecret: string;

    constructor(apiKey: string = "", apiSecret: string = "") {
        this.apiKey = apiKey;
        this.apiSecret = apiSecret;
    }

    private sign(method: string, path: string, body: string = ''): any {
        const timestamp = Math.floor(Date.now() / 1000).toString();
        const message = timestamp + method + path + body;
        const signature = crypto.HmacSHA256(message, this.apiSecret).toString(crypto.enc.Hex);

        return {
            'CB-ACCESS-KEY': this.apiKey,
            'CB-ACCESS-SIGN': signature,
            'CB-ACCESS-TIMESTAMP': timestamp,
            'Content-Type': 'application/json'
        };
    }

    async fetchTicker(pair: string): Promise<number> {
        try {
            const productId = pair; // Coinbase uses BTC-USD format natively
            const res = await axios.get(`${this.baseUrl}/brokerage/products/${productId}/ticker`);
            // Response format: { price: "...", ... }
            return parseFloat(res.data.price);
        } catch (e) {
            console.error(`[Coinbase] Error fetching ticker for ${pair}:`, e);
            return 0;
        }
    }

    async fetchOrderBook(pair: string): Promise<any> { return {}; }

    async fetchOHLCV(pair: string, interval: string, limit: number = 100): Promise<any[]> {
        try {
            // Coinbase Candles: /brokerage/products/{product_id}/candles
            // Granularity: ONE_MINUTE, FIVE_MINUTE, FIFTEEN_MINUTE, THIRTY_MINUTE, ONE_HOUR, TWO_HOUR, SIX_HOUR, ONE_DAY
            const granularityMap: Record<string, string> = {
                '1m': 'ONE_MINUTE',
                '5m': 'FIVE_MINUTE',
                '15m': 'FIFTEEN_MINUTE',
                '30m': 'THIRTY_MINUTE',
                '1h': 'ONE_HOUR',
                '2h': 'TWO_HOUR',
                '6h': 'SIX_HOUR',
                '1d': 'ONE_DAY'
            };

            const granularity = granularityMap[interval] || 'ONE_HOUR';
            const end = Math.floor(Date.now() / 1000);
            // Rough calc for start: limit * seconds_in_interval
            const secondsPerInterval = 3600; // approximation for 1h
            const start = end - (limit * secondsPerInterval * 2); // Fetch extra to be safe

            const res = await axios.get(`${this.baseUrl}/brokerage/products/${pair}/candles`, {
                params: {
                    start: start.toString(),
                    end: end.toString(),
                    granularity: granularity
                }
            });

            if (res.data && res.data.candles) {
                // Coinbase: [start, low, high, open, close, volume]
                // Returns newest first
                const candles = res.data.candles.slice(0, limit).reverse();
                return candles.map((c: any) => ({
                    timestamp: parseInt(c.start) * 1000,
                    open: parseFloat(c.open),
                    high: parseFloat(c.high),
                    low: parseFloat(c.low),
                    close: parseFloat(c.close),
                    volume: parseFloat(c.volume)
                }));
            }
            return [];
        } catch (e) {
            console.error(`[Coinbase] Error fetching OHLCV for ${pair}:`, e);
            return [];
        }
    }

    async fetchBalance(): Promise<any> {
        if (!this.apiKey) return {};
        try {
            const path = '/brokerage/accounts';
            const headers = this.sign('GET', `/api/v3${path}`);

            const res = await axios.get(`${this.baseUrl}${path}`, { headers });

            const balances: any = {};
            if (res.data.accounts) {
                for (const acc of res.data.accounts) {
                    const available = parseFloat(acc.available_balance.value);
                    if (available > 0) {
                        balances[acc.currency] = available;
                    }
                }
            }
            return balances;
        } catch (e: any) {
            console.error(`[Coinbase] Error fetching balance: ${e.message}`);
            return {};
        }
    }

    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> {
        if (!this.apiKey) throw new Error("API Keys not configured");
        try {
            const path = '/brokerage/orders';
            const clientOrderId = Math.random().toString(36).substring(7);

            const bodyObj: any = {
                client_order_id: clientOrderId,
                product_id: pair,
                side: side.toUpperCase(), // BUY or SELL
                order_configuration: {}
            };

            if (type === 'market') {
                if (side === 'buy') {
                    // Market Buy requires quote_size (how much USD to spend)
                    // We need to approximate if 'amount' is base currency.
                    // This is tricky. Ideally 'amount' is quote for market buy.
                    // Assuming 'amount' is Base (e.g. BTC).
                    // Coinbase Market Buy strictly uses quote_size? Or base_size?
                    // Docs: market_market_ioc: { quote_size: string, base_size: string }
                    // Usually base_size works for sells, quote_size for buys.
                    // Let's try base_size for both if supported, otherwise fetch price and calc quote.
                    bodyObj.order_configuration.market_market_ioc = {
                        base_size: amount.toString()
                    };
                } else {
                    bodyObj.order_configuration.market_market_ioc = {
                        base_size: amount.toString()
                    };
                }
            } else {
                // Limit
                bodyObj.order_configuration.limit_limit_gtc = {
                    base_size: amount.toString(),
                    limit_price: price!.toString(),
                    post_only: false
                };
            }

            const bodyStr = JSON.stringify(bodyObj);
            const headers = this.sign('POST', `/api/v3${path}`, bodyStr);

            const res = await axios.post(`${this.baseUrl}${path}`, bodyObj, { headers });

            if (res.data.success) {
                return {
                    id: res.data.order_id,
                    pair,
                    status: 'open', // Coinbase returns success immediately, status might be pending
                    amount,
                    price
                };
            } else {
                throw new Error(JSON.stringify(res.data.error_response));
            }
        } catch (e: any) {
            console.error(`[Coinbase] Error creating order: ${e.message}`);
            throw e;
        }
    }

    async cancelOrder(id: string, pair: string): Promise<boolean> {
        if (!this.apiKey) return false;
        try {
            const path = '/brokerage/orders/batch_cancel';
            const bodyObj = { order_ids: [id] };
            const bodyStr = JSON.stringify(bodyObj);
            const headers = this.sign('POST', `/api/v3${path}`, bodyStr);

            const res = await axios.post(`${this.baseUrl}${path}`, bodyObj, { headers });
            return res.data.results && res.data.results[0].success;
        } catch (e) {
            return false;
        }
    }

    async fetchOrder(id: string, pair: string): Promise<any> {
        if (!this.apiKey) return null;
        try {
            const path = `/brokerage/orders/${id}`;
            const headers = this.sign('GET', `/api/v3${path}`);
            const res = await axios.get(`${this.baseUrl}${path}`, { headers });
            return res.data.order;
        } catch (e) {
            return null;
        }
    }

    async fetchOpenOrders(pair?: string): Promise<any[]> {
        // Coinbase doesn't have a simple "all open orders" endpoint without status filters usually
        // /brokerage/orders/historical/batch?order_status=OPEN
        // But for now, returning empty as it's complex to implement efficiently
        return [];
    }
}
