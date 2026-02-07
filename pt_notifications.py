#!/usr/bin/env python3
"""
PowerTrader AI - Notification System
====================================
Multi-platform notification system with Email, Discord, and Telegram support.
Features rate limiting, async support, and analytics integration.

Usage:
    from pt_notifications import NotificationManager

    manager = NotificationManager()
    await manager.send("Trade completed for BTC", level="info")
    await manager.send("Critical error: API key invalid", level="critical")

    # With specific platforms
    await manager.send("Important update", platforms=["discord", "telegram"])
"""

import asyncio
import logging
import sqlite3
from datetime import datetime, timedelta
from dataclasses import dataclass, asdict
from typing import List, Dict, Optional, Set, Any, Callable
from pathlib import Path
from contextlib import contextmanager, asynccontextmanager
from enum import Enum
import json

try:
    import yagmail

    EMAIL_AVAILABLE = True
except ImportError:
    EMAIL_AVAILABLE = False

try:
    from discord_webhook import DiscordWebhook, DiscordEmbed

    DISCORD_AVAILABLE = True
except ImportError:
    DISCORD_AVAILABLE = False

try:
    from telegram import Bot
    from telegram.error import TelegramError

    TELEGRAM_AVAILABLE = True
except ImportError:
    TELEGRAM_AVAILABLE = False

from pt_config import ConfigManager, NotificationConfig

DB_PATH = Path("hub_data/notifications.db")

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class NotificationLevel(Enum):
    INFO = "info"
    WARNING = "warning"
    ERROR = "error"
    CRITICAL = "critical"


class NotificationPlatform(Enum):
    EMAIL = "email"
    DISCORD = "discord"
    TELEGRAM = "telegram"


# NotificationConfig is imported from pt_config


@dataclass
class NotificationRecord:
    id: Optional[int]
    timestamp: datetime
    level: str
    platform: str
    message: str
    success: bool
    error_message: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None


class RateLimiter:
    def __init__(self, max_calls: int, period: timedelta = timedelta(minutes=1)):
        self.max_calls = max_calls
        self.period = period
        self.calls: List[datetime] = []
        self._lock = asyncio.Lock()

    async def acquire(self) -> bool:
        async with self._lock:
            now = datetime.now()
            self.calls = [c for c in self.calls if c > now - self.period]

            if len(self.calls) >= self.max_calls:
                return False

            self.calls.append(now)
            return True

    def reset(self):
        self.calls = []


