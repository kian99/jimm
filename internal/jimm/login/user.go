package login

import (
	"context"
	"database/sql"

	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
)

// UserLogin fetches a user based on their identityName and updates their last login time.
func (j *loginManager) UserLogin(ctx context.Context, identityName string) (*openfga.User, error) {
	const op = errors.Op("jimm.UserLogin")
	user, err := j.getUser(ctx, identityName)
	if err != nil {
		return nil, errors.E(op, err, errors.CodeUnauthorized)
	}
	err = j.updateUserLastLogin(ctx, identityName)
	if err != nil {
		return nil, errors.E(op, err, errors.CodeUnauthorized)
	}
	return user, nil
}

// getUser fetches the user specified by the user's email or the service accounts ID
// and returns an openfga User that can be used to verify user's permissions.
func (j *loginManager) getUser(ctx context.Context, identifier string) (*openfga.User, error) {
	const op = errors.Op("jimm.GetUser")

	user, err := dbmodel.NewIdentity(identifier)
	if err != nil {
		return nil, errors.E(op, err)
	}

	if err := j.store.GetIdentity(ctx, user); err != nil {
		return nil, err
	}
	u := openfga.NewUser(user, j.authSvc)

	isJimmAdmin, err := openfga.IsAdministrator(ctx, u, j.jimmTag)
	if err != nil {
		return nil, errors.E(op, err)
	}
	u.JimmAdmin = isJimmAdmin

	return u, nil
}

// updateUserLastLogin updates the user's last login time in the database.
func (j *loginManager) updateUserLastLogin(ctx context.Context, identifier string) error {
	const op = errors.Op("jimm.UpdateUserLastLogin")
	user, err := dbmodel.NewIdentity(identifier)
	if err != nil {
		return err
	}
	if err := j.store.Transaction(func(tx *db.Database) error {
		if err := tx.GetIdentity(ctx, user); err != nil {
			return err
		}
		user.LastLogin = sql.NullTime{
			Time:  j.store.DB.Config.NowFunc(),
			Valid: true,
		}
		return tx.UpdateIdentity(ctx, user)
	}); err != nil {
		return errors.E(op, err)
	}
	return nil
}
