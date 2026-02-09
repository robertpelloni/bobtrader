import { ConfigManager } from '../config/ConfigManager';
import { NotificationDatabase } from './NotificationDatabase';
import { RateLimiter } from './RateLimiter';
import { NotificationLevel, NotificationPlatform, INotificationRecord, INotificationStats } from './types';
import { EmailNotifier } from './notifiers/EmailNotifier';
import { DiscordNotifier } from './notifiers/DiscordNotifier';
import { TelegramNotifier } from './notifiers/TelegramNotifier';
import { BaseNotifier } from './notifiers/BaseNotifier';

export class NotificationManager {
    private static instance: NotificationManager;
    private configManager: ConfigManager;
    private db: NotificationDatabase;
    private notifiers: Record<string, BaseNotifier>;
    private rateLimiters: Record<string, RateLimiter>;

    private constructor() {
        this.configManager = ConfigManager.getInstance();
        this.db = new NotificationDatabase();

        this.notifiers = {
            [NotificationPlatform.EMAIL]: new EmailNotifier(this.db),
            [NotificationPlatform.DISCORD]: new DiscordNotifier(this.db),
            [NotificationPlatform.TELEGRAM]: new TelegramNotifier(this.db)
        };

        const config = this.configManager.get("notifications") || {};

        this.rateLimiters = {
            [NotificationPlatform.EMAIL]: new RateLimiter(config.rate_limit_emails_per_minute || 5),
            [NotificationPlatform.DISCORD]: new RateLimiter(config.rate_limit_discord_per_minute || 30),
            [NotificationPlatform.TELEGRAM]: new RateLimiter(config.rate_limit_telegram_per_minute || 20)
        };
    }

    public static getInstance(): NotificationManager {
        if (!NotificationManager.instance) {
            NotificationManager.instance = new NotificationManager();
        }
        return NotificationManager.instance;
    }

    public async send(message: string, level: NotificationLevel = NotificationLevel.INFO, platforms?: NotificationPlatform[]): Promise<Record<string, boolean>> {
        const config = this.configManager.get("notifications");
        if (!config?.enabled) {
            return {};
        }

        const results: Record<string, boolean> = {};
        const targetPlatforms = platforms || this.getActivePlatforms(level, config);

        for (const platform of targetPlatforms) {
            const notifier = this.notifiers[platform];
            const rateLimiter = this.rateLimiters[platform];

            if (notifier && notifier.isAvailable()) {
                if (rateLimiter.acquire()) {
                    try {
                        results[platform] = await notifier.send(message, level);
                    } catch (e) {
                         console.error(`[NotificationManager] Error sending to ${platform}:`, e);
                         results[platform] = false;
                    }
                } else {
                    console.warn(`[NotificationManager] Rate limit exceeded for ${platform}`);
                    results[platform] = false;
                }
            }
        }

        return results;
    }

    private getActivePlatforms(level: NotificationLevel, config: any): NotificationPlatform[] {
        const active: NotificationPlatform[] = [];

        if (!config?.platforms) return [];

        const levelPlatforms = config.level_platforms?.[level] || {};

        for (const [platform, enabled] of Object.entries(config.platforms)) {
             if (enabled) {
                 // Default to true if not explicitly disabled for this level
                 if (levelPlatforms[platform] !== false) {
                     active.push(platform as NotificationPlatform);
                 }
             }
        }
        return active;
    }

    public async info(message: string) { return this.send(message, NotificationLevel.INFO); }
    public async warning(message: string) { return this.send(message, NotificationLevel.WARNING); }
    public async error(message: string) { return this.send(message, NotificationLevel.ERROR); }
    public async critical(message: string) { return this.send(message, NotificationLevel.CRITICAL); }

    public getRecent(limit: number = 20): INotificationRecord[] {
        return this.db.getRecent(limit);
    }

    public getStats(): INotificationStats {
        return this.db.getStats();
    }
}
