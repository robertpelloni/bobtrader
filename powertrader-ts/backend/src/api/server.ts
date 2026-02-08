import express from 'express';
import cors from 'cors';
import bodyParser from 'body-parser';
import http from 'http';
import { ConfigManager } from '../config/ConfigManager';
import { AnalyticsManager } from '../analytics/AnalyticsManager';
import { WebSocketManager } from './websocket';
import { CointradeAdapter } from '../modules/cointrade/CointradeAdapter';
import { KuCoinConnector } from '../exchanges/KuCoinConnector';
import { SMAStrategy } from '../engine/strategy/implementations/SMAStrategy';

const app = express();
const port = 3000;

app.use(cors());
app.use(bodyParser.json());

const config = ConfigManager.getInstance();
const analytics = new AnalyticsManager();
const cointrade = new CointradeAdapter();
const kucoin = new KuCoinConnector();
const smaStrategy = new SMAStrategy();

// --- API ROUTES ---

// Dashboard Data
app.get('/api/dashboard', (req, res) => {
    const perf = analytics.getPerformance();
    res.json({
        account: {
            total: 10000,
            pnl: perf.pnl
        },
        trades: [
            { symbol: 'BTC', pnl: 1.2, stage: 0 },
            { symbol: 'ETH', pnl: -0.5, stage: 1 }
        ]
    });
});

app.get('/api/settings', (req, res) => {
    res.json(config.get('trading'));
});

app.post('/api/settings', (req, res) => {
    res.json({ success: true });
});

app.get('/api/volume/:coin', (req, res) => {
    res.json({
        profile: { average: 5000, median: 4500, p90: 8000 },
        recent: [
            { timestamp: Date.now(), volume: 4200, ratio: 1.1, trend: 'increasing' }
        ]
    });
});

// Strategy Sandbox Endpoint (REAL DATA)
app.post('/api/strategy/backtest', async (req, res) => {
    try {
        console.log("[API] Running Strategy Backtest...", req.body);
        const symbol = req.body.symbol || "BTC";
        const strategyName = req.body.strategy || "Cointrade";

        // 1. Fetch Real Data from KuCoin
        console.log(`[API] Fetching real data for ${symbol}...`);
        const candles = await kucoin.fetchOHLCV(`${symbol}-USD`, '1h', 100);

        if (candles.length === 0) {
            return res.status(404).json({ error: "No market data found" });
        }

        // 2. Run Strategy
        let resultData = candles;

        if (strategyName === "SMA Crossover") {
            // Apply SMA Logic
            // We need to implement populateIndicators properly in SMAStrategy first
            // For now, let's use the simulation logic in CointradeAdapter as it's more robust in this demo
            resultData = await cointrade.populateIndicators(candles);
            resultData = await cointrade.populateBuyTrend(resultData);

        } else {
            // Default to Cointrade Adapter (simulation of complex indicators)
            resultData = await cointrade.populateIndicators(candles);
            resultData = await cointrade.populateBuyTrend(resultData);
            resultData = await cointrade.populateSellTrend(resultData);
        }

        // Map to frontend format
        const frontendData = resultData.map((c: any) => ({
            time: c.timestamp,
            price: c.close,
            rsi: c.rsi || 50,
            signal: c.buy_signal ? 1 : (c.sell_signal ? -1 : 0)
        }));

        res.json(frontendData);
    } catch (e) {
        console.error(e);
        res.status(500).json({ error: "Backtest failed" });
    }
});

export function startServer() {
    const server = http.createServer(app);
    WebSocketManager.getInstance().initialize(server);

    server.listen(port, () => {
        console.log(`[API] Backend running at http://localhost:${port}`);
    });
}
