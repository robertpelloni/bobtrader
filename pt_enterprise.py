"""
PowerTrader AI — Enterprise Features (v4.0.0)
================================================
Multi-account management, role-based access control (RBAC),
audit logging, and compliance reporting.

Features:
    1. AccountManager — multi-account portfolio management
    2. RBACManager — role-based access control with permissions
    3. AuditLogger — immutable audit trail for all actions
    4. ComplianceReporter — generate compliance reports

Usage:
    from pt_enterprise import AccountManager, RBACManager, AuditLogger

    accounts = AccountManager()
    accounts.create_account("main_fund", capital=100000)

    rbac = RBACManager()
    rbac.create_user("trader1", role="trader")

    audit = AuditLogger()
    audit.log("trade_executed", user="trader1", details={...})
"""

from __future__ import annotations
import json
import time
import hashlib
import uuid
from dataclasses import dataclass, asdict, field
from datetime import datetime
from typing import List, Dict, Optional
from pathlib import Path
from enum import Enum


# =============================================================================
# RBAC — ROLE-BASED ACCESS CONTROL
# =============================================================================

class Permission(Enum):
    VIEW_DASHBOARD = "view_dashboard"
    VIEW_TRADES = "view_trades"
    EXECUTE_TRADES = "execute_trades"
    MANAGE_ALERTS = "manage_alerts"
    MANAGE_STRATEGIES = "manage_strategies"
    VIEW_ANALYTICS = "view_analytics"
    MANAGE_ACCOUNTS = "manage_accounts"
    MANAGE_USERS = "manage_users"
    VIEW_AUDIT = "view_audit"
    EXPORT_DATA = "export_data"
    ADMIN_ALL = "admin_all"


ROLE_PERMISSIONS = {
    "viewer": [
        Permission.VIEW_DASHBOARD, Permission.VIEW_TRADES, Permission.VIEW_ANALYTICS,
    ],
    "trader": [
        Permission.VIEW_DASHBOARD, Permission.VIEW_TRADES, Permission.VIEW_ANALYTICS,
        Permission.EXECUTE_TRADES, Permission.MANAGE_ALERTS, Permission.MANAGE_STRATEGIES,
    ],
    "manager": [
        Permission.VIEW_DASHBOARD, Permission.VIEW_TRADES, Permission.VIEW_ANALYTICS,
        Permission.EXECUTE_TRADES, Permission.MANAGE_ALERTS, Permission.MANAGE_STRATEGIES,
        Permission.MANAGE_ACCOUNTS, Permission.VIEW_AUDIT, Permission.EXPORT_DATA,
    ],
    "admin": [Permission.ADMIN_ALL],
}


@dataclass
class User:
    user_id: str
    username: str
    role: str
    email: str = ""
    created_at: str = ""
    last_login: str = ""
    is_active: bool = True
    api_key: str = ""


class RBACManager:
    """Role-based access control for multi-user environments."""

    USERS_FILE = Path("enterprise/users.json")

    def __init__(self):
        self.USERS_FILE.parent.mkdir(exist_ok=True)
        self.users: Dict[str, User] = self._load()

    def _load(self) -> Dict[str, User]:
        if self.USERS_FILE.exists():
            try:
                data = json.loads(self.USERS_FILE.read_text())
                return {uid: User(**u) for uid, u in data.items()}
            except Exception:
                pass
        return {}

    def _save(self):
        data = {uid: asdict(u) for uid, u in self.users.items()}
        self.USERS_FILE.write_text(json.dumps(data, indent=2))

    def create_user(self, username: str, role: str = "viewer",
                    email: str = "") -> User:
        """Create a new user with the given role."""
        if role not in ROLE_PERMISSIONS:
            raise ValueError(f"Invalid role: {role}. Valid: {list(ROLE_PERMISSIONS.keys())}")

        user_id = str(uuid.uuid4())[:8]
        api_key = hashlib.sha256(f"{username}{time.time()}".encode()).hexdigest()[:32]

        user = User(
            user_id=user_id,
            username=username,
            role=role,
            email=email,
            created_at=datetime.now().isoformat(),
            api_key=api_key,
        )
        self.users[user_id] = user
        self._save()
        return user

    def check_permission(self, user_id: str, permission: Permission) -> bool:
        """Check if a user has the given permission."""
        user = self.users.get(user_id)
        if not user or not user.is_active:
            return False

        role_perms = ROLE_PERMISSIONS.get(user.role, [])
        return Permission.ADMIN_ALL in role_perms or permission in role_perms

    def authenticate(self, api_key: str) -> Optional[User]:
        """Authenticate user by API key."""
        for user in self.users.values():
            if user.api_key == api_key and user.is_active:
                user.last_login = datetime.now().isoformat()
                self._save()
                return user
        return None

    def list_users(self) -> List[dict]:
        return [asdict(u) for u in self.users.values()]


