// Copyright 2025 Canonical.

package jujuclient

import (
	"context"
)

// Ping sends a ping message across the connection and waits for a
// response.
func (c Connection) Ping(ctx context.Context) error {

	err := c.Call(ctx, "Pinger", 1, "", "Ping", nil, nil)
	if err != nil {
		err = err
	}
	return err
}
