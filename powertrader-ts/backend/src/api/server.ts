import express from 'express';
import cors from 'cors';
import bodyParser from 'body-parser';
import http from 'http';
import { ConfigManager } from '../config/ConfigManager';
import { AnalyticsManager } from '../analytics/AnalyticsManager';
import { WebSocketManager } from './websocket';
import { CointradeAdapter } from '../modules/cointrade/CointradeAdapter';
import { KuCoinConnector } from '../exchanges/KuCoinConnector';
import { StrategyFactory } from '../engine/strategy/StrategyFactory';

const app = express();
const port = 3000;

app.use(cors());
app.use(bodyParser.json());

const config = ConfigManager.getInstance();
const analytics = new AnalyticsManager();
const kucoin = new KuCoinConnector();

// --- API ROUTES ---

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
    // In a real implementation: config.set('trading', req.body);
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

// Strategy Management
app.get('/api/strategies', (req, res) => {
    const strategies = StrategyFactory.getAll().map(s => ({
        name: s.name,
        interval: s.interval
    }));
    const active = config.get('trading.active_strategy') || "SMAStrategy";
    res.json({ strategies, active });
});

app.post('/api/strategies/config', (req, res) => {
    const { strategy } = req.body;
    // In real app, we would ConfigManager.set('trading.active_strategy', strategy)
    console.log(`[API] Switching active strategy to: ${strategy}`);
    res.json({ success: true, active: strategy });
});

app.post('/api/strategy/backtest', async (req, res) => {
    try {
        console.log("[API] Running Strategy Backtest...", req.body);
        const symbol = req.body.symbol || "BTC";
        const strategyName = req.body.strategy || "Cointrade";

        console.log(`[API] Fetching real data for ${symbol}...`);
        const candles = await kucoin.fetchOHLCV(`${symbol}-USD`, '1h', 100);

        if (candles.length === 0) {
            return res.status(404).json({ error: "No market data found" });
        }

        let strategy = StrategyFactory.get(strategyName);
        if (!strategy && strategyName === "Cointrade") {
             // Special case for Cointrade if not registered in factory yet or named differently
             strategy = StrategyFactory.get("Cointrade (External)");
        }

        if (!strategy) {
             return res.status(400).json({ error: `Strategy ${strategyName} not found` });
        }

        const enrichedData = await strategy.populateIndicators(candles);
        const withBuy = await strategy.populateBuyTrend(enrichedData);
        const resultData = await strategy.populateSellTrend(withBuy);

        const frontendData = resultData.map((c: any) => ({
            time: c.timestamp,
            price: c.close,
            // Map strategy-specific indicators to generic chart fields
            rsi: c.rsi || 50,
            macd: c.macd || 0,
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
