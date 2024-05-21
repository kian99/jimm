// Copyright 2024 Canonical Ltd.

package jimm

import (
	"context"

	"github.com/canonical/jimm/internal/errors"
	"github.com/canonical/jimm/internal/openfga"
	ofganames "github.com/canonical/jimm/internal/openfga/names"
	jimmnames "github.com/canonical/jimm/pkg/names"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
)

// AddServiceAccount checks that no one owns the service account yet
// and then adds a relation between the logged in user and the service account.
func (j *JIMM) AddServiceAccount(ctx context.Context, u *openfga.User, clientId string) error {
	op := errors.Op("jimm.AddServiceAccount")

	svcTag := jimmnames.NewServiceAccountTag(clientId)
	key := openfga.Tuple{
		Relation: ofganames.AdministratorRelation,
		Target:   ofganames.ConvertTag(svcTag),
	}
	keyWithUser := key
	keyWithUser.Object = ofganames.ConvertTag(u.ResourceTag())

	ok, err := j.OpenFGAClient.CheckRelation(ctx, keyWithUser, false)
	if err != nil {
		return errors.E(op, err)
	}
	// If the user already has administration permission over the
	// service account then return early.
	if ok {
		return nil
	}

	tuples, _, err := j.OpenFGAClient.ReadRelatedObjects(ctx, key, 10, "")
	if err != nil {
		return errors.E(op, err)
	}
	if len(tuples) > 0 {
		return errors.E(op, "service account already owned")
	}
	addTuple := openfga.Tuple{
		Object:   ofganames.ConvertTag(u.ResourceTag()),
		Relation: ofganames.AdministratorRelation,
		Target:   ofganames.ConvertTag(svcTag),
	}
	err = j.OpenFGAClient.AddRelation(ctx, addTuple)
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// GrantServiceAccountAccess creates an administrator relation between the tags provided
// and the service account. The provided tags must be users or groups (with the member relation)
// otherwise OpenFGA will report an error.
func (j *JIMM) GrantServiceAccountAccess(ctx context.Context, u *openfga.User, svcAccTag jimmnames.ServiceAccountTag, entities []string) error {
	op := errors.Op("jimm.GrantServiceAccountAccess")
	tags := make([]*ofganames.Tag, 0, len(entities))
	// Validate tags
	for _, val := range entities {
		tag, err := j.ParseTag(ctx, val)
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
	err := j.AuthorizationClient().AddRelation(ctx, tuples...)
	if err != nil {
		zapctx.Error(ctx, "failed to add tuple(s)", zap.NamedError("add-relation-error", err))
		return errors.E(op, errors.CodeOpenFGARequestFailed, err)
	}
	return nil
}
