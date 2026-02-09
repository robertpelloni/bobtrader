import { BaseNotifier } from './BaseNotifier';
import { NotificationLevel } from '../types';
import axios from 'axios';
import { NotificationDatabase } from '../NotificationDatabase';

export class TelegramNotifier extends BaseNotifier {
    constructor(db: NotificationDatabase) {
        super(db);
    }

    public isAvailable(): boolean {
        const config = this.config.get("notifications");
        return !!(config?.telegram_bot_token && config?.telegram_chat_id);
    }

    public async send(message: string, level: NotificationLevel): Promise<boolean> {
        if (!this.isAvailable()) return false;

        const config = this.config.get("notifications");
        const url = `https://api.telegram.org/bot${config.telegram_bot_token}/sendMessage`;
        const text = `<b>[${level.toUpperCase()}]</b>\n\n${message}`;

        try {
            await axios.post(url, {
                chat_id: config.telegram_chat_id,
                text: text,
                parse_mode: 'HTML'
            });
            this.log(true, level, message, 'telegram');
            return true;
        } catch (e: any) {
            console.error(`[TelegramNotifier] Failed to send message:`, e);
            this.log(false, level, message, 'telegram', e.message);
            return false;
        }
    }
}
