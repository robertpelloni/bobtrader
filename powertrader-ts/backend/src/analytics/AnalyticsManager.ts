import Database from 'better-sqlite3';
import path from 'path';

export class AnalyticsManager {
    private db: Database.Database;

    constructor() {
        const dbPath = path.join(process.cwd(), 'trades.db');
        this.db = new Database(dbPath);
        this.init();
    }

    private init(): void {
        this.db.exec(`
            CREATE TABLE IF NOT EXISTS trades (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                symbol TEXT NOT NULL,
                side TEXT NOT NULL,
                amount REAL NOT NULL,
                price REAL NOT NULL,
                timestamp INTEGER NOT NULL
            )
        `);
    }

    public logTrade(trade: any): void {
        const stmt = this.db.prepare('INSERT INTO trades (symbol, side, amount, price, timestamp) VALUES (?, ?, ?, ?, ?)');
        stmt.run(trade.symbol, trade.side, trade.amount, trade.price, Date.now());
    }

    public getPerformance(): any {
        // Mock performance calculation
        return {
            totalTrades: 150,
            winRate: 0.65,
            pnl: 1250.50
        };
    }
}
