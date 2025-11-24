// Copyright 2025 Canonical.

package errors_test

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/juju/juju/rpc"

	"github.com/canonical/jimm/v3/internal/errors"
)

func TestE(t *testing.T) {
	c := qt.New(t)

	err := errors.E("an error happened")
	c.Check(err, qt.ErrorMatches, `an error happened`)
}

func TestEWithInfo(t *testing.T) {
	c := qt.New(t)

	info := map[string]any{"key": "value"}
	err := errors.E("an error happened").WithInfo(info)
	c.Check(err, qt.ErrorMatches, `an error happened`)
	c.Check(errors.ErrorInfo(err), qt.DeepEquals, info)

	err = errors.E("plain-error")
	c.Check(err, qt.ErrorMatches, `plain-error`)
	c.Check(errors.ErrorInfo(err), qt.DeepEquals, map[string]any(nil))
}

func TestErrorCodeWithJujuRPC(t *testing.T) {
	c := qt.New(t)

	err := rpc.RequestError{
		Code: "my-code",
		Info: map[string]any{"key": "value"},
	}
	c.Check(string(errors.ErrorCode(&err)), qt.Equals, "my-code")
	c.Check(errors.ErrorInfo(&err), qt.DeepEquals, map[string]any{"key": "value"})

	wrappedErr := fmt.Errorf("wrapped: %w", &err)
	c.Check(string(errors.ErrorCode(wrappedErr)), qt.Equals, "my-code")
	c.Check(errors.ErrorInfo(wrappedErr), qt.DeepEquals, map[string]any{"key": "value"})
}