class NotificationDatabase:
    def __init__(self, db_path: Path = DB_PATH):
        self.db_path = db_path
        self.db_path.parent.mkdir(parents=True, exist_ok=True)
        self._init_db()

    @contextmanager
    def _get_conn(self):
        conn = sqlite3.connect(self.db_path, detect_types=sqlite3.PARSE_DECLTYPES)
        conn.row_factory = sqlite3.Row
        try:
            yield conn
            conn.commit()
        finally:
            conn.close()

    def _init_db(self):
        with self._get_conn() as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS notifications (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    timestamp TIMESTAMP NOT NULL,
                    level TEXT NOT NULL,
                    platform TEXT NOT NULL,
                    message TEXT NOT NULL,
                    success BOOLEAN NOT NULL,
                    error_message TEXT,
                    metadata TEXT,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            """)

            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_notifications_timestamp ON notifications(timestamp)"
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_notifications_level ON notifications(level)"
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_notifications_platform ON notifications(platform)"
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_notifications_success ON notifications(success)"
            )

    def log_notification(
        self,
        level: str,
        platform: str,
        message: str,
        success: bool,
        error_message: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> NotificationRecord:
        timestamp = datetime.now()
        metadata_json = json.dumps(metadata) if metadata else None

        with self._get_conn() as conn:
            cursor = conn.execute(
                """
                INSERT INTO notifications 
                (timestamp, level, platform, message, success, error_message, metadata)
                VALUES (?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    timestamp,
                    level,
                    platform,
                    message,
                    success,
                    error_message,
                    metadata_json,
                ),
            )
            record_id = cursor.lastrowid

        return NotificationRecord(
            id=record_id,
            timestamp=timestamp,
            level=level,
            platform=platform,
            message=message,
            success=success,
            error_message=error_message,
            metadata=metadata,
        )

    def get_notifications(
        self,
        level: Optional[str] = None,
        platform: Optional[str] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        success: Optional[bool] = None,
        limit: int = 100,
    ) -> List[NotificationRecord]:
        query = "SELECT * FROM notifications WHERE 1=1"
        params = []

        if level:
            query += " AND level = ?"
            params.append(level)
        if platform:
            query += " AND platform = ?"
            params.append(platform)
        if start_date:
            query += " AND timestamp >= ?"
            params.append(start_date)
        if end_date:
            query += " AND timestamp <= ?"
            params.append(end_date)
        if success is not None:
            query += " AND success = ?"
            params.append(success)

        query += " ORDER BY timestamp DESC LIMIT ?"
        params.append(limit)

        with self._get_conn() as conn:
            rows = conn.execute(query, params).fetchall()

        records = []
        for r in rows:
            metadata = json.loads(r["metadata"]) if r["metadata"] else None
            records.append(
                NotificationRecord(
                    id=r["id"],
                    timestamp=datetime.fromisoformat(r["timestamp"])
                    if isinstance(r["timestamp"], str)
                    else r["timestamp"],
                    level=r["level"],
                    platform=r["platform"],
                    message=r["message"],
                    success=r["success"],
                    error_message=r["error_message"],
                    metadata=metadata,
                )
            )

        return records

    def get_statistics(
        self, start_date: Optional[datetime] = None, end_date: Optional[datetime] = None
    ) -> Dict[str, Any]:
        query = "SELECT * FROM notifications WHERE 1=1"
        params = []

        if start_date:
            query += " AND timestamp >= ?"
            params.append(start_date)
        if end_date:
            query += " AND timestamp <= ?"
            params.append(end_date)

        with self._get_conn() as conn:
            rows = conn.execute(query, params).fetchall()

        if not rows:
            return {
                "total": 0,
                "successful": 0,
                "failed": 0,
                "by_level": {},
                "by_platform": {},
                "success_rate": 0.0,
            }

        total = len(rows)
        successful = sum(1 for r in rows if r["success"])
        failed = total - successful

        by_level = {}
        by_platform = {}

        for r in rows:
            level = r["level"]
            platform = r["platform"]

            if level not in by_level:
                by_level[level] = {"total": 0, "successful": 0}
            by_level[level]["total"] += 1
            if r["success"]:
                by_level[level]["successful"] += 1

            if platform not in by_platform:
                by_platform[platform] = {"total": 0, "successful": 0}
            by_platform[platform]["total"] += 1
            if r["success"]:
                by_platform[platform]["successful"] += 1

        return {
            "total": total,
            "successful": successful,
            "failed": failed,
            "by_level": by_level,
            "by_platform": by_platform,
            "success_rate": (successful / total * 100) if total > 0 else 0.0,
        }


class BaseNotifier:
    def __init__(self, config: NotificationConfig, db: NotificationDatabase):
        self.config = config
        self.db = db
        self.enabled = False

    def is_available(self) -> bool:
        return False

    async def send(
        self, message: str, level: NotificationLevel = NotificationLevel.INFO, **kwargs
    ) -> bool:
        return False

    def _log(
        self, success: bool, level: str, message: str, error: Optional[str] = None
    ):
        self.db.log_notification(
            level=level,
            platform=self.__class__.__name__.replace("Notifier", "").lower(),
            message=message,
            success=success,
            error_message=error,
        )


