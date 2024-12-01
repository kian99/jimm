// Copyright 2024 Canonical.

package jimm

import (
	"context"
	"fmt"

	"github.com/juju/names/v5"

	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
	ofganames "github.com/canonical/jimm/v3/internal/openfga/names"
)

// CheckPermission loops over the desired permissions in desiredPerms and adds these permissions
// to cachedPerms if they exist. If the user does not have any of the desired permissions then an
// error is returned.
// Note that cachedPerms map is modified and returned.
func (j *JIMM) CheckPermission(ctx context.Context, user *openfga.User, cachedPerms map[string]string, desiredPerms map[string]interface{}) (map[string]string, error) {
	const op = errors.Op("jimm.CheckPermission")
	for key, val := range desiredPerms {
		if _, ok := cachedPerms[key]; !ok {
			stringVal, ok := val.(string)
			if !ok {
				return nil, errors.E(op, fmt.Sprintf("failed to get permission assertion: expected %T, got %T", stringVal, val))
			}
			tag, err := names.ParseTag(key)
			if err != nil {
				return cachedPerms, errors.E(op, fmt.Sprintf("failed to parse tag %s", key))
			}
			relation, err := ofganames.ToOpenFGARelation(stringVal)
			if err != nil {
				return cachedPerms, errors.E(op, fmt.Sprintf("failed to parse relation %s", stringVal), err)
			}
			check, err := openfga.CheckRelation(ctx, user, tag, relation)
			if err != nil {
				return cachedPerms, errors.E(op, err)
			}
			if !check {
				return cachedPerms, errors.E(op, fmt.Sprintf("Missing permission for %s:%s", key, val))
			}
			cachedPerms[key] = stringVal
		}
	}
	return cachedPerms, nil
}

// CheckBatchPermissions loops over the desired permissions in desiredPerms and checks all the permissions.
// The resulting map will only contain entries for the desired permissions that the user can access.
func (j *JIMM) CheckBatchPermissions(ctx context.Context, user *openfga.User, desiredPerms map[names.Tag]openfga.Relation) (map[names.Tag]openfga.Relation, error) {
	permissions := make(map[names.Tag]openfga.Relation)
	const op = errors.Op("jimm.CheckPermission")
	for tag, relation := range desiredPerms {
		allowed, err := openfga.CheckRelation(ctx, user, tag, relation)
		if err != nil {
			return nil, errors.E(op, err)
		}
		if allowed {
			permissions[tag] = relation
		}
	}
	return permissions, nil
}
