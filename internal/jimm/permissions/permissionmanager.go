package permissions

import (
	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
	"github.com/juju/names/v5"
)

// permissionManager provides a means to manage roles within JIMM.
type permissionManager struct {
	store   *db.Database
	authSvc *openfga.OFGAClient
	// JIMM's UUID and tag
	uuid string
	tag  names.ControllerTag
}

// NewPermissionManager returns a new permission manager that provides permission
// checks, additions, and removals
func NewPermissionManager(store *db.Database, authSvc *openfga.OFGAClient, uuid string, tag names.ControllerTag) (*permissionManager, error) {
	if store == nil {
		return nil, errors.E("role store cannot be nil")
	}
	if authSvc == nil {
		return nil, errors.E("role authorisation service cannot be nil")
	}
	return &permissionManager{store, authSvc, uuid, tag}, nil
}
