// Copyright 2025 Canonical.

package role

import (
	"context"

	"github.com/canonical/jimm/v3/internal/common/pagination"
	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
)

// roleManager provides a means to manage roles within JIMM.
type roleManager struct {
	store   *db.Database
	authSvc *openfga.OFGAClient
}

// NewRoleManager returns a new RoleManager that persists the roles in the provided store.
func NewRoleManager(store *db.Database, authSvc *openfga.OFGAClient) (*roleManager, error) {
	if store == nil {
		return nil, errors.New("role store cannot be nil")
	}
	if authSvc == nil {
		return nil, errors.New("role authorisation service cannot be nil")
	}
	return &roleManager{store, authSvc}, nil
}

// AddRole adds a role to JIMM.
func (rm *roleManager) AddRole(ctx context.Context, user *openfga.User, roleName string) (*dbmodel.RoleEntry, error) {
	if !user.JimmAdmin {
		return nil, errors.New("unauthorized").WithCode(errors.CodeUnauthorized)
	}

	re, err := rm.store.AddRole(ctx, roleName)
	if err != nil {
		return nil, err
	}
	return re, nil
}

// GetRoleByUUID returns a role based on the provided UUID.
func (rm *roleManager) GetRoleByUUID(ctx context.Context, user *openfga.User, uuid string) (*dbmodel.RoleEntry, error) {
	return rm.getRole(ctx, user, &dbmodel.RoleEntry{UUID: uuid})
}

// GetRoleByName returns a role based on the provided name.
func (rm *roleManager) GetRoleByName(ctx context.Context, user *openfga.User, name string) (*dbmodel.RoleEntry, error) {
	return rm.getRole(ctx, user, &dbmodel.RoleEntry{Name: name})
}

// RemoveRole removes the role from JIMM in both the store and authorisation store.
func (rm *roleManager) RemoveRole(ctx context.Context, user *openfga.User, roleName string) error {
	if !user.JimmAdmin {
		return errors.New("unauthorized").WithCode(errors.CodeUnauthorized)
	}

	re := &dbmodel.RoleEntry{
		Name: roleName,
	}
	err := rm.store.GetRole(ctx, re)
	if err != nil {
		return err
	}

	// TODO(ale8k):
	// Would be nice to have a way to create a transaction to get, remove tuples, if successful, delete role
	// somehow. We could pass a callback and change the db methods?
	if err := rm.authSvc.RemoveRole(ctx, re.ResourceTag()); err != nil {
		return err
	}

	if err := rm.store.RemoveRole(ctx, re); err != nil {
		return err
	}

	return nil
}

// RenameRole renames a role in JIMM's DB.
func (rm *roleManager) RenameRole(ctx context.Context, user *openfga.User, oldName, newName string) error {
	if !user.JimmAdmin {
		return errors.New("unauthorized").WithCode(errors.CodeUnauthorized)
	}

	err := rm.store.UpdateRoleName(ctx, oldName, newName)
	if err != nil {
		return err
	}

	return nil
}

// ListRoles returns a list of roles known to JIMM.
// `match` will filter the list fuzzy matching role's name or uuid.
func (rm *roleManager) ListRoles(ctx context.Context, user *openfga.User, pagination pagination.LimitOffsetPagination, match string) ([]dbmodel.RoleEntry, error) {
	if !user.JimmAdmin {
		return nil, errors.New("unauthorized").WithCode(errors.CodeUnauthorized)
	}

	res, err := rm.store.ListRoles(ctx, pagination.Limit(), pagination.Offset(), match)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// CountRoles returns the number of roles that exist.
func (rm *roleManager) CountRoles(ctx context.Context, user *openfga.User) (int, error) {
	if !user.JimmAdmin {
		return 0, errors.New("unauthorized").WithCode(errors.CodeUnauthorized)
	}
	count, err := rm.store.CountRoles(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// getRole returns a role based on the provided UUID or name.
func (rm *roleManager) getRole(ctx context.Context, user *openfga.User, role *dbmodel.RoleEntry) (*dbmodel.RoleEntry, error) {
	if !user.JimmAdmin {
		return nil, errors.New("unauthorized").WithCode(errors.CodeUnauthorized)
	}

	if err := rm.store.GetRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}
