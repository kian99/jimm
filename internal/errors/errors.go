// Copyright 2025 Canonical.

// Package errors contains types to help handle errors in the system.
package errors

import (
	stderr "errors"
	"fmt"

	jujuparams "github.com/juju/juju/rpc/params"

	apiparams "github.com/canonical/jimm/v3/pkg/api/params"
)

// An Error is an error in the JIMM system.
type Error struct {
	// Code is a code attached to the error.
	Code Code

	// Message is a human-readable error description.
	Message string

	// Info is additional information about the error
	// that clients need to use.
	Info map[string]any

	// Err contains the underlying error, if there is one.
	Err error
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	if e.Code != "" {
		return string(e.Code)
	}
	return "unknown error"
}

// Unwrap implements the Unwrap method used by errors.Unwrap.
func (e Error) Unwrap() error {
	return e.Err
}

// ErrorCode returns the value of this error's Code.
func (e Error) ErrorCode() string {
	return string(e.Code)
}

// ErrorInfo returns the value of this error's Info.
func (e Error) ErrorInfo() map[string]any {
	return e.Info
}

// E constructs errors for use throughout the JIMM application.
// The initial construction of the error only accepts a string.
// Modify the error further using the methods on the returned
// object, including `Wrap`, `WithCode`, `WithInfo`, etc.
func E(msg string) Error {
	return Error{Message: msg}
}

// New creates a new Error with the given message.
func New(msg string) Error {
	return Error{Message: msg}
}

// Newf creates a new Error using fmt.Errorf for formatting.
// This allows the use of %w verb for error wrapping.
func Newf(format string, args ...any) Error {
	return Error{Err: fmt.Errorf(format, args...)}
}

func Wrap(err error) Error {
	return Error{Err: err}
}

func (e Error) Wrap(err error) Error {
	e.Err = err
	return e
}

func (e Error) WithCode(code Code) Error {
	e.Code = code
	return e
}

func (e Error) WithInfo(info map[string]any) Error {
	e.Info = info
	return e
}

func (e Error) WithMessage(msg string) Error {
	e.Message = msg
	return e
}

// WithMessagef sets a formatted message on the error.
func (e Error) WithMessagef(format string, args ...any) Error {
	e.Message = fmt.Sprintf(format, args...)
	return e
}

// A Code is a code which describes the class of error. Where possible
// these codes are identical to the codes returned in the juju API.
type Code string

const (
	CodeAlreadyExists                Code = jujuparams.CodeAlreadyExists
	CodeBadRequest                   Code = jujuparams.CodeBadRequest
	CodeCloudRegionRequired          Code = jujuparams.CodeCloudRegionRequired
	CodeConnectionFailed             Code = "connection failed"
	CodeDatabaseLocked               Code = "database locked"
	CodeForbidden                    Code = jujuparams.CodeForbidden
	CodeIncompatibleClouds           Code = jujuparams.CodeIncompatibleClouds
	CodeModelNotFound                Code = jujuparams.CodeModelNotFound
	CodeModelMigrating               Code = "model migrating"
	CodeNotFound                     Code = jujuparams.CodeNotFound
	CodeNotImplemented               Code = jujuparams.CodeNotImplemented
	CodeNotSupported                 Code = jujuparams.CodeNotSupported
	CodeRedirect                     Code = jujuparams.CodeRedirect
	CodeServerConfiguration          Code = "server configuration"
	CodeStillAlive                   Code = apiparams.CodeStillAlive
	CodeUnauthorized                 Code = jujuparams.CodeUnauthorized
	CodeServerError                  Code = "server error"
	CodeSessionTokenInvalid          Code = jujuparams.CodeSessionTokenInvalid
	CodeUpgradeInProgress            Code = jujuparams.CodeUpgradeInProgress
	CodeFailedToParseTupleKey        Code = "failed to parse tuple"
	CodeFailedToResolveTupleResource Code = "failed resolve resource"
	CodeOpenFGARequestFailed         Code = "failed request to OpenFGA"
	CodeJWKSRetrievalFailed          Code = "jwks retrieval failure"
)

// ErrorCode returns the error code from the given error.
// It unwraps the error chain to find the first error that implements
// the ErrorCode() string method that also has a non-empty code. If no
// such error is found, an empty Code is returned.
func ErrorCode(err error) Code {
	for err != nil {
		if v, ok := err.(interface{ ErrorCode() string }); ok {
			if code := v.ErrorCode(); code != "" {
				return Code(code)
			}
		}
		err = stderr.Unwrap(err)
	}
	return ""
}

// ErrorInfo returns additional information about the error.
// It unwraps the error chain to find the first error that implements
// the ErrorInfo() map[string]any method that also has non-nil info.
// If no such error is found, nil is returned.
func ErrorInfo(err error) map[string]any {
	for err != nil {
		if v, ok := err.(interface{ ErrorInfo() map[string]any }); ok {
			if info := v.ErrorInfo(); info != nil {
				return info
			}
		}
		err = stderr.Unwrap(err)
	}
	return nil
}