class EmailNotifier(BaseNotifier):
    def __init__(self, config: NotificationConfig, db: NotificationDatabase):
        super().__init__(config, db)
        self.enabled = (
            config.email_address and config.email_app_password and EMAIL_AVAILABLE
        )
        self.yag = None

        if self.enabled:
            try:
                self.yag = yagmail.SMTP(config.email_address, config.email_app_password)
            except Exception as e:
                logger.error(f"Failed to initialize EmailNotifier: {e}")
                self.enabled = False

    def is_available(self) -> bool:
        return self.enabled and EMAIL_AVAILABLE

    async def send(
        self,
        message: str,
        level: NotificationLevel = NotificationLevel.INFO,
        subject: Optional[str] = None,
        **kwargs,
    ) -> bool:
        if not self.is_available():
            logger.warning("Email notifier not available")
            self._log(
                False, level.value, message, "Email not configured or unavailable"
            )
            return False

        try:
            if subject is None:
                subject = f"[PowerTrader AI - {level.value.upper()}] Notification"

            await asyncio.get_event_loop().run_in_executor(
                None,
                lambda: self.yag.send(
                    to=self.config.email_address, subject=subject, contents=message
                ),
            )

            self._log(True, level.value, message)
            logger.info(f"Email sent successfully: {message[:50]}...")
            return True

        except Exception as e:
            logger.error(f"Failed to send email: {e}")
            self._log(False, level.value, message, str(e))
            return False


class DiscordNotifier(BaseNotifier):
    def __init__(self, config: NotificationConfig, db: NotificationDatabase):
        super().__init__(config, db)
        self.enabled = config.discord_webhook_url and DISCORD_AVAILABLE

    def is_available(self) -> bool:
        return self.enabled and DISCORD_AVAILABLE

    async def send(
        self,
        message: str,
        level: NotificationLevel = NotificationLevel.INFO,
        embed_color: Optional[int] = None,
        **kwargs,
    ) -> bool:
        if not self.is_available():
            logger.warning("Discord notifier not available")
            self._log(
                False, level.value, message, "Discord not configured or unavailable"
            )
            return False

        try:
            webhook = DiscordWebhook(
                url=self.config.discord_webhook_url, content=message
            )

            if embed_color is None:
                color_map = {
                    NotificationLevel.INFO: 0x00BFFF,
                    NotificationLevel.WARNING: 0xFFAA00,
                    NotificationLevel.ERROR: 0xFF0000,
                    NotificationLevel.CRITICAL: 0x8B0000,
                }
                embed_color = color_map.get(level, 0x00BFFF)

            embed = DiscordEmbed(
                title=f"PowerTrader AI - {level.value.upper()}",
                description=message,
                color=embed_color,
            )
            embed.set_timestamp()

            webhook.add_embed(embed)

            await asyncio.get_event_loop().run_in_executor(
                None, lambda: webhook.execute()
            )

            self._log(True, level.value, message)
            logger.info(f"Discord message sent successfully: {message[:50]}...")
            return True

        except Exception as e:
            logger.error(f"Failed to send Discord message: {e}")
            self._log(False, level.value, message, str(e))
            return False


class TelegramNotifier(BaseNotifier):
    def __init__(self, config: NotificationConfig, db: NotificationDatabase):
        super().__init__(config, db)
        self.enabled = (
            config.telegram_bot_token and config.telegram_chat_id and TELEGRAM_AVAILABLE
        )
        self.bot = None

        if self.enabled:
            try:
                self.bot = Bot(token=config.telegram_bot_token)
            except Exception as e:
                logger.error(f"Failed to initialize TelegramNotifier: {e}")
                self.enabled = False

    def is_available(self) -> bool:
        return self.enabled and TELEGRAM_AVAILABLE

    async def send(
        self,
        message: str,
        level: NotificationLevel = NotificationLevel.INFO,
        parse_mode: Optional[str] = "HTML",
        **kwargs,
    ) -> bool:
        if not self.is_available():
            logger.warning("Telegram notifier not available")
            self._log(
                False, level.value, message, "Telegram not configured or unavailable"
            )
            return False

        try:
            formatted_message = f"<b>[{level.value.upper()}]</b>\n\n{message}"

            await asyncio.get_event_loop().run_in_executor(
                None,
                lambda: self.bot.send_message(
                    chat_id=self.config.telegram_chat_id,
                    text=formatted_message,
                    parse_mode=parse_mode,
                ),
            )

            self._log(True, level.value, message)
            logger.info(f"Telegram message sent successfully: {message[:50]}...")
            return True

        except TelegramError as e:
            logger.error(f"Failed to send Telegram message: {e}")
            self._log(False, level.value, message, str(e))
            return False
        except Exception as e:
            logger.error(f"Unexpected error sending Telegram message: {e}")
            self._log(False, level.value, message, str(e))
            return False

    async def close(self):
        if self.bot:
            await asyncio.get_event_loop().run_in_executor(
                None, lambda: self.bot.shutdown()
            )


