"""
Configuration Management System for PowerTrader AI

This module provides centralized configuration management with validation,
hot-reload support, environment variable support, and migration path
from existing scattered settings.

Author: PowerTrader AI Team
Version: 2.0.0
Created: 2026-01-18
License: Apache 2.0
"""

import os
import json
import yaml
import logging
from typing import Dict, Any, Optional, List, Type, TypeVar, get_type_hints
from dataclasses import dataclass, asdict, fields
from datetime import timedelta
from pathlib import Path
import threading
import hashlib


T = TypeVar("T")


@dataclass
class TradingConfig:
    main_neural_dir: str = ""
    coins: List[str] = None
    trade_start_level: int = 3
    start_allocation_pct: float = 0.005
    dca_multiplier: float = 2.0
    dca_levels: List[float] = None
    max_dca_buys_per_24h: int = 2
    pm_start_pct_no_dca: float = 5.0
    pm_start_pct_with_dca: float = 2.5
    trailing_gap_pct: float = 0.5
    default_timeframe: str = "1hour"
    timeframes: List[str] = None
    candles_limit: int = 120
    ui_refresh_seconds: float = 1.0
    chart_refresh_seconds: float = 10.0
    hub_data_dir: str = ""
    script_neural_runner2: str = "pt_thinker.py"
    script_neural_trainer: str = "pt_trainer.py"
    script_trader: str = "pt_trader.py"
    auto_start_scripts: bool = False

    def __post_init__(self):
        if self.coins is None:
            self.coins = ["BTC", "ETH", "XRP", "BNB", "DOGE"]
        if self.dca_levels is None:
            self.dca_levels = [-2.5, -5.0, -10.0, -20.0, -30.0, -40.0, -50.0]
        if self.timeframes is None:
            self.timeframes = [
                "1min",
                "5min",
                "15min",
                "30min",
                "1hour",
                "2hour",
                "4hour",
                "8hour",
                "12hour",
                "1day",
                "1week",
            ]


@dataclass
class NotificationPlatformConfig:
    enabled: bool = True
    platforms: Dict[str, bool] = None

    def __post_init__(self):
        if self.platforms is None:
            self.platforms = {"email": True, "discord": True, "telegram": True}


@dataclass
class NotificationConfig:
    enabled: bool = True
    platforms: Dict[str, bool] = None
    email_address: Optional[str] = None
    email_app_password: Optional[str] = None
    discord_webhook_url: Optional[str] = None
    telegram_bot_token: Optional[str] = None
    telegram_chat_id: Optional[str] = None
    rate_limit_emails_per_minute: int = 5
    rate_limit_discord_per_minute: int = 10
    rate_limit_telegram_per_minute: int = 10
    level_platforms: Dict[str, Dict[str, bool]] = None

    def __post_init__(self):
        if self.platforms is None:
            self.platforms = {"email": True, "discord": True, "telegram": True}
        if self.level_platforms is None:
            self.level_platforms = {
                "info": {"email": True, "discord": True, "telegram": True},
                "warning": {"email": True, "discord": True, "telegram": True},
                "error": {"email": True, "discord": True, "telegram": True},
                "critical": {"email": True, "discord": True, "telegram": True},
            }


@dataclass
class ExchangeConfig:
    kucoin_api_key: Optional[str] = None
    kucoin_api_secret: Optional[str] = None
    kucoin_api_passphrase: Optional[str] = None
    binance_api_key: Optional[str] = None
    binance_api_secret: Optional[str] = None
    coinbase_api_key: Optional[str] = None
    coinbase_api_secret: Optional[str] = None


@dataclass
class AnalyticsConfig:
    enabled: bool = True
    database_path: str = "hub_data/trades.db"
    retention_days: int = 365
    log_trades: bool = True
    log_performance: bool = True


@dataclass
class PositionSizingConfig:
    enabled: bool = False
    default_risk_pct: float = 0.02
    min_risk_pct: float = 0.01
    max_risk_pct: float = 0.10


@dataclass
class CorrelationConfig:
    enabled: bool = False
    alert_threshold: float = 0.8
    check_periods: List[int] = None

    def __post_init__(self):
        if self.check_periods is None:
            self.check_periods = [7, 30, 90]


