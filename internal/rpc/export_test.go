// Copyright 2025 Canonical.

package rpc

import (
	"net/http"
	"net/url"
)

type Message = message
type MultiBackendTransport = multiBackendTransport
type Dialer = dialer

func NewMultiBackendTransport(transport http.RoundTripper, urls []*url.URL) (*MultiBackendTransport, error) {
	return newMultiBackendTransport(transport, urls)
}
