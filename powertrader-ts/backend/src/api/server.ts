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
import { BacktestEngine } from '../engine/backtest/BacktestEngine';
import { HyperOpt } from '../extensions/hyperopt/HyperOpt';
import { DeepThinker } from '../thinker/DeepThinker';
import { HistoricalData } from '../engine/backtest/HistoricalData';

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

app.post('/api/hyperopt/run', async (req, res) => {
    try {
        console.log("[API] Running HyperOpt...", req.body);
        const config = req.body;

        // Defaults
        const hyperOptConfig = {
            strategyName: config.strategy || "SMAStrategy",
            pair: `${config.symbol || "BTC"}-USD`,
            timeframe: config.timeframe || "1h",
            startDate: config.startDate ? new Date(config.startDate).getTime() : (Date.now() - (30 * 24 * 60 * 60 * 1000)),
            endDate: config.endDate ? new Date(config.endDate).getTime() : Date.now(),
            populationSize: config.populationSize || 20,
            generations: config.generations || 10,
            parameterSpace: config.parameterSpace || {
                period: { min: 5, max: 100, step: 1 }
            }
        };

        const optimizer = new HyperOpt();
        const result = await optimizer.optimize(hyperOptConfig);

        res.json(result);
    } catch (e: any) {
        console.error(e);
        res.status(500).json({ error: e.message || "HyperOpt failed" });
    }
});

app.post('/api/strategy/backtest', async (req, res) => {
    try {
        console.log("[API] Running Full Backtest...", req.body);
        const symbol = req.body.symbol || "BTC";
        const strategyName = req.body.strategy || "SMAStrategy";
        const timeframe = req.body.timeframe || "1h";

        // Default to last 30 days if not specified
        const now = Date.now();
        const thirtyDays = 30 * 24 * 60 * 60 * 1000;
        const endDate = req.body.endDate ? new Date(req.body.endDate).getTime() : now;
        const startDate = req.body.startDate ? new Date(req.body.startDate).getTime() : (now - thirtyDays);
        const initialBalance = req.body.initialBalance || 10000;

        let strategy = StrategyFactory.get(strategyName);
        if (!strategy && strategyName === "Cointrade") {
             strategy = StrategyFactory.get("Cointrade (External)");
        }

        if (!strategy) {
             return res.status(400).json({ error: `Strategy ${strategyName} not found` });
        }

        const engine = new BacktestEngine();
        const result = await engine.run({
            strategy: strategyName,
            pair: `${symbol}-USD`,
            startDate,
            endDate,
            initialBalance,
            timeframe
        }, strategy);

        res.json(result);
    } catch (e: any) {
        console.error(e);
        res.status(500).json({ error: e.message || "Backtest failed" });
    }
});

// --- AI Evolution (DeepThinker) ---
const deepThinker = new DeepThinker();

app.post('/api/ai/train', async (req, res) => {
    try {
        const { symbol = "BTC", timeframe = "1h", lookback = 30 } = req.body;

        console.log(`[API] Training DeepThinker for ${symbol} ${timeframe}...`);

        const now = Date.now();
        const start = now - (lookback * 24 * 60 * 60 * 1000); // Days
        const candles = await HistoricalData.fetch(`${symbol}-USD`, timeframe, start, now);

        if (candles.length < 100) {
            return res.status(400).json({ error: "Insufficient data for training" });
        }

        // Run training in background (or await if simple)
        const history = await deepThinker.train(candles);

        res.json({ success: true, history });
    } catch (e: any) {
        console.error(e);
        res.status(500).json({ error: e.message });
    }
});

app.post('/api/ai/predict', async (req, res) => {
    try {
        const { symbol = "BTC", timeframe = "1h" } = req.body;

        // Fetch last 100 candles for context
        const now = Date.now();
        const start = now - (5 * 24 * 60 * 60 * 1000);
        const candles = await HistoricalData.fetch(`${symbol}-USD`, timeframe, start, now);

        const prediction = await deepThinker.predict(candles);
        const lastClose = candles[candles.length - 1].close;
        const direction = prediction > lastClose ? "UP" : "DOWN";

        res.json({
            prediction,
            lastClose,
            direction,
            diff: prediction - lastClose
        });
    } catch (e: any) {
        console.error(e);
        res.status(500).json({ error: e.message });
    }
});

// System Status Dashboard
app.get('/api/system/status', (req, res) => {
    res.json({
        version: "3.1.0",
        modules: {
            trader: { status: "active", version: "2.8.0" },
            thinker: { status: "active", engine: config.get("trading.active_ai") || "DeepThinker" },
            analytics: { status: "active", db: "sqlite" },
            notifications: { status: "active", enabled: !!config.get("notifications.enabled") }
        },
        exchanges: {
            active: config.get("trading.active_exchange"),
            available: ["robinhood", "kucoin", "binance", "coinbase", "uniswap"]
        },
        submodules: [
            { name: "cointrade", version: "1.4.2", path: "powertrader-ts/backend/src/modules/cointrade" },
            { name: "hyperopt", version: "1.1.0", path: "powertrader-ts/backend/src/extensions/hyperopt" }
        ],
        project_structure: {
            root: "/app/powertrader-ts",
            backend: "/app/powertrader-ts/backend",
            frontend: "/app/powertrader-ts/frontend",
            data: "/app/powertrader-ts/hub_data"
        }
    });
});

export function startServer() {
    const server = http.createServer(app);
    WebSocketManager.getInstance().initialize(server);

    server.listen(port, () => {
        console.log(`[API] Backend running at http://localhost:${port}`);
    });
}