@dataclass
class SystemConfig:
    log_level: str = "INFO"
    log_file: str = "hub_data/powertrader.log"
    max_log_size_mb: int = 10
    backup_log_count: int = 5
    debug_mode: bool = False


@dataclass
class PowerTraderConfig:
    trading: TradingConfig = None
    notifications: NotificationConfig = None
    exchanges: ExchangeConfig = None
    analytics: AnalyticsConfig = None
    position_sizing: PositionSizingConfig = None
    correlation: CorrelationConfig = None
    system: SystemConfig = None

    def __post_init__(self):
        if self.trading is None:
            self.trading = TradingConfig()
        if self.notifications is None:
            self.notifications = NotificationConfig()
        if self.exchanges is None:
            self.exchanges = ExchangeConfig()
        if self.analytics is None:
            self.analytics = AnalyticsConfig()
        if self.position_sizing is None:
            self.position_sizing = PositionSizingConfig()
        if self.correlation is None:
            self.correlation = CorrelationConfig()
        if self.system is None:
            self.system = SystemConfig()


class ConfigValidator:
    """Validates configuration against schema and constraints."""

    @staticmethod
    def validate_trading(config: TradingConfig) -> List[str]:
        errors = []

        if not 1 <= config.trade_start_level <= 7:
            errors.append("trade_start_level must be between 1 and 7")

        if not 0.001 <= config.start_allocation_pct <= 0.5:
            errors.append("start_allocation_pct must be between 0.1% and 50%")

        if not 1.0 <= config.dca_multiplier <= 5.0:
            errors.append("dca_multiplier must be between 1.0 and 5.0")

        if not 0 <= config.max_dca_buys_per_24h <= 10:
            errors.append("max_dca_buys_per_24h must be between 0 and 10")

        if not 1.0 <= config.pm_start_pct_no_dca <= 50.0:
            errors.append("pm_start_pct_no_dca must be between 1% and 50%")

        if not 1.0 <= config.pm_start_pct_with_dca <= 50.0:
            errors.append("pm_start_pct_with_dca must be between 1% and 50%")

        if not 0.1 <= config.trailing_gap_pct <= 5.0:
            errors.append("trailing_gap_pct must be between 0.1% and 5.0")

        if config.default_timeframe not in config.timeframes:
            errors.append(f"default_timeframe must be one of {config.timeframes}")

        return errors

    @staticmethod
    def validate_notifications(config: NotificationConfig) -> List[str]:
        errors = []

        if config.enabled:
            if config.platforms.get("email", False):
                if not config.email_address:
                    errors.append("Email platform enabled but email_address not set")
            if config.platforms.get("discord", False):
                if not config.discord_webhook_url:
                    errors.append(
                        "Discord platform enabled but discord_webhook_url not set"
                    )
            if config.platforms.get("telegram", False):
                if not config.telegram_bot_token:
                    errors.append(
                        "Telegram platform enabled but telegram_bot_token not set"
                    )
                if not config.telegram_chat_id:
                    errors.append(
                        "Telegram platform enabled but telegram_chat_id not set"
                    )

        return errors

    @staticmethod
    def validate_system(config: SystemConfig) -> List[str]:
        errors = []

        valid_log_levels = ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]
        if config.log_level not in valid_log_levels:
            errors.append(f"log_level must be one of {valid_log_levels}")

        return errors

    @classmethod
    def validate(cls, config: PowerTraderConfig) -> List[str]:
        """Validate entire configuration."""
        all_errors = []

        all_errors.extend(cls.validate_trading(config.trading))
        all_errors.extend(cls.validate_notifications(config.notifications))
        all_errors.extend(cls.validate_system(config.system))

        return all_errors


