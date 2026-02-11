import { startServer } from './api/server';
import { Trader } from './trader/Trader';
import { RobinhoodConnector } from './exchanges/RobinhoodConnector';
import { KuCoinConnector } from './exchanges/KuCoinConnector';
import { BinanceConnector } from './exchanges/BinanceConnector';
import { CoinbaseConnector } from './exchanges/CoinbaseConnector';
import { UniswapConnector } from './exchanges/UniswapConnector';
import { PaperExchange } from './extensions/paper_trading/PaperExchange';
import { ConfigManager } from './config/ConfigManager';

// Initialize Components
const config = ConfigManager.getInstance();
const tradingConfig = config.get("trading");
const mode = tradingConfig?.execution_mode || "paper";
const activeExchange = tradingConfig?.active_exchange || "robinhood";
const exchangeConfig = config.get("exchanges") || {};
const defiConfig = config.get("defi") || {};

let exchange;
if (mode === "paper") {
    exchange = new PaperExchange(10000);
    console.log("[System] Running in PAPER mode");
} else {
    // LIVE MODE
    console.log(`[System] Running in LIVE mode on ${activeExchange.toUpperCase()}`);

    switch (activeExchange) {
        case "kucoin":
            exchange = new KuCoinConnector(
                exchangeConfig.kucoin?.key,
                exchangeConfig.kucoin?.secret,
                exchangeConfig.kucoin?.passphrase
            );
            break;
        case "binance":
            exchange = new BinanceConnector(
                exchangeConfig.binance?.key,
                exchangeConfig.binance?.secret
            );
            break;
        case "coinbase":
            exchange = new CoinbaseConnector(
                exchangeConfig.coinbase?.key,
                exchangeConfig.coinbase?.secret
            );
            break;
        case "uniswap":
            exchange = new UniswapConnector(
                defiConfig.rpc_url,
                defiConfig.private_key
            );
            break;
        case "robinhood":
        default:
            exchange = new RobinhoodConnector(
                exchangeConfig.robinhood?.key || process.env.RH_KEY || "",
                exchangeConfig.robinhood?.secret || process.env.RH_SECRET || ""
            );
            break;
    }
}

const trader = new Trader(exchange);

// Start Services
trader.start();
startServer();
