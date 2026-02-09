export enum NotificationLevel {
    INFO = "info",
    WARNING = "warning",
    ERROR = "error",
    CRITICAL = "critical"
}

export enum NotificationPlatform {
    EMAIL = "email",
    DISCORD = "discord",
    TELEGRAM = "telegram"
}

export interface INotificationRecord {
    id?: number;
    timestamp: Date;
    level: string;
    platform: string;
    message: string;
    success: boolean;
    error_message?: string;
    metadata?: any;
}

export interface INotificationStats {
    total: number;
    successful: number;
    failed: number;
    by_level: Record<string, { total: number; successful: number }>;
    by_platform: Record<string, { total: number; successful: number }>;
    success_rate: number;
}