# =============================================================================
# ACCOUNT MANAGER
# =============================================================================

@dataclass
class TradingAccount:
    account_id: str
    name: str
    capital: float
    available: float
    positions: List[dict] = field(default_factory=list)
    pnl: float = 0.0
    owner_id: str = ""
    created_at: str = ""
    is_active: bool = True


class AccountManager:
    """Multi-account portfolio management."""

    ACCOUNTS_FILE = Path("enterprise/accounts.json")

    def __init__(self):
        self.ACCOUNTS_FILE.parent.mkdir(exist_ok=True)
        self.accounts: Dict[str, TradingAccount] = self._load()

    def _load(self) -> Dict[str, TradingAccount]:
        if self.ACCOUNTS_FILE.exists():
            try:
                data = json.loads(self.ACCOUNTS_FILE.read_text())
                return {aid: TradingAccount(**a) for aid, a in data.items()}
            except Exception:
                pass
        return {}

    def _save(self):
        data = {aid: asdict(a) for aid, a in self.accounts.items()}
        self.ACCOUNTS_FILE.write_text(json.dumps(data, indent=2))

    def create_account(self, name: str, capital: float,
                       owner_id: str = "") -> TradingAccount:
        """Create a new trading account."""
        account_id = f"acct_{str(uuid.uuid4())[:8]}"
        account = TradingAccount(
            account_id=account_id,
            name=name,
            capital=capital,
            available=capital,
            owner_id=owner_id,
            created_at=datetime.now().isoformat(),
        )
        self.accounts[account_id] = account
        self._save()
        return account

    def get_account(self, account_id: str) -> Optional[TradingAccount]:
        return self.accounts.get(account_id)

    def get_all_accounts_summary(self) -> dict:
        """Summary across all accounts."""
        total_capital = sum(a.capital for a in self.accounts.values() if a.is_active)
        total_pnl = sum(a.pnl for a in self.accounts.values() if a.is_active)
        total_positions = sum(len(a.positions) for a in self.accounts.values() if a.is_active)

        return {
            "total_accounts": len([a for a in self.accounts.values() if a.is_active]),
            "total_capital": round(total_capital, 2),
            "total_pnl": round(total_pnl, 2),
            "total_positions": total_positions,
            "accounts": [asdict(a) for a in self.accounts.values() if a.is_active],
        }


# =============================================================================
# AUDIT LOGGER
# =============================================================================

class AuditLogger:
    """Immutable audit trail for all system actions."""

    AUDIT_FILE = Path("enterprise/audit_log.jsonl")

    def __init__(self):
        self.AUDIT_FILE.parent.mkdir(exist_ok=True)

    def log(self, action: str, user: str = "system",
            details: Optional[dict] = None, severity: str = "info"):
        """Log an audit event (append-only)."""
        entry = {
            "id": str(uuid.uuid4())[:12],
            "timestamp": datetime.now().isoformat(),
            "action": action,
            "user": user,
            "severity": severity,
            "details": details or {},
            "checksum": "",
        }

        # Create integrity checksum
        content = json.dumps({k: v for k, v in entry.items() if k != "checksum"}, sort_keys=True)
        entry["checksum"] = hashlib.sha256(content.encode()).hexdigest()[:16]

        with open(self.AUDIT_FILE, "a") as f:
            f.write(json.dumps(entry) + "\n")

    def get_recent(self, n: int = 50, user: str = "",
                   action: str = "") -> List[dict]:
        """Get recent audit entries with optional filters."""
        if not self.AUDIT_FILE.exists():
            return []

        entries = []
        for line in self.AUDIT_FILE.read_text().strip().split("\n"):
            if not line:
                continue
            try:
                entry = json.loads(line)
                if user and entry.get("user") != user:
                    continue
                if action and entry.get("action") != action:
                    continue
                entries.append(entry)
            except Exception:
                pass

        return entries[-n:]

    def verify_integrity(self) -> dict:
        """Verify checksum integrity of audit log."""
        if not self.AUDIT_FILE.exists():
            return {"status": "empty", "entries": 0}

        total = 0
        valid = 0
        tampered = []

        for line in self.AUDIT_FILE.read_text().strip().split("\n"):
            if not line:
                continue
            try:
                entry = json.loads(line)
                total += 1
                content = json.dumps({k: v for k, v in entry.items() if k != "checksum"}, sort_keys=True)
                computed = hashlib.sha256(content.encode()).hexdigest()[:16]
                if computed == entry.get("checksum"):
                    valid += 1
                else:
                    tampered.append(entry.get("id", "?"))
            except Exception:
                total += 1

        return {
            "status": "valid" if total == valid else "tampered",
            "total_entries": total,
            "valid_entries": valid,
            "tampered_ids": tampered,
        }