class NotificationManager:
    def __init__(
        self,
        config: Optional[NotificationConfig] = None,
        db_path: Optional[Path] = None,
    ):
        self.db = NotificationDatabase(db_path) if db_path else NotificationDatabase()
        self.config = self._load_config() if config is None else config

        self.email_notifier = EmailNotifier(self.config, self.db)
        self.discord_notifier = DiscordNotifier(self.config, self.db)
        self.telegram_notifier = TelegramNotifier(self.config, self.db)

        self.rate_limiters = {
            "email": RateLimiter(self.config.rate_limit_emails_per_minute),
            "discord": RateLimiter(self.config.rate_limit_discord_per_minute),
            "telegram": RateLimiter(self.config.rate_limit_telegram_per_minute),
        }

        self.notifiers = {
            "email": self.email_notifier,
            "discord": self.discord_notifier,
            "telegram": self.telegram_notifier,
        }

    def _load_config(self) -> NotificationConfig:
        try:
            cm = ConfigManager()
            # Ensure notifications config is initialized
            if cm.get().notifications is None:
                cm.get().notifications = NotificationConfig()
                cm.save()
            return cm.get().notifications
        except Exception as e:
            logger.error(f"Failed to load config from ConfigManager, using defaults: {e}")
            return NotificationConfig()

    def save_config(self):
        try:
            cm = ConfigManager()
            cm.get().notifications = self.config
            cm.save()
            logger.info("Configuration saved via ConfigManager")
        except Exception as e:
            logger.error(f"Failed to save config via ConfigManager: {e}")

    def update_config(self, **kwargs):
        for key, value in kwargs.items():
            if hasattr(self.config, key):
                setattr(self.config, key, value)
        self.save_config()

    async def send(
        self,
        message: str,
        level: NotificationLevel = NotificationLevel.INFO,
        platforms: Optional[List[str]] = None,
        **kwargs,
    ) -> Dict[str, bool]:
        if not self.config.enabled:
            logger.info("Notifications are disabled globally")
            return {}

        if platforms is None:
            platforms = [
                p
                for p, enabled in self.config.platforms.items()
                if enabled
                and self.config.level_platforms.get(level.value, {}).get(p, False)
            ]

        results = {}
        tasks = []

        for platform in platforms:
            if platform not in self.notifiers:
                logger.warning(f"Unknown platform: {platform}")
                results[platform] = False
                continue

            notifier = self.notifiers[platform]

            if not notifier.is_available():
                logger.warning(f"Platform {platform} not available")
                results[platform] = False
                continue

            rate_limiter = self.rate_limiters[platform]

            if not await rate_limiter.acquire():
                logger.warning(f"Rate limit exceeded for {platform}")
                results[platform] = False
                continue

            tasks.append((platform, notifier.send(message, level, **kwargs)))

        for platform, task in tasks:
            try:
                results[platform] = await task
            except Exception as e:
                logger.error(f"Error sending to {platform}: {e}")
                results[platform] = False

        return results

    async def send_info(self, message: str, **kwargs) -> Dict[str, bool]:
        return await self.send(message, NotificationLevel.INFO, **kwargs)

    async def send_warning(self, message: str, **kwargs) -> Dict[str, bool]:
        return await self.send(message, NotificationLevel.WARNING, **kwargs)

    async def send_error(self, message: str, **kwargs) -> Dict[str, bool]:
        return await self.send(message, NotificationLevel.ERROR, **kwargs)

    async def send_critical(self, message: str, **kwargs) -> Dict[str, bool]:
        return await self.send(message, NotificationLevel.CRITICAL, **kwargs)

    def get_notifications(self, **kwargs) -> List[NotificationRecord]:
        return self.db.get_notifications(**kwargs)

    def get_statistics(self, **kwargs) -> Dict[str, Any]:
        return self.db.get_statistics(**kwargs)

    def print_statistics(
        self, start_date: Optional[datetime] = None, end_date: Optional[datetime] = None
    ):
        stats = self.get_statistics(start_date=start_date, end_date=end_date)

        print("\n" + "=" * 60)
        print("POWERTRADER AI - NOTIFICATION STATISTICS")
        print("=" * 60)

        period_str = "All time"
        if start_date and end_date:
            period_str = (
                f"{start_date.strftime('%Y-%m-%d')} to {end_date.strftime('%Y-%m-%d')}"
            )
        elif start_date:
            period_str = f"Since {start_date.strftime('%Y-%m-%d')}"
        elif end_date:
            period_str = f"Until {end_date.strftime('%Y-%m-%d')}"

        print(f"\nPeriod: {period_str}")
        print("-" * 40)
        print(f"Total Notifications:  {stats['total']:>10}")
        print(f"Successful:           {stats['successful']:>10}")
        print(f"Failed:               {stats['failed']:>10}")
        print(f"Success Rate:         {stats['success_rate']:>9.1f}%")

        if stats["by_level"]:
            print(f"\nBY LEVEL")
            print("-" * 40)
            for level, data in stats["by_level"].items():
                rate = (
                    (data["successful"] / data["total"] * 100)
                    if data["total"] > 0
                    else 0
                )
                print(
                    f"{level.upper():<10} {data['total']:>5} sent  {rate:>5.1f}% success"
                )

        if stats["by_platform"]:
            print(f"\nBY PLATFORM")
            print("-" * 40)
            for platform, data in stats["by_platform"].items():
                rate = (
                    (data["successful"] / data["total"] * 100)
                    if data["total"] > 0
                    else 0
                )
                print(
                    f"{platform.upper():<10} {data['total']:>5} sent  {rate:>5.1f}% success"
                )

        print("\n" + "=" * 60)

    async def close(self):
        if self.telegram_notifier.enabled:
            await self.telegram_notifier.close()


