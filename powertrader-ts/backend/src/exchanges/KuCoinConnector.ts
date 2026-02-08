import { IExchangeConnector } from "../../engine/connector/IExchangeConnector";
import axios from 'axios';

export class KuCoinConnector implements IExchangeConnector {
    name = "KuCoin";
    private baseUrl = "https://api.kucoin.com";

    async fetchTicker(pair: string): Promise<number> {
        try {
            // KuCoin symbol format: BTC-USDT
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
            // Map interval strings to KuCoin types
            // 1min, 3min, 5min, 15min, 30min, 1hour, 2hour, 4hour, 6hour, 8hour, 12hour, 1day, 1week
            let type = interval.replace('h', 'hour').replace('m', 'min').replace('d', 'day');
            if (type === '1h') type = '1hour';

            // Calculate startAt based on limit to save bandwidth (optional, but good practice)
            // For now, simple latest fetch

            const res = await axios.get(`${this.baseUrl}/api/v1/market/candles`, {
                params: {
                    symbol: symbol,
                    type: type
                }
            });

            if (res.data && res.data.data) {
                // KuCoin: [time, open, close, high, low, volume, turnover]
                // We want oldest to newest, but KuCoin returns newest first. Reverse it.
                const raw = res.data.data.slice(0, limit).reverse();

                return raw.map((c: string[]) => ({
                    timestamp: parseInt(c[0]),
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

    async fetchBalance(): Promise<any> { return {}; }
    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> { return {}; }
    async cancelOrder(id: string, pair: string): Promise<boolean> { return true; }
    async fetchOrder(id: string, pair: string): Promise<any> { return {}; }
    async fetchOpenOrders(pair?: string): Promise<any[]> { return []; }
}