class ConfigManager:
    """Centralized configuration management with hot-reload support."""

    CONFIG_FILE = "config.yaml"
    ENV_PREFIX = "POWERTRADER_"
    _instance = None
    _lock = threading.Lock()

    def __new__(cls, *args, **kwargs):
        with cls._lock:
            if cls._instance is None:
                cls._instance = super(ConfigManager, cls).__new__(cls)
        return cls._instance

    def __init__(self, config_dir: Optional[str] = None):
        self.config_dir = Path(config_dir) if config_dir else Path.cwd()
        self.config_path = self.config_dir / self.CONFIG_FILE
        self._config: Optional[PowerTraderConfig] = None
        self._config_hash: Optional[str] = None
        self._watcher_thread: Optional[threading.Thread] = None
        self._watcher_running = False
        self._callbacks: List[callable] = []

        self._load_config()

    def _get_env_var(self, key: str, default: Any = None) -> Any:
        """Get environment variable with POWERTRADER_ prefix."""
        env_key = f"{self.ENV_PREFIX}{key}"
        value = os.getenv(env_key, default)

        if value is not None:
            if value.lower() in ("true", "1", "yes"):
                return True
            elif value.lower() in ("false", "0", "no"):
                return False
            try:
                return int(value)
            except ValueError:
                try:
                    return float(value)
                except ValueError:
                    return str(value)

        return default

    def _load_yaml(self) -> Dict[str, Any]:
        """Load YAML configuration file."""
        if not self.config_path.exists():
            return {}

        try:
            with open(self.config_path, "r") as f:
                return yaml.safe_load(f) or {}
        except Exception as e:
            logging.warning(f"Error loading config file: {e}")
            return {}

    def _save_yaml(self, data: Dict[str, Any]) -> bool:
        """Save YAML configuration file."""
        try:
            self.config_path.parent.mkdir(parents=True, exist_ok=True)
            with open(self.config_path, "w") as f:
                yaml.dump(data, f, default_flow_style=False, sort_keys=False)
            return True
        except Exception as e:
            logging.error(f"Error saving config file: {e}")
            return False

    def _get_config_hash(self, data: Dict[str, Any]) -> str:
        """Calculate hash of configuration for change detection."""
        config_str = json.dumps(data, sort_keys=True)
        return hashlib.sha256(config_str.encode()).hexdigest()

    def _migrate_from_json(self, json_path: Path) -> Dict[str, Any]:
        """Migrate existing JSON configuration to YAML format."""
        if not json_path.exists():
            return {}

        try:
            with open(json_path, "r") as f:
                return json.load(f) or {}
        except Exception as e:
            logging.warning(f"Error migrating from JSON: {e}")
            return {}

    def _load_config(self) -> None:
        """Load and validate configuration from file and environment."""
        yaml_data = self._load_yaml()

        json_settings_path = self.config_dir / "gui_settings.json"
        json_data = self._migrate_from_json(json_settings_path)

        merged_data = self._merge_configs(yaml_data, json_data)
        self._apply_env_overrides(merged_data)

        self._config = self._dict_to_config(merged_data)
        self._config_hash = self._get_config_hash(merged_data)

        validation_errors = ConfigValidator.validate(self._config)
        if validation_errors:
            logging.warning(f"Configuration validation errors: {validation_errors}")

    def _merge_configs(self, *configs: Dict[str, Any]) -> Dict[str, Any]:
        """Merge multiple config dictionaries with later configs taking precedence."""
        result = {}
        for config in configs:
            for key, value in config.items():
                if (
                    isinstance(value, dict)
                    and key in result
                    and isinstance(result[key], dict)
                ):
                    result[key] = {**result[key], **value}
                else:
                    result[key] = value
        return result

    def _apply_env_overrides(self, data: Dict[str, Any]) -> None:
        """Apply environment variable overrides to config data."""
        env_mappings = {
            "KUCOIN_API_KEY": ("exchanges", "kucoin_api_key"),
            "KUCOIN_API_SECRET": ("exchanges", "kucoin_api_secret"),
            "KUCOIN_API_PASSPHRASE": ("exchanges", "kucoin_api_passphrase"),
            "BINANCE_API_KEY": ("exchanges", "binance_api_key"),
            "BINANCE_API_SECRET": ("exchanges", "binance_api_secret"),
            "COINBASE_API_KEY": ("exchanges", "coinbase_api_key"),
            "COINBASE_API_SECRET": ("exchanges", "coinbase_api_secret"),
            "EMAIL_ADDRESS": ("notifications", "email_address"),
            "EMAIL_PASSWORD": ("notifications", "email_app_password"),
            "DISCORD_WEBHOOK": ("notifications", "discord_webhook_url"),
            "TELEGRAM_BOT_TOKEN": ("notifications", "telegram_bot_token"),
            "TELEGRAM_CHAT_ID": ("notifications", "telegram_chat_id"),
            "LOG_LEVEL": ("system", "log_level"),
            "DEBUG_MODE": ("system", "debug_mode"),
        }

        for env_key, (section, config_key) in env_mappings.items():
            env_value = self._get_env_var(env_key)
            if env_value is not None:
                if section not in data:
                    data[section] = {}
                data[section][config_key] = env_value

    def _dict_to_config(self, data: Dict[str, Any]) -> PowerTraderConfig:
        """Convert dictionary to PowerTraderConfig dataclass."""
        config_dict = {}

        # Explicitly convert each section to its dataclass
        if "trading" in data and isinstance(data["trading"], dict):
            config_dict["trading"] = TradingConfig(**data["trading"])

        if "notifications" in data and isinstance(data["notifications"], dict):
            config_dict["notifications"] = NotificationConfig(**data["notifications"])

        if "exchanges" in data and isinstance(data["exchanges"], dict):
            config_dict["exchanges"] = ExchangeConfig(**data["exchanges"])

        if "analytics" in data and isinstance(data["analytics"], dict):
            config_dict["analytics"] = AnalyticsConfig(**data["analytics"])

        if "position_sizing" in data and isinstance(data["position_sizing"], dict):
            config_dict["position_sizing"] = PositionSizingConfig(**data["position_sizing"])

        if "correlation" in data and isinstance(data["correlation"], dict):
            config_dict["correlation"] = CorrelationConfig(**data["correlation"])

        if "system" in data and isinstance(data["system"], dict):
            config_dict["system"] = SystemConfig(**data["system"])

        return PowerTraderConfig(**config_dict)

    def _config_to_dict(self) -> Dict[str, Any]:
        """Convert PowerTraderConfig to dictionary."""
        return asdict(self._config, dict_factory=dict)

    def get(self) -> PowerTraderConfig:
        """Get current configuration."""
        return self._config

    def get_value(self, section: str, key: str, default: Any = None) -> Any:
        """Get specific configuration value."""
        config_dict = self._config_to_dict()
        return config_dict.get(section, {}).get(key, default)

    def set_value(self, section: str, key: str, value: Any) -> bool:
        """Set specific configuration value."""
        if not hasattr(self._config, section):
            return False

        section_config = getattr(self._config, section)
        if not hasattr(section_config, key):
            return False

        setattr(section_config, key, value)
        return self.save()

    def set_section(self, section: str, config: Any) -> bool:
        """Set entire configuration section."""
        if not hasattr(self._config, section):
            return False

        setattr(self._config, section, config)
        return self.save()

    def reload(self) -> bool:
        """Reload configuration from file."""
        try:
            self._load_config()
            for callback in self._callbacks:
                callback(self._config)
            return True
        except Exception as e:
            logging.error(f"Error reloading config: {e}")
            return False

    def save(self) -> bool:
        """Save current configuration to file."""
        config_dict = self._config_to_dict()

        new_hash = self._get_config_hash(config_dict)

        if self._config_hash and new_hash == self._config_hash:
            return True

        success = self._save_yaml(config_dict)

        if success:
            self._config_hash = new_hash

            for callback in self._callbacks:
                callback(self._config)

        return success

    def register_callback(self, callback: callable) -> None:
        """Register callback for configuration changes."""
        self._callbacks.append(callback)

    def unregister_callback(self, callback: callable) -> None:
        """Unregister callback for configuration changes."""
        if callback in self._callbacks:
            self._callbacks.remove(callback)

    def start_watcher(self, check_interval: float = 5.0) -> None:
        """Start file watcher for automatic reloading."""
        if self._watcher_running:
            return

        self._watcher_running = True
        self._watcher_thread = threading.Thread(
            target=self._watch_config_file, args=(check_interval,), daemon=True
        )
        self._watcher_thread.start()

    def stop_watcher(self) -> None:
        """Stop file watcher."""
        self._watcher_running = False
        if self._watcher_thread:
            self._watcher_thread.join(timeout=2.0)

    def _watch_config_file(self, check_interval: float) -> None:
        """Watch configuration file for changes."""
        last_modified = (
            self.config_path.stat().st_mtime if self.config_path.exists() else 0
        )

        while self._watcher_running:
            try:
                if self.config_path.exists():
                    current_modified = self.config_path.stat().st_mtime

                    if current_modified != last_modified:
                        self.reload()
                        last_modified = current_modified

                import time

                time.sleep(check_interval)
            except Exception as e:
                logging.error(f"Error watching config file: {e}")
                time.sleep(check_interval)

    def export_to_dict(self) -> Dict[str, Any]:
        """Export configuration as dictionary (for GUI)."""
        return self._config_to_dict()

    def export_to_json(self, pretty: bool = True) -> str:
        """Export configuration as JSON string (for GUI)."""
        config_dict = self._config_to_dict()
        if pretty:
            return json.dumps(config_dict, indent=2, sort_keys=False)
        return json.dumps(config_dict, separators=(",", ":"))

    def create_default_config(self) -> bool:
        """Create default configuration file."""
        self._config = PowerTraderConfig()
        return self.save()

    def validate(self) -> List[str]:
        """Validate current configuration."""
        return ConfigValidator.validate(self._config)


