// Copyright 2025 Canonical.

//go:build integration

package cmd_test

import (
	"github.com/juju/cmd/v3/cmdtesting"
	jujuparams "github.com/juju/juju/rpc/params"
	jujutesting "github.com/juju/testing"
	gc "gopkg.in/check.v1"

	"github.com/canonical/jimm/v3/cmd/jaas/cmd"
	"github.com/canonical/jimm/v3/internal/testutils/cmdtest"
	apiparams "github.com/canonical/jimm/v3/pkg/api/params"
)

type removeCloudFromControllerSuite struct {
	cmdtest.JimmCmdSuite

	api *fakeRemoveCloudFromControllerAPI
}

var _ = gc.Suite(&removeCloudFromControllerSuite{})

func (s *removeCloudFromControllerSuite) SetUpTest(c *gc.C) {
	s.JimmCmdSuite.SetUpTest(c)
	s.api = &fakeRemoveCloudFromControllerAPI{}
}

func (s *removeCloudFromControllerSuite) TestRemoveCloudFromController(c *gc.C) {
	bClient := s.SetupCLIAccess(c, "alice@canonical.com")

	command := cmd.NewRemoveCloudFromControllerCommandForTesting(
		s.ClientStore(),
		bClient,
		func() (cmd.JIMMAPI, error) {
			return s.api, nil
		})
	ctx, err := cmdtesting.RunCommand(c, command, "controller-1", "test-cloud")
	c.Assert(err, gc.IsNil)
	s.api.CheckCallNames(c, "RemoveCloudFromController", "Close")
	s.api.CheckCalls(c, []jujutesting.StubCall{{
		FuncName: "RemoveCloudFromController",
		Args: []interface{}{&apiparams.RemoveCloudFromControllerRequest{
			ControllerName: "controller-1",
			CloudTag:       "cloud-test-cloud",
		}},
	}, {
		FuncName: "Close",
		Args:     []interface{}{nil},
	}})
	c.Assert(cmdtesting.Stderr(ctx), gc.Equals, "Cloud \"test-cloud\" removed from controller \"controller-1\".\n")
}

func (s *removeCloudFromControllerSuite) TestRemoveCloudFromControllerWrongArguments(c *gc.C) {
	bClient := s.SetupCLIAccess(c, "alice@canonical.com")

	command := cmd.NewRemoveCloudFromControllerCommandForTesting(
		s.ClientStore(),
		bClient,
		func() (cmd.JIMMAPI, error) {
			return s.api, nil
		})
	_, err := cmdtesting.RunCommand(c, command, "controller-1")
	c.Assert(err, gc.ErrorMatches, "missing arguments")
	_, err = cmdtesting.RunCommand(c, command, "controller-1", "cloud", "fake-arg")
	c.Assert(err, gc.ErrorMatches, "too many arguments")
}

func (s *removeCloudFromControllerSuite) TestRemoveCloudFromControllerCloudNotFound(c *gc.C) {
	bClient := s.SetupCLIAccess(c, "alice@canonical.com")

	command := cmd.NewRemoveCloudFromControllerCommandForTesting(
		s.ClientStore(),
		bClient,
		nil)
	_, err := cmdtesting.RunCommand(c, command, "controller-1", "test-cloud")
	c.Assert(err, gc.ErrorMatches, ".*cloud \"test-cloud\" not found.*")
}

type fakeRemoveCloudFromControllerAPI struct {
	jujutesting.Stub
}

func (api *fakeRemoveCloudFromControllerAPI) Close() error {
	api.AddCall("Close", nil)
	return api.NextErr()
}

func (api *fakeRemoveCloudFromControllerAPI) RemoveCloudFromController(params *apiparams.RemoveCloudFromControllerRequest) error {
	api.AddCall("RemoveCloudFromController", params)
	return api.NextErr()
}

// Stub implementations for JIMMAPI interface - not used in these tests
func (api *fakeRemoveCloudFromControllerAPI) GetJobInfo(req *apiparams.GetJobInfoRequest) (apiparams.GetJobInfoResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) StopJob(req *apiparams.StopJobRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) StartBootstrapJob(req *apiparams.BootstrapParams) (*apiparams.StartJobResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) StartDestroyControllerJob(req *apiparams.DestroyControllerRequest) (*apiparams.StartJobResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) AddCloudToController(req *apiparams.AddCloudToControllerRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) AddController(req *apiparams.AddControllerRequest) (apiparams.ControllerInfo, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) ListControllers() ([]apiparams.ControllerInfo, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) RemoveController(req *apiparams.RemoveControllerRequest) (apiparams.ControllerInfo, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) SetControllerDeprecated(req *apiparams.SetControllerDeprecatedRequest) (apiparams.ControllerInfo, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) ListMigrationTargets(req *apiparams.ListMigrationTargetsRequest) ([]apiparams.ControllerInfo, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) PrepareModelMigration(req *apiparams.PrepareModelMigrationRequest) (apiparams.PrepareModelMigrationResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) MigrateModel(req *apiparams.MigrateModelRequest) (*jujuparams.InitiateMigrationResults, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) ImportModel(req *apiparams.ImportModelRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) UpdateMigratedModel(req *apiparams.UpdateMigratedModelRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) FullModelStatus(req *apiparams.FullModelStatusRequest) (jujuparams.FullStatus, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) FindAuditEvents(req *apiparams.FindAuditEventsRequest) (apiparams.AuditEvents, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) GrantAuditLogAccess(req *apiparams.AuditLogAccessRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) RevokeAuditLogAccess(req *apiparams.AuditLogAccessRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) PurgeLogs(req *apiparams.PurgeLogsRequest) (*apiparams.PurgeLogsResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) AddGroup(req *apiparams.AddGroupRequest) (apiparams.AddGroupResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) GetGroup(req *apiparams.GetGroupRequest) (apiparams.GetGroupResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) RenameGroup(req *apiparams.RenameGroupRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) RemoveGroup(req *apiparams.RemoveGroupRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) ListGroups(req *apiparams.ListGroupsRequest) ([]apiparams.Group, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) AddRole(req *apiparams.AddRoleRequest) (apiparams.AddRoleResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) GetRole(req *apiparams.GetRoleRequest) (apiparams.GetRoleResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) RenameRole(req *apiparams.RenameRoleRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) RemoveRole(req *apiparams.RemoveRoleRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) ListRoles(req *apiparams.ListRolesRequest) ([]apiparams.Role, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) AddRelation(req *apiparams.AddRelationRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) RemoveRelation(req *apiparams.RemoveRelationRequest) error {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) CheckRelation(req *apiparams.CheckRelationRequest) (apiparams.CheckRelationResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) ListRelationshipTuples(req *apiparams.ListRelationshipTuplesRequest) (*apiparams.ListRelationshipTuplesResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) CrossModelQuery(req *apiparams.CrossModelQueryRequest) (*apiparams.CrossModelQueryResponse, error) {
	panic("not implemented")
}
func (api *fakeRemoveCloudFromControllerAPI) UpgradeTo(req *apiparams.UpgradeToRequest) (apiparams.UpgradeToResponse, error) {
	panic("not implemented")
}
