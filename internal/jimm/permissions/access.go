// Copyright 2024 Canonical.

package permissions

import (
	"context"

	"github.com/canonical/ofga"
	"github.com/juju/names/v5"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
	ofganames "github.com/canonical/jimm/v3/internal/openfga/names"
	jimmnames "github.com/canonical/jimm/v3/pkg/names"
)

// GrantAuditLogAccess grants audit log access for the target user.
func (j *permissionManager) GrantAuditLogAccess(ctx context.Context, user *openfga.User, targetUserTag names.UserTag) error {
	const op = errors.Op("jimm.GrantAuditLogAccess")

	access := user.GetControllerAccess(ctx, j.tag)
	if access != ofganames.AdministratorRelation {
		return errors.E(op, errors.CodeUnauthorized, "unauthorized")
	}

	targetUser := &dbmodel.Identity{}
	targetUser.SetTag(targetUserTag)
	err := j.store.GetIdentity(ctx, targetUser)
	if err != nil {
		return errors.E(op, err)
	}

	err = openfga.NewUser(targetUser, j.authSvc).SetControllerAccess(ctx, j.tag, ofganames.AuditLogViewerRelation)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// RevokeAuditLogAccess revokes audit log access for the target user.
func (j *permissionManager) RevokeAuditLogAccess(ctx context.Context, user *openfga.User, targetUserTag names.UserTag) error {
	const op = errors.Op("jimm.RevokeAuditLogAccess")

	access := user.GetControllerAccess(ctx, j.tag)
	if access != ofganames.AdministratorRelation {
		return errors.E(op, errors.CodeUnauthorized, "unauthorized")
	}

	targetUser := &dbmodel.Identity{}
	targetUser.SetTag(targetUserTag)
	err := j.store.GetIdentity(ctx, targetUser)
	if err != nil {
		return errors.E(op, err)
	}

	err = openfga.NewUser(targetUser, j.authSvc).UnsetAuditLogViewerAccess(ctx, j.tag)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// GrantServiceAccountAccess creates an administrator relation between the tags provided
// and the service account. The provided tags must be users or groups (with the member relation)
// otherwise OpenFGA will report an error.
func (j *permissionManager) GrantServiceAccountAccess(ctx context.Context, u *openfga.User, svcAccTag jimmnames.ServiceAccountTag, entities []string) error {
	op := errors.Op("jimm.GrantServiceAccountAccess")
	tags := make([]*ofganames.Tag, 0, len(entities))
	// Validate tags
	for _, val := range entities {
		tag, err := j.parseAndValidateTag(ctx, val)
		if err != nil {
			return errors.E(op, err)
		}
		if tag.Kind != openfga.UserType && tag.Kind != openfga.GroupType {
			return errors.E(op, "invalid entity - not user or group")
		}
		if tag.Kind == openfga.GroupType {
			tag.Relation = ofganames.MemberRelation
		}
		tags = append(tags, tag)
	}
	tuples := make([]openfga.Tuple, 0, len(tags))
	svcAccEntity := ofganames.ConvertTag(svcAccTag)
	for _, tag := range tags {
		tuple := openfga.Tuple{
			Object:   tag,
			Relation: ofganames.AdministratorRelation,
			Target:   svcAccEntity,
		}
		tuples = append(tuples, tuple)
	}
	err := j.authSvc.AddRelation(ctx, tuples...)
	if err != nil {
		zapctx.Error(ctx, "failed to add tuple(s)", zap.NamedError("add-relation-error", err))
		return errors.E(op, errors.CodeOpenFGARequestFailed, err)
	}
	return nil
}

// GetUserControllerAccess returns the user's level of access to the desired controller.
func (j *permissionManager) GetUserControllerAccess(ctx context.Context, user *openfga.User, controller names.ControllerTag) (openfga.Relation, error) {
	return user.GetControllerAccess(ctx, controller), nil
}

// GetUserModelAccess returns the access level a user has against a specific model.
func (j *permissionManager) GetUserModelAccess(ctx context.Context, user *openfga.User, model names.ModelTag) (openfga.Relation, error) {
	return user.GetModelAccess(ctx, model), nil
}

// GetUserCloudAccess returns users access level for the specified cloud.
func (j *permissionManager) GetUserCloudAccess(ctx context.Context, user *openfga.User, cloud names.CloudTag) (openfga.Relation, error) {
	return user.GetCloudAccess(ctx, cloud), nil
}

// OpenFGACleanup queries OpenFGA for all existing tuples, tries to resolve each tuple and removes those
// that JIMM cannot resolved - orphaned tuples. JIMM not being able to resolve a tuple means that the
// corresponding entity has been removed from JIMM's database.
//
// This approach to cleaning up tuples is intended to be temporary while we implement
// a better approach to eventual consistency of JIMM's database objects and OpenFGA tuples.
func (j *permissionManager) OpenFGACleanup(ctx context.Context) error {
	var (
		continuationToken string
		err               error
		tuples            []ofga.Tuple
	)
	for {
		tuples, continuationToken, err = j.authSvc.ReadRelatedObjects(ctx, openfga.Tuple{}, 20, continuationToken)
		if err != nil {
			zapctx.Error(ctx, "reading all tuples", zap.Error(err))
			return err
		}

		orphanedTuples := j.orphanedTuples(ctx, tuples...)
		if len(orphanedTuples) > 0 {
			zapctx.Debug(ctx, "removing orphaned tuples", zap.Any("tuples", orphanedTuples))
			err = j.authSvc.RemoveRelation(ctx, orphanedTuples...)
			if err != nil {
				zapctx.Warn(ctx, "failed to clean up orphaned tuples", zap.Error(err))
			}
		}
		if continuationToken == "" {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
}

func (j *permissionManager) orphanedTuples(ctx context.Context, tuples ...openfga.Tuple) []openfga.Tuple {
	orphanedTuples := []openfga.Tuple{}
	for _, tuple := range tuples {
		_, err := j.ToJAASTag(ctx, tuple.Object, true)
		if err != nil {
			if errors.ErrorCode(err) == errors.CodeNotFound {
				orphanedTuples = append(orphanedTuples, tuple)
				continue
			}
		}
		_, err = j.ToJAASTag(ctx, tuple.Target, true)
		if err != nil {
			if errors.ErrorCode(err) == errors.CodeNotFound {
				orphanedTuples = append(orphanedTuples, tuple)
				continue
			}
		}
	}
	return orphanedTuples
}
