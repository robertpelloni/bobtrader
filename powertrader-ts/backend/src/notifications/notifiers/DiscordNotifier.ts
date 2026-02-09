import { BaseNotifier } from './BaseNotifier';
import { NotificationLevel } from '../types';
import axios from 'axios';
import { NotificationDatabase } from '../NotificationDatabase';

export class DiscordNotifier extends BaseNotifier {
    constructor(db: NotificationDatabase) {
        super(db);
    }

    public isAvailable(): boolean {
        const config = this.config.get("notifications");
        return !!config?.discord_webhook_url;
    }

    public async send(message: string, level: NotificationLevel): Promise<boolean> {
        if (!this.isAvailable()) return false;

        const config = this.config.get("notifications");
        const colorMap: Record<string, number> = {
            [NotificationLevel.INFO]: 0x00BFFF,
            [NotificationLevel.WARNING]: 0xFFAA00,
            [NotificationLevel.ERROR]: 0xFF0000,
            [NotificationLevel.CRITICAL]: 0x8B0000
        };

        const embed = {
            title: `PowerTrader AI - ${level.toUpperCase()}`,
            description: message,
            color: colorMap[level] || 0x00BFFF,
            timestamp: new Date().toISOString()
        };

        try {
            await axios.post(config.discord_webhook_url, {
                embeds: [embed]
            });
            this.log(true, level, message, 'discord');
            return true;
        } catch (e: any) {
            console.error(`[DiscordNotifier] Failed to send webhook:`, e);
            this.log(false, level, message, 'discord', e.message);
            return false;
        }
    }
}
