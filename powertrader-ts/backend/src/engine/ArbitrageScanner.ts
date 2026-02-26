import { ConfigManager } from "../config/ConfigManager";
import { IExchangeConnector } from "./connector/IExchangeConnector";
import { BinanceConnector } from "../exchanges/BinanceConnector";
import { KuCoinConnector } from "../exchanges/KuCoinConnector";
import { CoinbaseConnector } from "../exchanges/CoinbaseConnector";
import { RobinhoodConnector } from "../exchanges/RobinhoodConnector";

interface Opportunity {
    symbol: string;
    buyExchange: string;
    buyPrice: number;
    sellExchange: string;
    sellPrice: number;
    spread: number;
    spreadPct: number;
    timestamp: number;
}

export class ArbitrageScanner {
    private exchanges: Map<string, IExchangeConnector> = new Map();
    private config: ConfigManager;

    constructor() {
        this.config = ConfigManager.getInstance();
        this.initializeExchanges();
    }

    private initializeExchanges() {
        // Instantiate connectors. In a real scenario, use keys from config.
        // For scanning, we only need public data (tickers), so empty keys might suffice for some.
        // Or we assume the primary connector keys are used.
        this.exchanges.set("Binance", new BinanceConnector("",""));
        this.exchanges.set("KuCoin", new KuCoinConnector("","",""));
        this.exchanges.set("Coinbase", new CoinbaseConnector("",""));
        // this.exchanges.set("Robinhood", new RobinhoodConnector("","")); // RH is harder to get public quotes without login sometimes
    }

    public async scan(symbol: string): Promise<Opportunity | null> {
        // Use a standard array instead of mutating outside
        const tasks = Array.from(this.exchanges.entries()).map(async ([name, connector]) => {
            try {
                // Try USD first, then USDT
                let price = 0;
                try {
                    price = await connector.fetchTicker(`${symbol}-USD`);
                } catch {
                    try {
                        price = await connector.fetchTicker(`${symbol}-USDT`);
                    } catch {}
                }

                if (price > 0) return { exchange: name, price };
            } catch (e) {
                // Silent failure for individual fetch
            }
            return null;
        });

        const results = await Promise.all(tasks);
        const prices = results.filter(r => r !== null) as { exchange: string, price: number }[];

        if (prices.length < 2) return null;

        // Find min and max
        prices.sort((a, b) => a.price - b.price);

        const lowest = prices[0];
        const highest = prices[prices.length - 1];

        const spread = highest.price - lowest.price;
        const spreadPct = (spread / lowest.price) * 100;

        // Arbitrage Threshold (e.g., 0.5% to cover fees)
        if (spreadPct > 0.5) {
            return {
                symbol,
                buyExchange: lowest.exchange,
                buyPrice: lowest.price,
                sellExchange: highest.exchange,
                sellPrice: highest.price,
                spread,
                spreadPct,
                timestamp: Date.now()
            };
        }

        return null;
    }
}
