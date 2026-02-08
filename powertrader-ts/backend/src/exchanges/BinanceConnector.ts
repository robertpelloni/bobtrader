import { IExchangeConnector } from "../../engine/connector/IExchangeConnector";
import axios from 'axios';

export class BinanceConnector implements IExchangeConnector {
    name = "Binance";
    private baseUrl = "https://api.binance.us"; // Using US endpoint

    async fetchTicker(pair: string): Promise<number> {
        try {
            // Binance symbol format: BTCUSD
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
            // Map interval
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
                // Binance: [Open time, Open, High, Low, Close, Volume, ...]
                // Returns oldest first.
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

    async fetchBalance(): Promise<any> { return {}; }
    async createOrder(pair: string, type: 'market'|'limit', side: 'buy'|'sell', amount: number, price?: number): Promise<any> { return {}; }
    async cancelOrder(id: string, pair: string): Promise<boolean> { return true; }
    async fetchOrder(id: string, pair: string): Promise<any> { return {}; }
    async fetchOpenOrders(pair?: string): Promise<any[]> { return []; }
}
