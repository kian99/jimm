// Copyright 2025 Canonical.

package jujuclient

import (
	"context"

	"github.com/canonical/jimm/v3/internal/errors"
)

// ControllerHostKey retrieves the public SSH host key for the controller.
func (c Connection) ControllerHostKey(ctx context.Context) ([]byte, error) {
	return nil, errors.E(errors.CodeNotImplemented, "ControllerHostKey not implemented")
}
