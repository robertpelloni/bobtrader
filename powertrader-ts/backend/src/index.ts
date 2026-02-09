import { startServer } from './api/server';
import { Trader } from './trader/Trader';
import { RobinhoodConnector } from './exchanges/RobinhoodConnector';
import { PaperExchange } from './extensions/paper_trading/PaperExchange';
import { ConfigManager } from './config/ConfigManager';

// Initialize Components
const config = ConfigManager.getInstance();
const tradingConfig = config.get("trading");
const mode = tradingConfig?.execution_mode || "paper";

let exchange;
if (mode === "live") {
    // TODO: Load keys from secure storage or environment
    exchange = new RobinhoodConnector(process.env.RH_KEY || "", process.env.RH_SECRET || "");
    console.log("[System] Running in LIVE mode with RobinhoodConnector");
} else {
    exchange = new PaperExchange(10000);
    console.log("[System] Running in PAPER mode");
}

const trader = new Trader(exchange);

// Start Services
trader.start();
startServer();
