// Copyright 2020 Canonical Ltd.

package jujuclient_test

import (
	"context"
	"fmt"

	jujuparams "github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/core/network"
	jujuversion "github.com/juju/juju/version"
	"github.com/juju/names/v4"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/CanonicalLtd/jimm/internal/dbmodel"
	"github.com/CanonicalLtd/jimm/internal/jimm"
	"github.com/CanonicalLtd/jimm/internal/jimmtest"
	"github.com/CanonicalLtd/jimm/internal/jujuclient"
	"github.com/CanonicalLtd/jimm/internal/vault"
)

type jujuclientSuite struct {
	jimmtest.JujuSuite

	Dialer jimm.Dialer
	API    jimm.API
}

func (s *jujuclientSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)

	s.Dialer = &jujuclient.Dialer{}
	var err error
	info := s.APIInfo(c)
	hpss := make(dbmodel.HostPorts, 0, len(info.Addrs))
	for _, addr := range info.Addrs {
		hp, err := network.ParseMachineHostPort(addr)
		if err != nil {
			continue
		}
		hpss = append(hpss, []jujuparams.HostPort{{
			Address: jujuparams.FromMachineAddress(hp.MachineAddress),
			Port:    hp.Port(),
		}})
	}
	ctl := dbmodel.Controller{
		Name:          s.ControllerConfig.ControllerName(),
		CACertificate: info.CACert,
		AdminUser:     info.Tag.Id(),
		AdminPassword: info.Password,
		Addresses:     hpss,
	}
	s.API, err = s.Dialer.Dial(context.Background(), &ctl, names.ModelTag{})
	c.Assert(err, gc.Equals, nil)
}

func (s *jujuclientSuite) TearDownTest(c *gc.C) {
	if s.API != nil {
		err := s.API.Close()
		s.API = nil
		c.Assert(err, gc.Equals, nil)
	}
	s.JujuConnSuite.TearDownTest(c)
}

type dialSuite struct {
	jujuclientSuite
}

var _ = gc.Suite(&dialSuite{})

func (s *dialSuite) TestDial(c *gc.C) {
	info := s.APIInfo(c)
	ctl := dbmodel.Controller{
		Name:          s.ControllerConfig.ControllerName(),
		CACertificate: info.CACert,
		AdminUser:     info.Tag.Id(),
		AdminPassword: info.Password,
		PublicAddress: info.Addrs[0],
	}
	api, err := s.Dialer.Dial(context.Background(), &ctl, names.ModelTag{})
	c.Assert(err, gc.Equals, nil)
	defer api.Close()
	c.Check(ctl.UUID, gc.Equals, "deadbeef-1bad-500d-9000-4b1d0d06f00d")
	c.Check(ctl.AgentVersion, gc.Equals, jujuversion.Current.String())
	addrs := make([]string, len(ctl.Addresses))
	for i, addr := range ctl.Addresses {
		addrs[i] = fmt.Sprintf("%s:%d", addr[0].Value, addr[0].Port)
	}
	c.Check(addrs, jc.DeepEquals, info.Addrs)
}

type cExtended struct {
	*gc.C
}

func (t *cExtended) Name() string {
	return t.TestName()
}

func (s *dialSuite) TestDialWithCredentialsStoredInVault(c *gc.C) {
	jimmtest.StartVault()
	defer jimmtest.StopVault()

	client, path, creds, ok := jimmtest.VaultClient(&cExtended{c})
	if !ok {
		c.Skip("vault not available")
	}
	store := &vault.VaultStore{
		Client:     client,
		AuthSecret: creds,
		AuthPath:   jimmtest.VaultAuthPath,
		KVPath:     path,
	}

	info := s.APIInfo(c)
	ctl := dbmodel.Controller{
		Name:          s.ControllerConfig.ControllerName(),
		CACertificate: info.CACert,
		PublicAddress: info.Addrs[0],
		AdminUser:     info.Tag.Id(),
		AdminPassword: info.Password,
	}

	err := store.PutControllerCredentials(
		context.Background(),
		ctl.Name,
		info.Tag.Id(),
		info.Password,
	)
	c.Assert(err, gc.IsNil)

	dialer := &jujuclient.Dialer{
		ControllerCredentialsStore: store,
	}

	api, err := dialer.Dial(context.Background(), &ctl, names.ModelTag{})
	c.Assert(err, gc.Equals, nil)
	defer api.Close()
	c.Check(ctl.UUID, gc.Equals, "deadbeef-1bad-500d-9000-4b1d0d06f00d")
	c.Check(ctl.AgentVersion, gc.Equals, jujuversion.Current.String())
	addrs := make([]string, len(ctl.Addresses))
	for i, addr := range ctl.Addresses {
		addrs[i] = fmt.Sprintf("%s:%d", addr[0].Value, addr[0].Port)
	}
	c.Check(addrs, gc.DeepEquals, info.Addrs)
}