def create_notification_manager(
    email_address: Optional[str] = None,
    email_app_password: Optional[str] = None,
    discord_webhook_url: Optional[str] = None,
    telegram_bot_token: Optional[str] = None,
    telegram_chat_id: Optional[str] = None,
    enabled: bool = True,
) -> NotificationManager:
    config = NotificationConfig(
        enabled=enabled,
        email_address=email_address,
        email_app_password=email_app_password,
        discord_webhook_url=discord_webhook_url,
        telegram_bot_token=telegram_bot_token,
        telegram_chat_id=telegram_chat_id,
    )

    manager = NotificationManager(config)
    manager.save_config()

    return manager


async def test_notifications(manager: NotificationManager):
    print("\n" + "=" * 60)
    print("POWERTRADER AI - NOTIFICATION TEST")
    print("=" * 60)

    print("\nTesting INFO level...")
    results = await manager.send_info(
        "This is an INFO test notification from PowerTrader AI."
    )
    print(f"Results: {results}")

    print("\nTesting WARNING level...")
    results = await manager.send_warning(
        "This is a WARNING test notification from PowerTrader AI."
    )
    print(f"Results: {results}")

    print("\nTesting ERROR level...")
    results = await manager.send_error(
        "This is an ERROR test notification from PowerTrader AI."
    )
    print(f"Results: {results}")

    print("\nTesting CRITICAL level...")
    results = await manager.send_critical(
        "This is a CRITICAL test notification from PowerTrader AI."
    )
    print(f"Results: {results}")

    print("\n" + "=" * 60)


