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

func TestNew(t *testing.T) {
	c := qt.New(t)

	err := errors.New("test error")
	c.Check(err, qt.ErrorMatches, `test error`)
	c.Check(err.Message, qt.Equals, "test error")
}

func TestNewf(t *testing.T) {
	c := qt.New(t)

	baseErr := fmt.Errorf("base error")
	err := errors.Newf("wrapped error: %w", baseErr)
	c.Check(err.Error(), qt.Matches, `wrapped error: base error`)
	c.Check(err.Unwrap(), qt.ErrorMatches, `wrapped error: base error`)
}

func TestWithMessagef(t *testing.T) {
	c := qt.New(t)

	baseErr := fmt.Errorf("base error")
	err := errors.Wrap(baseErr).WithMessagef("formatted: %s %d", "test", 42)
	c.Check(err.Error(), qt.Equals, "formatted: test 42")
	c.Check(err.Unwrap(), qt.Equals, baseErr)
}

func TestWrap(t *testing.T) {
	c := qt.New(t)

	baseErr := fmt.Errorf("base error")
	err := errors.Wrap(baseErr)
	c.Check(err.Error(), qt.Equals, "base error")
	c.Check(err.Unwrap(), qt.Equals, baseErr)
}

func TestWithCode(t *testing.T) {
	c := qt.New(t)

	err := errors.New("test error").WithCode(errors.CodeNotFound)
	c.Check(err.Error(), qt.Equals, "test error")
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeNotFound)
}
