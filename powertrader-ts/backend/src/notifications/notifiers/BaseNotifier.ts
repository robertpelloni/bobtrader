import { INotificationRecord, NotificationLevel } from '../types';
import { NotificationDatabase } from '../NotificationDatabase';
import { ConfigManager } from '../../config/ConfigManager';

export abstract class BaseNotifier {
    protected db: NotificationDatabase;
    protected config: ConfigManager;

    constructor(db: NotificationDatabase) {
        this.db = db;
        this.config = ConfigManager.getInstance();
    }

    public abstract isAvailable(): boolean;

    public abstract send(message: string, level: NotificationLevel, ...args: any[]): Promise<boolean>;

    protected log(success: boolean, level: string, message: string, platform: string, error?: string): void {
        this.db.log({
            timestamp: new Date(),
            level: level,
            platform: platform,
            message: message,
            success: success,
            error_message: error
        });
    }
}
