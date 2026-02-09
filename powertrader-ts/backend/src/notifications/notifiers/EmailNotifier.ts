import { BaseNotifier } from './BaseNotifier';
import { NotificationLevel } from '../types';
import nodemailer from 'nodemailer';
import { NotificationDatabase } from '../NotificationDatabase';

export class EmailNotifier extends BaseNotifier {
    private transporter?: nodemailer.Transporter;

    constructor(db: NotificationDatabase) {
        super(db);
        this.init();
    }

    private init() {
        const config = this.config.get("notifications");
        if (config?.email_address && config?.email_app_password) {
            this.transporter = nodemailer.createTransport({
                service: 'gmail',
                auth: {
                    user: config.email_address,
                    pass: config.email_app_password
                }
            });
        }
    }

    public isAvailable(): boolean {
        return !!this.transporter;
    }

    public async send(message: string, level: NotificationLevel, subject?: string): Promise<boolean> {
        if (!this.transporter) return false;

        const config = this.config.get("notifications");
        const finalSubject = subject || `[PowerTrader AI - ${level.toUpperCase()}] Notification`;

        try {
            await this.transporter.sendMail({
                from: config.email_address,
                to: config.email_address,
                subject: finalSubject,
                text: message
            });
            this.log(true, level, message, 'email');
            return true;
        } catch (e: any) {
            console.error(`[EmailNotifier] Failed to send email:`, e);
            this.log(false, level, message, 'email', e.message);
            return false;
        }
    }
}
