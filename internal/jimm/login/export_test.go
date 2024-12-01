package login

import (
	"context"

	"github.com/canonical/jimm/v3/internal/openfga"
)

// Login is a type alias to export loginManager for use in tests.
type LoginManager = loginManager

func (j *LoginManager) GetUser(ctx context.Context, identifier string) (*openfga.User, error) {
	return j.getUser(ctx, identifier)
}

func (j *LoginManager) UpdateUserLastLogin(ctx context.Context, identifier string) error {
	return j.updateUserLastLogin(ctx, identifier)
}
