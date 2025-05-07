// Copyright 2025 Canonical.

package jujuclient_test

import (
	"context"

	gc "gopkg.in/check.v1"
)

type sshClientSuite struct {
	jujuclientSuite
}

// TODO: Uncomment this once we implement the test.
// The jujuclient test suites are prohibitive since
// they require a running Juju controller.
// var _ = gc.Suite(&sshClientSuite{})

func (s *sshClientSuite) TestControllerHostKey(c *gc.C) {
	ctx := context.Background()

	_, err := s.API.ControllerHostKey(ctx)
	c.Assert(err, gc.Equals, nil)
}
