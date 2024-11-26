// Copyright 2024 Canonical.

package mocks

import (
	"context"

	"github.com/juju/names/v5"

	"github.com/canonical/jimm/v3/internal/common/pagination"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
	ofganames "github.com/canonical/jimm/v3/internal/openfga/names"
	apiparams "github.com/canonical/jimm/v3/pkg/api/params"
	jimmnames "github.com/canonical/jimm/v3/pkg/names"
)

// PermissionService is an implementation of the jimm.PermissionManager interface.
type PermissionService struct {
	AddRelation_               func(ctx context.Context, user *openfga.User, tuples []apiparams.RelationshipTuple) error
	RemoveRelation_            func(ctx context.Context, user *openfga.User, tuples []apiparams.RelationshipTuple) error
	CheckRelation_             func(ctx context.Context, user *openfga.User, tuple apiparams.RelationshipTuple, trace bool) (_ bool, err error)
	ListRelationshipTuples_    func(ctx context.Context, user *openfga.User, tuple apiparams.RelationshipTuple, pageSize int32, continuationToken string) ([]openfga.Tuple, string, error)
	ListObjectRelations_       func(ctx context.Context, user *openfga.User, object string, pageSize int32, continuationToken pagination.EntitlementToken) ([]openfga.Tuple, pagination.EntitlementToken, error)
	ToJAASTag_                 func(ctx context.Context, tag *ofganames.Tag, resolveUUIDs bool) (string, error)
	GrantAuditLogAccess_       func(ctx context.Context, user *openfga.User, targetUserTag names.UserTag) error
	RevokeAuditLogAccess_      func(ctx context.Context, user *openfga.User, targetUserTag names.UserTag) error
	GrantServiceAccountAccess_ func(ctx context.Context, u *openfga.User, svcAccTag jimmnames.ServiceAccountTag, entities []string) error
}

func (j *PermissionService) AddRelation(ctx context.Context, user *openfga.User, tuples []apiparams.RelationshipTuple) error {
	if j.AddRelation_ == nil {
		return errors.E(errors.CodeNotImplemented)
	}
	return j.AddRelation_(ctx, user, tuples)
}

func (j *PermissionService) RemoveRelation(ctx context.Context, user *openfga.User, tuples []apiparams.RelationshipTuple) error {
	if j.RemoveRelation_ == nil {
		return errors.E(errors.CodeNotImplemented)
	}
	return j.RemoveRelation_(ctx, user, tuples)
}

func (j *PermissionService) CheckRelation(ctx context.Context, user *openfga.User, tuple apiparams.RelationshipTuple, trace bool) (_ bool, err error) {
	if j.CheckRelation_ == nil {
		return false, errors.E(errors.CodeNotImplemented)
	}
	return j.CheckRelation_(ctx, user, tuple, trace)
}

func (j *PermissionService) ListRelationshipTuples(ctx context.Context, user *openfga.User, tuple apiparams.RelationshipTuple, pageSize int32, continuationToken string) ([]openfga.Tuple, string, error) {
	if j.ListRelationshipTuples_ == nil {
		return []openfga.Tuple{}, "", errors.E(errors.CodeNotImplemented)
	}
	return j.ListRelationshipTuples_(ctx, user, tuple, pageSize, continuationToken)
}

func (j *PermissionService) ListObjectRelations(ctx context.Context, user *openfga.User, object string, pageSize int32, entitlementToken pagination.EntitlementToken) ([]openfga.Tuple, pagination.EntitlementToken, error) {
	if j.ListObjectRelations_ == nil {
		return []openfga.Tuple{}, pagination.EntitlementToken{}, errors.E(errors.CodeNotImplemented)
	}
	return j.ListObjectRelations_(ctx, user, object, pageSize, entitlementToken)
}

func (j *PermissionService) ToJAASTag(ctx context.Context, tag *ofganames.Tag, resolveUUIDs bool) (string, error) {
	if j.ToJAASTag_ == nil {
		return "", errors.E(errors.CodeNotImplemented)
	}
	return j.ToJAASTag(ctx, tag, resolveUUIDs)
}
func (j *PermissionService) GrantAuditLogAccess(ctx context.Context, user *openfga.User, targetUserTag names.UserTag) error {
	if j.GrantAuditLogAccess_ == nil {
		return errors.E(errors.CodeNotImplemented)
	}
	return j.GrantAuditLogAccess(ctx, user, targetUserTag)
}
func (j *PermissionService) RevokeAuditLogAccess(ctx context.Context, user *openfga.User, targetUserTag names.UserTag) error {
	if j.RevokeAuditLogAccess_ == nil {
		return errors.E(errors.CodeNotImplemented)
	}
	return j.RevokeAuditLogAccess(ctx, user, targetUserTag)
}
func (j *PermissionService) GrantServiceAccountAccess(ctx context.Context, u *openfga.User, svcAccTag jimmnames.ServiceAccountTag, entities []string) error {
	if j.GrantServiceAccountAccess_ == nil {
		return errors.E(errors.CodeNotImplemented)
	}
	return j.GrantServiceAccountAccess(ctx, u, svcAccTag, entities)
}
