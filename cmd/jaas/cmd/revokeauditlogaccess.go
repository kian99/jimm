// Copyright 2025 Canonical.

package cmd

import (
	"github.com/juju/cmd/v3"
	"github.com/juju/gnuflag"
	jujuapi "github.com/juju/juju/api"
	jujucmd "github.com/juju/juju/cmd"
	"github.com/juju/juju/cmd/modelcmd"
	"github.com/juju/juju/jujuclient"
	"github.com/juju/names/v5"

	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/pkg/api"
	apiparams "github.com/canonical/jimm/v3/pkg/api/params"
)

const (
	revokeAuditLogAccessDoc = `
Revokes user access to audit logs.
`
	revokeAuditLogAccessExample = `
    juju revoke-audit-log user@canonical.com
`
)

// NewrevokeAuditLogAccess returns a command used to revoke
// users access to audit logs.
func NewRevokeAuditLogAccessCommand() cmd.Command {
	cmd := &revokeAuditLogAccessCommand{
		store: jujuclient.NewFileClientStore(),
	}
	cmd.revokeAuditLogAccessAPIFunc = cmd.newClient

	return modelcmd.WrapBase(cmd)
}

// revokeAuditLogAccess displays full
// model status.
type revokeAuditLogAccessCommand struct {
	modelcmd.ControllerCommandBase

	revokeAuditLogAccessAPIFunc func() (JIMMAPI, error)
	store                       jujuclient.ClientStore
	dialOpts                    *jujuapi.DialOpts
	username                    string
}

func (c *revokeAuditLogAccessCommand) Info() *cmd.Info {
	return jujucmd.Info(&cmd.Info{
		Name:     "revoke-audit-log",
		Args:     "<user>",
		Purpose:  "revokes access to audit logs.",
		Doc:      revokeAuditLogAccessDoc,
		Examples: revokeAuditLogAccessExample,
	})
}

// SetFlags implements Command.SetFlags.
func (c *revokeAuditLogAccessCommand) SetFlags(f *gnuflag.FlagSet) {
	c.CommandBase.SetFlags(f)
}

// Init implements the cmd.Command interface.
func (c *revokeAuditLogAccessCommand) Init(args []string) error {
	if len(args) == 0 {
		return errors.E("missing username")
	}
	c.username, args = args[0], args[1:]
	if len(args) > 0 {
		return errors.E("unknown arguments")
	}
	return nil
}

// Run implements Command.Run.
func (c *revokeAuditLogAccessCommand) Run(ctxt *cmd.Context) error {
	userTag := names.NewUserTag(c.username)

	client, err := c.revokeAuditLogAccessAPIFunc()
	if err != nil {
		return errors.E(err, "could not create JIMM client")
	}
	defer client.Close()

	err = client.RevokeAuditLogAccess(&apiparams.AuditLogAccessRequest{
		UserTag: userTag.String(),
	})
	if err != nil {
		return errors.E(err)
	}

	return nil
}

func (c *revokeAuditLogAccessCommand) newClient() (JIMMAPI, error) {
	currentController, err := c.store.CurrentController()
	if err != nil {
		return nil, errors.E(err, "could not determine controller")
	}

	apiCaller, err := c.NewAPIRootWithDialOpts(c.store, currentController, "", c.dialOpts)
	if err != nil {
		return nil, err
	}

	return api.NewClient(apiCaller), nil
}
