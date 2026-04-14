package enterprise_test

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/enterprise"
)

func TestAccountManager_RBAC(t *testing.T) {
	am := enterprise.NewAccountManager()

	// 1. Viewer User
	viewer := enterprise.User{
		ID:         "u1",
		Role:       "viewer",
		AccountIDs: map[string]bool{"acc1": true},
	}
	am.RegisterUser(viewer)

	// 2. Trader User
	trader := enterprise.User{
		ID:         "u2",
		Role:       "trader",
		AccountIDs: map[string]bool{"acc1": true, "acc2": true},
	}
	am.RegisterUser(trader)

	// 3. Admin User
	admin := enterprise.User{
		ID:         "u3",
		Role:       "admin",
		AccountIDs: make(map[string]bool), // Admin bypasses specific account checks
	}
	am.RegisterUser(admin)

	ctx := context.Background()

	// TEST: Viewer checking portfolio on permitted account
	err := am.CheckPermission(ctx, "u1", "acc1", enterprise.PermViewPortfolio)
	if err != nil {
		t.Errorf("Viewer should view portfolio on acc1: %v", err)
	}

	// TEST: Viewer trying to trade
	err = am.CheckPermission(ctx, "u1", "acc1", enterprise.PermPlaceOrders)
	if err == nil {
		t.Errorf("Viewer should NOT be able to place orders")
	}

	// TEST: Trader trading on permitted account
	err = am.CheckPermission(ctx, "u2", "acc2", enterprise.PermPlaceOrders)
	if err != nil {
		t.Errorf("Trader should place orders on acc2: %v", err)
	}

	// TEST: Trader trading on un-permitted account
	err = am.CheckPermission(ctx, "u2", "acc3", enterprise.PermPlaceOrders)
	if err == nil {
		t.Errorf("Trader should NOT access unassigned account acc3")
	}

	// TEST: Admin bypassing account scoping
	err = am.CheckPermission(ctx, "u3", "acc999", enterprise.PermManageUsers)
	if err != nil {
		t.Errorf("Admin should bypass account checks and manage users: %v", err)
	}

	// TEST: Non-existent user
	err = am.CheckPermission(ctx, "ghost", "acc1", enterprise.PermViewPortfolio)
	if err == nil {
		t.Errorf("Non-existent user should be blocked")
	}
}