def main():
    import argparse

    parser = argparse.ArgumentParser(description="PowerTrader AI Notification System")
    subparsers = parser.add_subparsers(dest="command", help="Commands")

    subparsers.add_parser("test", help="Test notification system")

    stats_parser = subparsers.add_parser("stats", help="Show notification statistics")
    stats_parser.add_argument("--days", type=int, help="Statistics for last N days")

    list_parser = subparsers.add_parser("list", help="List recent notifications")
    list_parser.add_argument("--level", help="Filter by level")
    list_parser.add_argument("--platform", help="Filter by platform")
    list_parser.add_argument(
        "--limit", type=int, default=20, help="Number of notifications to show"
    )

    config_parser = subparsers.add_parser("config", help="Show or edit configuration")
    config_parser.add_argument("--email", help="Set email address")
    config_parser.add_argument("--discord", help="Set Discord webhook URL")
    config_parser.add_argument("--telegram-token", help="Set Telegram bot token")
    config_parser.add_argument("--telegram-chat", help="Set Telegram chat ID")

    args = parser.parse_args()

    manager = NotificationManager()

    if args.command == "test":
        asyncio.run(test_notifications(manager))

    elif args.command == "stats":
        start_date = None
        if args.days:
            start_date = datetime.now() - timedelta(days=args.days)
        manager.print_statistics(start_date=start_date)

    elif args.command == "list":
        notifications = manager.get_notifications(
            level=args.level, platform=args.platform, limit=args.limit
        )

        print("\n" + "=" * 80)
        print("RECENT NOTIFICATIONS")
        print("=" * 80)

        for n in notifications:
            status = "✓" if n.success else "✗"
            print(
                f"\n[{n.timestamp.strftime('%Y-%m-%d %H:%M:%S')}] {status} {n.platform.upper()} - {n.level.upper()}"
            )
            print(f"  {n.message}")
            if n.error_message:
                print(f"  Error: {n.error_message}")

        print("\n" + "=" * 80)

    elif args.command == "config":
        updates = {}
        if args.email:
            updates["email_address"] = args.email
        if args.discord:
            updates["discord_webhook_url"] = args.discord
        if args.telegram_token:
            updates["telegram_bot_token"] = args.telegram_token
        if args.telegram_chat:
            updates["telegram_chat_id"] = args.telegram_chat

        if updates:
            manager.update_config(**updates)
            print("Configuration updated successfully!")
        else:
            print("\nCurrent Configuration:")
            print("-" * 40)
            print(f"Enabled: {manager.config.enabled}")
            print(
                f"\nEmail: {'Configured' if manager.email_notifier.enabled else 'Not configured'}"
            )
            print(
                f"Discord: {'Configured' if manager.discord_notifier.enabled else 'Not configured'}"
            )
            print(
                f"Telegram: {'Configured' if manager.telegram_notifier.enabled else 'Not configured'}"
            )

            print(f"\nPlatform Settings:")
            for platform, enabled in manager.config.platforms.items():
                print(f"  {platform}: {enabled}")

            print(f"\nRate Limits:")
            print(f"  Email: {manager.config.rate_limit_emails_per_minute}/min")
            print(f"  Discord: {manager.config.rate_limit_discord_per_minute}/min")
            print(f"  Telegram: {manager.config.rate_limit_telegram_per_minute}/min")

    else:
        parser.print_help()
        print("\nQuick Setup:")
        print("  1. Edit config: pt_notifications.py config --email your@email.com")
        print("  2. Test notifications: pt_notifications.py test")
        print("  3. View statistics: pt_notifications.py stats")


if __name__ == "__main__":
    main()
