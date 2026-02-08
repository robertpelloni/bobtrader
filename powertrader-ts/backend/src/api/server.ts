import express from 'express';
import cors from 'cors';
import bodyParser from 'body-parser';
import http from 'http';
import { ConfigManager } from '../config/ConfigManager';
import { AnalyticsManager } from '../analytics/AnalyticsManager';
import { WebSocketManager } from './websocket';

const app = express();
const port = 3000;

app.use(cors());
app.use(bodyParser.json());

const config = ConfigManager.getInstance();
const analytics = new AnalyticsManager();

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

export function startServer() {
    // Use http.createServer to support both Express and WebSocket on same port
    const server = http.createServer(app);

    // Initialize WebSocket
    WebSocketManager.getInstance().initialize(server);

    server.listen(port, () => {
        console.log(`[API] Backend running at http://localhost:${port}`);
    });
}