def get_config(config_dir: Optional[str] = None) -> ConfigManager:
    """Get singleton instance of ConfigManager."""
    return ConfigManager(config_dir)


def main():
    """Main function for testing configuration management system."""
    import tempfile

    with tempfile.TemporaryDirectory() as tmpdir:
        config_path = Path(tmpdir)
        config_path.mkdir(exist_ok=True)

        manager = ConfigManager(str(config_path))

        print("Testing Configuration Management System...")

        print("\n--- Current Configuration ---")
        config = manager.get()
        print(f"Trading coins: {config.trading.coins}")
        print(f"Start allocation: {config.trading.start_allocation_pct * 100:.1f}%")
        print(f"Trade start level: {config.trading.trade_start_level}")
        print(f"Notifications enabled: {config.notifications.enabled}")
        print(
            f"Email notifications: {config.notifications.platforms.get('email', False)}"
        )
        print(f"Log level: {config.system.log_level}")

        print("\n--- Validation ---")
        errors = manager.validate()
        if errors:
            print("Validation errors found:")
            for error in errors:
                print(f"  - {error}")
        else:
            print("Configuration is valid!")

        print("\n--- Testing Configuration Updates ---")
        print("Updating trade start level to 4...")
        manager.set_value("trading", "trade_start_level", 4)
        print(f"New value: {manager.get().trading.trade_start_level}")

        print("\nUpdating email address...")
        manager.set_value("notifications", "email_address", "test@example.com")
        print(f"New value: {manager.get().notifications.email_address}")

        print("\n--- Testing Export ---")
        print("\nJSON Export (pretty):")
        print(manager.export_to_json(pretty=True)[:200] + "...")

        print("\nDictionary Export:")
        export_dict = manager.export_to_dict()
        print(f"Trading section: {export_dict.get('trading', {})}")

        print("\n--- Testing Hot Reload ---")
        print("Starting file watcher...")
        manager.start_watcher(check_interval=2.0)

        import time

        time.sleep(3)

        print("\nModifying config file externally...")
        with open(manager.config_path, "r") as f:
            content = f.read()

        new_content = content.replace("trade_start_level: 4", "trade_start_level: 5")
        with open(manager.config_path, "w") as f:
            f.write(new_content)

        print("Waiting for reload...")
        time.sleep(3)

        print(f"Reloaded value: {manager.get().trading.trade_start_level}")

        manager.stop_watcher()

        print("\n--- Configuration Management System Test Complete ---")


if __name__ == "__main__":
    main()
