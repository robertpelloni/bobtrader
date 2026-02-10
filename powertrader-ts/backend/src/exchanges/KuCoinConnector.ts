import { IExchangeConnector } from "../engine/connector/IExchangeConnector";
import axios from 'axios';
import crypto from 'crypto-js';
import { v4 as uuidv4 } from 'uuid';

export class KuCoinConnector implements IExchangeConnector {
    name = "KuCoin";
    private baseUrl = "https://api.kucoin.com";
    private apiKey: string;
    private apiSecret: string;
    private apiPassphrase: string;

    constructor(apiKey: string = "", apiSecret: string = "", apiPassphrase: string = "") {
        this.apiKey = apiKey;
        this.apiSecret = apiSecret;
        this.apiPassphrase = apiPassphrase;
    }

    private sign(endpoint: string, method: string = 'GET', body: any = ''): any {
        const timestamp = Date.now();
        const bodyStr = typeof body === 'object' ? JSON.stringify(body) : body;
        const what = timestamp + method + endpoint + bodyStr;

        const signature = crypto.HmacSHA256(what, this.apiSecret).toString(crypto.enc.Base64);
        const passphrase = crypto.HmacSHA256(this.apiPassphrase, this.apiSecret).toString(crypto.enc.Base64);

        return {
            'KC-API-KEY': this.apiKey,
            'KC-API-SIGN': signature,
            'KC-API-TIMESTAMP': timestamp,
            'KC-API-PASSPHRASE': passphrase,
            'KC-API-KEY-VERSION': '2',
            'Content-Type': 'application/json'
        };
    }

    async fetchTicker(pair: string): Promise<number> {
        try {
            const symbol = pair.replace("USD", "USDT");
            const res = await axios.get(`${this.baseUrl}/api/v1/market/orderbook/level1?symbol=${symbol}`);
            return parseFloat(res.data.data.price);
        } catch (e) {
            console.error(`[KuCoin] Error fetching ticker for ${pair}:`, e);
            return 0;
        }
    }

    async fetchOrderBook(pair: string): Promise<any> { return {}; }

    async fetchOHLCV(pair: string, interval: string, limit: number = 100): Promise<any[]> {
        try {
            const symbol = pair.replace("USD", "USDT");
            let type = interval.replace('h', 'hour').replace('m', 'min').replace('d', 'day');
            if (type === '1h') type = '1hour';

            const res = await axios.get(`${this.baseUrl}/api/v1/market/candles`, {
                params: {
                    symbol: symbol,
                    type: type
                }
            });

            if (res.data && res.data.data) {
                const raw = res.data.data.slice(0, limit).reverse();
                return raw.map((c: string[]) => ({
                    timestamp: parseInt(c[0]) * 1000, // KuCoin uses seconds for start/end but this endpoint seems to return seconds? No, usually seconds. Wait, the previous impl had parseInt(c[0]). If backtest engine expects ms, we should mult by 1000 if needed.
                    // Actually previous implementation didn't mult by 1000. Let's fix that too. Standard is MS.
                    // KuCoin docs: startAt long (seconds)
                    // The returned candles are [time, open, close, high, low, volume, turnover]
                    // The returned time is in seconds.
                    open: parseFloat(c[1]),
                    close: parseFloat(c[2]),
                    high: parseFloat(c[3]),
                    low: parseFloat(c[4]),
                    volume: parseFloat(c[5])
                }));
            }
            return [];
        } catch (e) {
            console.error(`[KuCoin] Error fetching OHLCV for ${pair}:`, e);
            return [];
        }
    }

    async fetchBalance(): Promise<any> {
        if (!this.apiKey) return {};
        try {
            const endpoint = '/api/v1/accounts';
            const headers = this.sign(endpoint, 'GET');
            const res = await axios.get(`${this.baseUrl}${endpoint}`, { headers });

            const balances: any = {};
            if (res.data.code === '200000') {
                for (const acc of res.data.data) {
                    if (acc.type === 'trade') {
                        balances[acc.currency] = parseFloat(acc.available);
                    }
                }
            }
            return balances;
        } catch (e: any) {
            console.error(`[KuCoin] Error fetching balance: ${e.message}`);
            return {};
        }
    }

    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> {
        if (!this.apiKey) throw new Error("API Keys not configured");
        try {
            const endpoint = '/api/v1/orders';
            const symbol = pair.replace("USD", "USDT");

            // KuCoin uses clientOid for idempotency
            const body: any = {
                clientOid: uuidv4(),
                side: side,
                symbol: symbol,
                type: type
            };

            if (type === 'limit') {
                if (!price) throw new Error("Limit order requires price");
                body.price = price.toString();
                body.size = amount.toString(); // Amount in base currency
            } else {
                // Market order
                if (side === 'buy') {
                    // For market buy, KuCoin usually requires 'funds' (amount in quote currency)
                    // But if we want to buy a specific amount of base currency, we might need 'size'?
                    // KuCoin API: market order parameters: size (amount in base), funds (amount in quote)
                    // We'll assume 'amount' is in base currency (e.g. BTC)
                    body.size = amount.toString();
                } else {
                    body.size = amount.toString();
                }
            }

            const headers = this.sign(endpoint, 'POST', body);
            const res = await axios.post(`${this.baseUrl}${endpoint}`, body, { headers });

            if (res.data.code === '200000') {
                return {
                    id: res.data.data.orderId,
                    pair,
                    status: 'open',
                    ...body
                };
            } else {
                throw new Error(res.data.msg);
            }
        } catch (e: any) {
            console.error(`[KuCoin] Error creating order: ${e.message}`);
            throw e;
        }
    }

    async cancelOrder(id: string, pair: string): Promise<boolean> {
        if (!this.apiKey) return false;
        try {
            const endpoint = `/api/v1/orders/${id}`;
            const headers = this.sign(endpoint, 'DELETE');
            const res = await axios.delete(`${this.baseUrl}${endpoint}`, { headers });
            return res.data.code === '200000';
        } catch (e) {
            return false;
        }
    }

    async fetchOrder(id: string, pair: string): Promise<any> {
        if (!this.apiKey) return null;
        try {
            const endpoint = `/api/v1/orders/${id}`;
            const headers = this.sign(endpoint, 'GET');
            const res = await axios.get(`${this.baseUrl}${endpoint}`, { headers });
            return res.data.data;
        } catch (e) {
            return null;
        }
    }

    async fetchOpenOrders(pair?: string): Promise<any[]> {
        if (!this.apiKey) return [];
        try {
            const endpoint = `/api/v1/orders?status=active${pair ? `&symbol=${pair.replace('USD', 'USDT')}` : ''}`;
            const headers = this.sign(endpoint, 'GET');
            const res = await axios.get(`${this.baseUrl}${endpoint}`, { headers });
            return res.data.data.items || [];
        } catch (e) {
            return [];
        }
    }
}
