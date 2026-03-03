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

    public async scan(symbol: string, orderSizeUsd: number = 1000): Promise<Opportunity | null> {
        // Advanced Arbitrage Scanner: Checks Order Book Depth (Slippage) instead of just top-of-book ticker.
        const tasks = Array.from(this.exchanges.entries()).map(async ([name, connector]) => {
            try {
                // We need to fetch the Order Book to calculate effective buy/sell prices for `orderSizeUsd`
                let obPair = `${symbol}-USD`;
                let book: any = null;

                try {
                    book = await connector.fetchOrderBook(obPair);
                } catch {
                    obPair = `${symbol}-USDT`;
                    try {
                        book = await connector.fetchOrderBook(obPair);
                    } catch {}
                }

                if (book && book.asks && book.bids && book.asks.length > 0 && book.bids.length > 0) {
                    // Calculate effective BUY price (slippage into Asks)
                    let buyCost = 0;
                    let amountToBuy = orderSizeUsd;
                    let effBuyPrice = 0;

                    for (const ask of book.asks) {
                        const price = ask[0];
                        const amountAvailableUsd = price * ask[1];
                        if (amountToBuy <= amountAvailableUsd) {
                            buyCost += amountToBuy;
                            effBuyPrice = buyCost / (orderSizeUsd / price); // Approximation
                            break;
                        } else {
                            buyCost += amountAvailableUsd;
                            amountToBuy -= amountAvailableUsd;
                        }
                    }
                    if (effBuyPrice === 0) effBuyPrice = book.asks[0][0]; // Fallback if book too thin

                    // Calculate effective SELL price (slippage into Bids)
                    let sellRevenue = 0;
                    let amountToSellUsd = orderSizeUsd;
                    let effSellPrice = 0;

                    for (const bid of book.bids) {
                        const price = bid[0];
                        const amountAvailableUsd = price * bid[1];
                        if (amountToSellUsd <= amountAvailableUsd) {
                            sellRevenue += amountToSellUsd;
                            effSellPrice = sellRevenue / (orderSizeUsd / price);
                            break;
                        } else {
                            sellRevenue += amountAvailableUsd;
                            amountToSellUsd -= amountAvailableUsd;
                        }
                    }
                    if (effSellPrice === 0) effSellPrice = book.bids[0][0];

                    return { exchange: name, effBuyPrice, effSellPrice };
                }

                // Fallback to Ticker if OB fails or is unsupported
                let price = await connector.fetchTicker(obPair);
                if (price > 0) return { exchange: name, effBuyPrice: price, effSellPrice: price };

            } catch (e) {
                // Silent failure for individual fetch
            }
            return null;
        });

        const results = await Promise.all(tasks);
        const exData = results.filter(r => r !== null) as { exchange: string, effBuyPrice: number, effSellPrice: number }[];

        if (exData.length < 2) return null;

        // Find the absolute best place to BUY (lowest Ask) and best place to SELL (highest Bid)
        let bestBuy = exData[0];
        let bestSell = exData[0];

        for (const data of exData) {
            if (data.effBuyPrice < bestBuy.effBuyPrice) bestBuy = data;
            if (data.effSellPrice > bestSell.effSellPrice) bestSell = data;
        }

        if (bestBuy.exchange === bestSell.exchange) return null; // No arb on same exchange

        const spread = bestSell.effSellPrice - bestBuy.effBuyPrice;
        const spreadPct = (spread / bestBuy.effBuyPrice) * 100;

        // Arbitrage Threshold (e.g., 0.5% to cover fees)
        if (spreadPct > 0.5) {
            return {
                symbol,
                buyExchange: bestBuy.exchange,
                buyPrice: bestBuy.effBuyPrice,
                sellExchange: bestSell.exchange,
                sellPrice: bestSell.effSellPrice,
                spread,
                spreadPct,
                timestamp: Date.now()
            };
        }

        return null;
    }
}
