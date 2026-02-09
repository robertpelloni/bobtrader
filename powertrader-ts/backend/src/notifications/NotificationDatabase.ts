import Database from 'better-sqlite3';
import path from 'path';
import { ConfigManager } from '../config/ConfigManager';
import { INotificationRecord, INotificationStats } from './types';

export class NotificationDatabase {
    private db: Database.Database;

    constructor() {
        const config = ConfigManager.getInstance();
        const hubDir = config.get("trading.hub_data_dir") || "hub_data";
        // Assuming process.cwd() is backend/
        const dbPath = path.resolve(process.cwd(), '..', hubDir, 'notifications.db');

        this.db = new Database(dbPath);
        this.init();
    }

    private init(): void {
        this.db.exec(`
            CREATE TABLE IF NOT EXISTS notifications (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                timestamp TEXT NOT NULL,
                level TEXT NOT NULL,
                platform TEXT NOT NULL,
                message TEXT NOT NULL,
                success INTEGER NOT NULL,
                error_message TEXT,
                metadata TEXT
            );

            CREATE INDEX IF NOT EXISTS idx_notifications_timestamp ON notifications(timestamp);
            CREATE INDEX IF NOT EXISTS idx_notifications_level ON notifications(level);
        `);
    }

    public log(record: Omit<INotificationRecord, 'id'>): void {
        try {
            const stmt = this.db.prepare(`
                INSERT INTO notifications
                (timestamp, level, platform, message, success, error_message, metadata)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            `);

            stmt.run(
                record.timestamp.toISOString(),
                record.level,
                record.platform,
                record.message,
                record.success ? 1 : 0,
                record.error_message || null,
                record.metadata ? JSON.stringify(record.metadata) : null
            );
        } catch (e) {
            console.error("[Notifications] Failed to log notification:", e);
        }
    }

    public getRecent(limit: number = 20): INotificationRecord[] {
        const rows = this.db.prepare('SELECT * FROM notifications ORDER BY timestamp DESC LIMIT ?').all(limit) as any[];
        return rows.map(r => ({
            id: r.id,
            timestamp: new Date(r.timestamp),
            level: r.level,
            platform: r.platform,
            message: r.message,
            success: !!r.success,
            error_message: r.error_message,
            metadata: r.metadata ? JSON.parse(r.metadata) : undefined
        }));
    }

    public getStats(): INotificationStats {
        const total = (this.db.prepare('SELECT COUNT(*) as count FROM notifications').get() as { count: number }).count;
        const successful = (this.db.prepare('SELECT COUNT(*) as count FROM notifications WHERE success = 1').get() as { count: number }).count;

        // Simplified stats
        return {
            total: total,
            successful: successful,
            failed: total - successful,
            by_level: {}, // TODO: Implement detailed breakdown
            by_platform: {}, // TODO: Implement detailed breakdown
            success_rate: total > 0 ? (successful / total) * 100 : 0
        };
    }
}
