package enterprise

import (
	"context"
	"fmt"
	"sync"
)

// Permission defines a specific action that can be performed.
type Permission string

const (
	PermViewPortfolio  Permission = "view_portfolio"
	PermViewAnalytics  Permission = "view_analytics"
	PermPlaceOrders    Permission = "place_orders"
	PermCancelOrders   Permission = "cancel_orders"
	PermManageStrategy Permission = "manage_strategy"
	PermManageRisk     Permission = "manage_risk"
	PermManageAccounts Permission = "manage_accounts"
	PermManageUsers    Permission = "manage_users"
	PermViewAuditLogs  Permission = "view_audit_logs"
	PermManageSettings Permission = "manage_settings"
	PermBypassLimits   Permission = "bypass_limits"
)

// Role defines a named collection of permissions.
type Role struct {
	Name        string
	Permissions map[Permission]bool
}

// User represents a system operator with specific access to a subset of accounts.
type User struct {
	ID         string
	Username   string
	Role       string
	AccountIDs map[string]bool // Which trading accounts this user can access
}

// AccountManager handles users, roles, and authorization checks.
type AccountManager struct {
	mu    sync.RWMutex
	roles map[string]Role
	users map[string]User
}

// NewAccountManager creates a new instance with default roles.
func NewAccountManager() *AccountManager {
	am := &AccountManager{
		roles: make(map[string]Role),
		users: make(map[string]User),
	}
	am.initDefaultRoles()
	return am
}

func (am *AccountManager) initDefaultRoles() {
	am.roles["viewer"] = Role{
		Name: "viewer",
		Permissions: map[Permission]bool{
			PermViewPortfolio: true,
			PermViewAnalytics: true,
		},
	}

	am.roles["trader"] = Role{
		Name: "trader",
		Permissions: map[Permission]bool{
			PermViewPortfolio: true,
			PermViewAnalytics: true,
			PermPlaceOrders:   true,
			PermCancelOrders:  true,
		},
	}

	am.roles["manager"] = Role{
		Name: "manager",
		Permissions: map[Permission]bool{
			PermViewPortfolio:  true,
			PermViewAnalytics:  true,
			PermPlaceOrders:    true,
			PermCancelOrders:   true,
			PermManageStrategy: true,
			PermManageRisk:     true,
			PermViewAuditLogs:  true,
		},
	}

	am.roles["admin"] = Role{
		Name: "admin",
		Permissions: map[Permission]bool{
			PermViewPortfolio:  true,
			PermViewAnalytics:  true,
			PermPlaceOrders:    true,
			PermCancelOrders:   true,
			PermManageStrategy: true,
			PermManageRisk:     true,
			PermManageAccounts: true,
			PermManageUsers:    true,
			PermViewAuditLogs:  true,
			PermManageSettings: true,
			PermBypassLimits:   true,
		},
	}
}

// RegisterUser adds or updates a user in the system.
func (am *AccountManager) RegisterUser(user User) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.roles[user.Role]; !exists {
		return fmt.Errorf("role %q does not exist", user.Role)
	}

	am.users[user.ID] = user
	return nil
}

// CheckPermission verifies if a user has the right to perform an action on a specific account.
func (am *AccountManager) CheckPermission(ctx context.Context, userID, accountID string, perm Permission) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	user, exists := am.users[userID]
	if !exists {
		return fmt.Errorf("access denied: user %q not found", userID)
	}

	// 1. Account Level Access
	// Admins bypass account scoping generally, but others must be explicitly granted access to the account
	if user.Role != "admin" {
		if !user.AccountIDs[accountID] {
			return fmt.Errorf("access denied: user %q does not have access to account %q", userID, accountID)
		}
	}

	// 2. Role Level Permission
	role, exists := am.roles[user.Role]
	if !exists {
		return fmt.Errorf("access denied: invalid role %q configured for user %q", user.Role, userID)
	}

	if !role.Permissions[perm] {
		return fmt.Errorf("access denied: user %q (role %q) lacks permission %q", userID, user.Role, perm)
	}

	return nil
}