# =============================================================================
# COMPLIANCE REPORTER
# =============================================================================

class ComplianceReporter:
    """Generate compliance reports from audit logs and trade data."""

    def __init__(self, audit_logger: Optional[AuditLogger] = None):
        self.audit = audit_logger or AuditLogger()

    def generate_report(self, period_days: int = 30) -> dict:
        """Generate a compliance report for the given period."""
        recent = self.audit.get_recent(n=10000)
        integrity = self.audit.verify_integrity()

        # Categorize actions
        trades = [e for e in recent if "trade" in e.get("action", "")]
        logins = [e for e in recent if "login" in e.get("action", "")]
        admin_actions = [e for e in recent if e.get("severity") in ("warning", "critical")]

        unique_users = set(e.get("user", "") for e in recent)

        return {
            "report_title": "PowerTrader AI Compliance Report",
            "generated_at": datetime.now().isoformat(),
            "period_days": period_days,
            "summary": {
                "total_events": len(recent),
                "trade_events": len(trades),
                "login_events": len(logins),
                "admin_events": len(admin_actions),
                "unique_users": len(unique_users),
            },
            "audit_integrity": integrity,
            "risk_flags": self._detect_risk_flags(recent),
            "recommendations": self._generate_recommendations(recent),
        }

    def _detect_risk_flags(self, events: List[dict]) -> List[dict]:
        """Detect potential compliance risks."""
        flags = []

        # Check for unauthorized access attempts
        failed = [e for e in events if "unauthorized" in e.get("action", "")]
        if failed:
            flags.append({"type": "unauthorized_access", "count": len(failed), "severity": "high"})

        # Check for unusual trading patterns
        trade_events = [e for e in events if "trade" in e.get("action", "")]
        if len(trade_events) > 100:
            flags.append({"type": "high_trade_volume", "count": len(trade_events), "severity": "medium"})

        return flags

    def _generate_recommendations(self, events: List[dict]) -> List[str]:
        """Generate compliance recommendations."""
        recs = []
        if not events:
            recs.append("Enable audit logging for all trade actions")
        recs.append("Review access permissions quarterly")
        recs.append("Ensure all API keys are rotated every 90 days")
        recs.append("Monitor for unusual trading patterns")
        return recs


# =============================================================================
# SELF-TEST
# =============================================================================

if __name__ == "__main__":
    print("=" * 60)
    print("Enterprise Features — Self-Test")
    print("=" * 60)

    # 1. RBAC
    print("\n1. RBACManager...")
    rbac = RBACManager()
    admin = rbac.create_user("admin_user", "admin", "admin@trading.com")
    trader = rbac.create_user("trader1", "trader", "t1@trading.com")
    viewer = rbac.create_user("viewer1", "viewer")
    print(f"   Created: admin={admin.user_id}, trader={trader.user_id}, viewer={viewer.user_id}")
    print(f"   Admin can manage_users: {rbac.check_permission(admin.user_id, Permission.MANAGE_USERS)}")
    print(f"   Trader can execute: {rbac.check_permission(trader.user_id, Permission.EXECUTE_TRADES)}")
    print(f"   Viewer can execute: {rbac.check_permission(viewer.user_id, Permission.EXECUTE_TRADES)}")

    # 2. Account Manager
    print("\n2. AccountManager...")
    mgr = AccountManager()
    acct1 = mgr.create_account("Main Fund", 100000, admin.user_id)
    acct2 = mgr.create_account("Algo Fund", 50000, trader.user_id)
    summary = mgr.get_all_accounts_summary()
    print(f"   Accounts: {summary['total_accounts']}")
    print(f"   Total Capital: ${summary['total_capital']:,.2f}")

    # 3. Audit Logger
    print("\n3. AuditLogger...")
    audit = AuditLogger()
    audit.log("user_created", admin.username, {"new_user": trader.username})
    audit.log("trade_executed", trader.username, {"coin": "BTC", "side": "buy", "qty": 0.1})
    audit.log("login", viewer.username, {"ip": "192.168.1.1"})
    recent = audit.get_recent(5)
    print(f"   Recent entries: {len(recent)}")
    integrity = audit.verify_integrity()
    print(f"   Integrity: {integrity['status']} ({integrity['valid_entries']}/{integrity['total_entries']})")

    # 4. Compliance
    print("\n4. ComplianceReporter...")
    reporter = ComplianceReporter(audit)
    report = reporter.generate_report()
    print(f"   Events: {report['summary']['total_events']}")
    print(f"   Risk flags: {len(report['risk_flags'])}")
    print(f"   Recommendations: {len(report['recommendations'])}")

    print("\n✅ Self-test complete")
