import { IExchangeConnector } from "../../engine/connector/IExchangeConnector";
import axios from 'axios';
import crypto from 'crypto-js';

export class BinanceConnector implements IExchangeConnector {
    name = "Binance";
    private baseUrl = "https://api.binance.us"; // Using US endpoint
    private apiKey: string;
    private apiSecret: string;

    constructor(apiKey: string = "", apiSecret: string = "") {
        this.apiKey = apiKey;
        this.apiSecret = apiSecret;
    }

    private sign(params: any): string {
        const query = Object.keys(params)
            .map(key => `${key}=${params[key]}`)
            .join('&');
        return crypto.HmacSHA256(query, this.apiSecret).toString(crypto.enc.Hex);
    }

    async fetchTicker(pair: string): Promise<number> {
        try {
            const symbol = pair.replace("-", "");
            const res = await axios.get(`${this.baseUrl}/api/v3/ticker/price?symbol=${symbol}`);
            return parseFloat(res.data.price);
        } catch (e) {
            console.error(`[Binance] Error fetching ticker for ${pair}:`, e);
            return 0;
        }
    }

    async fetchOrderBook(pair: string): Promise<any> { return {}; }

    async fetchOHLCV(pair: string, interval: string, limit: number = 100): Promise<any[]> {
        try {
            const symbol = pair.replace("-", "");
            let klineInterval = interval;
            if (interval === '1h') klineInterval = '1h';
            if (interval === '1d') klineInterval = '1d';

            const res = await axios.get(`${this.baseUrl}/api/v3/klines`, {
                params: {
                    symbol: symbol,
                    interval: klineInterval,
                    limit: limit
                }
            });

            if (Array.isArray(res.data)) {
                return res.data.map((c: any[]) => ({
                    timestamp: c[0],
                    open: parseFloat(c[1]),
                    high: parseFloat(c[2]),
                    low: parseFloat(c[3]),
                    close: parseFloat(c[4]),
                    volume: parseFloat(c[5])
                }));
            }
            return [];
        } catch (e) {
            console.error(`[Binance] Error fetching OHLCV for ${pair}:`, e);
            return [];
        }
    }

    async fetchBalance(): Promise<any> {
        if (!this.apiKey) return {};
        try {
            const endpoint = '/api/v3/account';
            const params: any = { timestamp: Date.now() };
            params.signature = this.sign(params);

            const res = await axios.get(`${this.baseUrl}${endpoint}`, {
                headers: { 'X-MBX-APIKEY': this.apiKey },
                params: params
            });

            const balances: any = {};
            for (const asset of res.data.balances) {
                const free = parseFloat(asset.free);
                if (free > 0) {
                    balances[asset.asset] = free;
                }
            }
            return balances;
        } catch (e: any) {
            console.error(`[Binance] Error fetching balance: ${e.message}`);
            return {};
        }
    }

    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> {
        if (!this.apiKey) throw new Error("API Keys not configured");
        try {
            const endpoint = '/api/v3/order';
            const symbol = pair.replace("-", "");

            const params: any = {
                symbol: symbol,
                side: side.toUpperCase(),
                type: type.toUpperCase(),
                quantity: amount,
                timestamp: Date.now()
            };

            if (type === 'limit') {
                if (!price) throw new Error("Limit order requires price");
                params.price = price;
                params.timeInForce = 'GTC';
            }

            params.signature = this.sign(params);

            // Binance requires POST with query params (or form data), not JSON body for V3 usually,
            // but Axios with params serializer handles it if we use 'params' option for GET?
            // For POST, it's usually query string in body or url.
            // Let's attach to URL string to be safe.
            const query = Object.keys(params).map(k => `${k}=${params[k]}`).join('&');

            const res = await axios.post(`${this.baseUrl}${endpoint}?${query}`, null, {
                headers: { 'X-MBX-APIKEY': this.apiKey }
            });

            return {
                id: res.data.orderId,
                pair,
                status: res.data.status.toLowerCase(),
                price: res.data.price,
                amount: res.data.origQty
            };
        } catch (e: any) {
            console.error(`[Binance] Error creating order: ${e.message}`);
            throw e;
        }
    }

    async cancelOrder(id: string, pair: string): Promise<boolean> {
        if (!this.apiKey) return false;
        try {
            const endpoint = '/api/v3/order';
            const symbol = pair.replace("-", "");
            const params: any = {
                symbol: symbol,
                orderId: id,
                timestamp: Date.now()
            };
            params.signature = this.sign(params);

            // DELETE usually takes params in query string
            const res = await axios.delete(`${this.baseUrl}${endpoint}`, {
                headers: { 'X-MBX-APIKEY': this.apiKey },
                params: params
            });
            return true;
        } catch (e) {
            return false;
        }
    }

    async fetchOrder(id: string, pair: string): Promise<any> {
        if (!this.apiKey) return null;
        try {
            const endpoint = '/api/v3/order';
            const symbol = pair.replace("-", "");
            const params: any = {
                symbol: symbol,
                orderId: id,
                timestamp: Date.now()
            };
            params.signature = this.sign(params);

            const res = await axios.get(`${this.baseUrl}${endpoint}`, {
                headers: { 'X-MBX-APIKEY': this.apiKey },
                params: params
            });
            return res.data;
        } catch (e) {
            return null;
        }
    }

    async fetchOpenOrders(pair?: string): Promise<any[]> {
        if (!this.apiKey) return [];
        try {
            const endpoint = '/api/v3/openOrders';
            const params: any = { timestamp: Date.now() };
            if (pair) params.symbol = pair.replace("-", "");

            params.signature = this.sign(params);

            const res = await axios.get(`${this.baseUrl}${endpoint}`, {
                headers: { 'X-MBX-APIKEY': this.apiKey },
                params: params
            });
            return res.data;
        } catch (e) {
            return [];
        }
    }
}
