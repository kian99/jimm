// Copyright 2024 Canonical.

package permissions_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/ofga"
	petname "github.com/dustinkirkland/golang-petname"
	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"
	"github.com/juju/juju/core/crossmodel"
	"github.com/juju/juju/state"
	"github.com/juju/names/v5"

	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/jimm/permissions"
	"github.com/canonical/jimm/v3/internal/openfga"
	ofganames "github.com/canonical/jimm/v3/internal/openfga/names"
	jimmnames "github.com/canonical/jimm/v3/pkg/names"
)

func (s *permissionManagerSuite) TestAuditLogAccess(c *qt.C) {
	c.Parallel()

	ctx := context.Background()

	// admin user can grant other users audit log access.
	err := s.manager.GrantAuditLogAccess(ctx, s.adminUser, s.user.ResourceTag())
	c.Assert(err, qt.IsNil)

	access := s.user.GetAuditLogViewerAccess(ctx, s.ctlTag)
	c.Assert(access, qt.Equals, ofganames.AuditLogViewerRelation)

	// re-granting access does not result in error.
	err = s.manager.GrantAuditLogAccess(ctx, s.adminUser, s.user.ResourceTag())
	c.Assert(err, qt.IsNil)

	// admin user can revoke other users audit log access.
	err = s.manager.RevokeAuditLogAccess(ctx, s.adminUser, s.user.ResourceTag())
	c.Assert(err, qt.IsNil)

	access = s.user.GetAuditLogViewerAccess(ctx, s.ctlTag)
	c.Assert(access, qt.Equals, ofganames.NoRelation)

	// re-revoking access does not result in error.
	err = s.manager.RevokeAuditLogAccess(ctx, s.adminUser, s.user.ResourceTag())
	c.Assert(err, qt.IsNil)

	// non-admin user cannot grant audit log access
	err = s.manager.GrantAuditLogAccess(ctx, s.user, s.adminUser.ResourceTag())
	c.Assert(err, qt.ErrorMatches, "unauthorized")

	// non-admin user cannot revoke audit log access
	err = s.manager.RevokeAuditLogAccess(ctx, s.user, s.adminUser.ResourceTag())
	c.Assert(err, qt.ErrorMatches, "unauthorized")
}

func (s *permissionManagerSuite) TestGrantServiceAccountAccess(c *qt.C) {
	c.Parallel()

	tests := []struct {
		about                     string
		grantServiceAccountAccess func(ctx context.Context, user *openfga.User, tags []string) error
		clientID                  string
		tags                      []string
		username                  string
		addGroups                 []string
		expectedError             string
	}{{
		about: "Valid request",
		grantServiceAccountAccess: func(ctx context.Context, user *openfga.User, tags []string) error {
			return nil
		},
		addGroups: []string{"1"},
		tags: []string{
			"user-alice",
			"user-bob",
			"group-1#member",
		},
		clientID: "fca1f605-736e-4d1f-bcd2-aecc726923be@serviceaccount",
		username: "alice",
	}, {
		about: "Group that doesn't exist",
		grantServiceAccountAccess: func(ctx context.Context, user *openfga.User, tags []string) error {
			return nil
		},
		tags: []string{
			"user-alice",
			"user-bob",
			// This group doesn't exist.
			"group-bar",
		},
		clientID:      "fca1f605-736e-4d1f-bcd2-aecc726923be@serviceaccount",
		username:      "alice",
		expectedError: "group bar not found",
	}, {
		about: "Invalid tags",
		grantServiceAccountAccess: func(ctx context.Context, user *openfga.User, tags []string) error {
			return nil
		},
		tags: []string{
			"user-alice",
			"user-bob",
			"controller-jimm",
		},
		clientID:      "fca1f605-736e-4d1f-bcd2-aecc726923be@serviceaccount",
		username:      "alice",
		expectedError: "invalid entity - not user or group",
	}}

	for _, test := range tests {
		test := test
		c.Run(test.about, func(c *qt.C) {
			if len(test.addGroups) > 0 {
				for _, name := range test.addGroups {
					_, err := s.db.AddGroup(context.Background(), name)
					c.Assert(err, qt.IsNil)
				}
			}
			svcAccountTag := jimmnames.NewServiceAccountTag(test.clientID)

			err := s.manager.GrantServiceAccountAccess(context.Background(), s.adminUser, svcAccountTag, test.tags)
			if test.expectedError == "" {
				c.Assert(err, qt.IsNil)
				for _, tag := range test.tags {
					parsedTag, err := s.manager.ParseAndValidateTag(context.Background(), tag)
					c.Assert(err, qt.IsNil)
					tuple := openfga.Tuple{
						Object:   parsedTag,
						Relation: ofganames.AdministratorRelation,
						Target:   ofganames.ConvertTag(jimmnames.NewServiceAccountTag(test.clientID)),
					}
					ok, err := s.ofgaClient.CheckRelation(context.Background(), tuple, false)
					c.Assert(err, qt.IsNil)
					c.Assert(ok, qt.IsTrue)
				}
			} else {
				c.Assert(err, qt.ErrorMatches, test.expectedError)
			}
		})
	}
}

func (s *permissionManagerSuite) TestParseAndValidateTag(c *qt.C) {
	c.Parallel()
	ctx := context.Background()

	user, _, _, model, _, _, _, _ := createTestControllerEnvironment(ctx, c, *s.db)

	jimmTag := "model-" + user.Name + "/" + model.Name + "#administrator"

	// JIMM tag syntax for models
	tag, err := s.manager.ParseAndValidateTag(ctx, jimmTag)
	c.Assert(err, qt.IsNil)
	c.Assert(tag.Kind.String(), qt.Equals, names.ModelTagKind)
	c.Assert(tag.ID, qt.Equals, model.UUID.String)
	c.Assert(tag.Relation.String(), qt.Equals, "administrator")

	jujuTag := "model-" + model.UUID.String + "#administrator"

	// Juju tag syntax for models
	tag, err = s.manager.ParseAndValidateTag(ctx, jujuTag)
	c.Assert(err, qt.IsNil)
	c.Assert(tag.ID, qt.Equals, model.UUID.String)
	c.Assert(tag.Kind.String(), qt.Equals, names.ModelTagKind)
	c.Assert(tag.Relation.String(), qt.Equals, "administrator")

	// JIMM tag only kind
	kindTag := "model"
	tag, err = s.manager.ParseAndValidateTag(ctx, kindTag)
	c.Assert(err, qt.IsNil)
	c.Assert(tag.ID, qt.Equals, "")
	c.Assert(tag.Kind.String(), qt.Equals, names.ModelTagKind)

	// JIMM tag not valid
	_, err = s.manager.ParseAndValidateTag(ctx, "")
	c.Assert(err, qt.ErrorMatches, "unknown tag kind")
}

func (s *permissionManagerSuite) TestResolveTags(c *qt.C) {
	c.Parallel()
	ctx := context.Background()

	identity, group, controller, model, offer, cloud, _, role := createTestControllerEnvironment(ctx, c, *s.db)

	testCases := []struct {
		desc     string
		input    string
		expected *ofga.Entity
	}{{
		desc:     "map identity name with relation",
		input:    "user-" + identity.Name + "#member",
		expected: ofganames.ConvertTagWithRelation(names.NewUserTag(identity.Name), ofganames.MemberRelation),
	}, {
		desc:     "map group name with relation",
		input:    "group-" + group.Name + "#member",
		expected: ofganames.ConvertTagWithRelation(jimmnames.NewGroupTag(group.UUID), ofganames.MemberRelation),
	}, {
		desc:     "map group UUID",
		input:    "group-" + group.UUID,
		expected: ofganames.ConvertTag(jimmnames.NewGroupTag(group.UUID)),
	}, {
		desc:     "map group UUID with relation",
		input:    "group-" + group.UUID + "#member",
		expected: ofganames.ConvertTagWithRelation(jimmnames.NewGroupTag(group.UUID), ofganames.MemberRelation),
	}, {
		desc:     "map role UUID",
		input:    "role-" + role.UUID,
		expected: ofganames.ConvertTag(jimmnames.NewRoleTag(role.UUID)),
	}, {
		desc:     "map role UUID with relation",
		input:    "role-" + role.UUID + "#assignee",
		expected: ofganames.ConvertTagWithRelation(jimmnames.NewRoleTag(role.UUID), ofganames.AssigneeRelation),
	}, {
		desc:     "map jimm controller",
		input:    "controller-" + "jimm",
		expected: ofganames.ConvertTag(s.ctlTag),
	}, {
		desc:     "map controller",
		input:    "controller-" + controller.Name + "#administrator",
		expected: ofganames.ConvertTagWithRelation(names.NewControllerTag(model.UUID.String), ofganames.AdministratorRelation),
	}, {
		desc:     "map controller UUID",
		input:    "controller-" + controller.UUID,
		expected: ofganames.ConvertTag(names.NewControllerTag(model.UUID.String)),
	}, {
		desc:     "map model",
		input:    "model-" + model.OwnerIdentityName + "/" + model.Name + "#administrator",
		expected: ofganames.ConvertTagWithRelation(names.NewModelTag(model.UUID.String), ofganames.AdministratorRelation),
	}, {
		desc:     "map model UUID",
		input:    "model-" + model.UUID.String,
		expected: ofganames.ConvertTag(names.NewModelTag(model.UUID.String)),
	}, {
		desc:     "map offer",
		input:    "applicationoffer-" + offer.URL + "#administrator",
		expected: ofganames.ConvertTagWithRelation(names.NewApplicationOfferTag(offer.UUID), ofganames.AdministratorRelation),
	}, {
		desc:     "map offer UUID",
		input:    "applicationoffer-" + offer.UUID,
		expected: ofganames.ConvertTag(names.NewApplicationOfferTag(offer.UUID)),
	}, {
		desc:     "map cloud",
		input:    "cloud-" + cloud.Name + "#administrator",
		expected: ofganames.ConvertTagWithRelation(names.NewCloudTag(cloud.Name), ofganames.AdministratorRelation),
	}}

	for _, tC := range testCases {
		c.Run(tC.desc, func(c *qt.C) {
			jujuTag, err := permissions.ResolveTag(s.ctlTag.Id(), s.db, tC.input)
			c.Assert(err, qt.IsNil)
			c.Assert(jujuTag, qt.DeepEquals, tC.expected)
		})
	}
}

func (s *permissionManagerSuite) TestResolveTupleObjectHandlesErrors(c *qt.C) {
	c.Parallel()
	ctx := context.Background()

	_, _, controller, model, offer, _, _, _ := createTestControllerEnvironment(ctx, c, *s.db)

	type test struct {
		input string
		want  string
	}

	tests := []test{
		// Resolves bad tuple objects in general
		{
			input: "unknowntag-blabla",
			want:  "failed to map tag, unknown kind: unknowntag",
		},
		// Resolves bad groups where they do not exist
		{
			input: "group-myspecialpokemon-his-name-is-youguessedit-diglett",
			want:  "group myspecialpokemon-his-name-is-youguessedit-diglett not found",
		},
		// Resolves bad controllers where they do not exist
		{
			input: "controller-mycontroller-that-does-not-exist",
			want:  "controller not found",
		},
		// Resolves bad models where the user cannot be obtained from the JIMM tag
		{
			input: "model-mycontroller-that-does-not-exist/mymodel",
			want:  "model not found",
		},
		// Resolves bad models where it cannot be found on the specified controller
		{
			input: "model-" + controller.Name + ":alex/",
			want:  "model name format incorrect, expected <model-owner>/<model-name>",
		},
		// Resolves bad applicationoffers where it cannot be found on the specified controller/model combo
		{
			input: "applicationoffer-" + controller.Name + ":alex/" + model.Name + "." + offer.UUID + "fluff",
			want:  "application offer not found",
		},
		{
			input: "abc",
			want:  "failed to setup tag resolver: tag is not properly formatted",
		},
		{
			input: "model-test-unknowncontroller-1:alice@canonical.com/test-model-1",
			want:  "model not found",
		},
	}
	for i, tc := range tests {
		c.Run(fmt.Sprintf("test %d", i), func(c *qt.C) {
			_, err := permissions.ResolveTag(s.ctlTag.Id(), s.db, tc.input)
			c.Assert(err, qt.ErrorMatches, tc.want)
		})
	}
}

func (s *permissionManagerSuite) TestToJAASTag(c *qt.C) {
	c.Parallel()
	ctx := context.Background()

	user, group, controller, model, applicationOffer, cloud, _, role := createTestControllerEnvironment(ctx, c, *s.db)

	serviceAccountId := petname.Generate(2, "-") + "@serviceaccount"

	tests := []struct {
		tag             *ofganames.Tag
		expectedJAASTag string
		expectedError   string
	}{{
		tag:             ofganames.ConvertTag(user.ResourceTag()),
		expectedJAASTag: "user-" + user.Name,
	}, {
		tag:             ofganames.ConvertTag(jimmnames.NewServiceAccountTag(serviceAccountId)),
		expectedJAASTag: "serviceaccount-" + serviceAccountId,
	}, {
		tag:             ofganames.ConvertTag(group.ResourceTag()),
		expectedJAASTag: "group-" + group.Name,
	}, {
		tag:             ofganames.ConvertTag(controller.ResourceTag()),
		expectedJAASTag: "controller-" + controller.Name,
	}, {
		tag:             ofganames.ConvertTag(model.ResourceTag()),
		expectedJAASTag: "model-" + user.Name + "/" + model.Name,
	}, {
		tag:             ofganames.ConvertTag(applicationOffer.ResourceTag()),
		expectedJAASTag: "applicationoffer-" + applicationOffer.URL,
	}, {
		tag:           &ofganames.Tag{},
		expectedError: "unexpected tag kind: ",
	}, {
		tag:             ofganames.ConvertTag(cloud.ResourceTag()),
		expectedJAASTag: "cloud-" + cloud.Name,
	}, {
		tag:             ofganames.ConvertTag(role.ResourceTag()),
		expectedJAASTag: "role-" + role.Name,
	}}
	for _, test := range tests {
		t, err := s.manager.ToJAASTag(ctx, test.tag, true)
		if test.expectedError != "" {
			c.Assert(err, qt.ErrorMatches, test.expectedError)
		} else {
			c.Assert(err, qt.IsNil)
			c.Assert(t, qt.Equals, test.expectedJAASTag)
		}
	}
}

func (s *permissionManagerSuite) TestToJAASTagNoUUIDResolution(c *qt.C) {
	c.Parallel()
	ctx := context.Background()

	user, group, controller, model, applicationOffer, cloud, _, role := createTestControllerEnvironment(ctx, c, *s.db)
	serviceAccountId := petname.Generate(2, "-") + "@serviceaccount"

	tests := []struct {
		tag             *ofganames.Tag
		expectedJAASTag string
		expectedError   string
	}{{
		tag:             ofganames.ConvertTag(user.ResourceTag()),
		expectedJAASTag: "user-" + user.Name,
	}, {
		tag:             ofganames.ConvertTag(jimmnames.NewServiceAccountTag(serviceAccountId)),
		expectedJAASTag: "serviceaccount-" + serviceAccountId,
	}, {
		tag:             ofganames.ConvertTag(group.ResourceTag()),
		expectedJAASTag: "group-" + group.UUID,
	}, {
		tag:             ofganames.ConvertTag(controller.ResourceTag()),
		expectedJAASTag: "controller-" + controller.UUID,
	}, {
		tag:             ofganames.ConvertTag(model.ResourceTag()),
		expectedJAASTag: "model-" + model.UUID.String,
	}, {
		tag:             ofganames.ConvertTag(applicationOffer.ResourceTag()),
		expectedJAASTag: "applicationoffer-" + applicationOffer.UUID,
	}, {
		tag:             ofganames.ConvertTag(cloud.ResourceTag()),
		expectedJAASTag: "cloud-" + cloud.Name,
	}, {
		tag:             ofganames.ConvertTag(role.ResourceTag()),
		expectedJAASTag: "role-" + role.UUID,
	}, {
		tag:             &ofganames.Tag{},
		expectedJAASTag: "-",
	}}
	for _, test := range tests {
		t, err := s.manager.ToJAASTag(ctx, test.tag, false)
		if test.expectedError != "" {
			c.Assert(err, qt.ErrorMatches, test.expectedError)
		} else {
			c.Assert(err, qt.IsNil)
			c.Assert(t, qt.Equals, test.expectedJAASTag)
		}
	}
}

// createTestControllerEnvironment is a utility function creating the necessary components of adding a:
//   - user
//   - user group
//   - controller
//   - model
//   - application offer
//   - cloud
//   - cloud credential
//   - role
//
// Into the test database, returning the dbmodels to be utilised for values within tests.
//
// It returns all of the latter, but in addition to those, also:
//   - an api client to make calls to an httptest instance of the server
//   - a closure containing a function to close the connection
//
// TODO(ale8k): Make this an implicit thing on the JIMM suite per test & refactor the current state.
// and make the suite argument an interface of the required calls we use here.
func createTestControllerEnvironment(ctx context.Context, c *qt.C, db db.Database) (
	dbmodel.Identity,
	dbmodel.GroupEntry,
	dbmodel.Controller,
	dbmodel.Model,
	dbmodel.ApplicationOffer,
	dbmodel.Cloud,
	dbmodel.CloudCredential,
	dbmodel.RoleEntry) {

	_, err := db.AddGroup(ctx, "test-group")
	c.Assert(err, qt.IsNil)
	group := dbmodel.GroupEntry{Name: "test-group"}
	err = db.GetGroup(ctx, &group)
	c.Assert(err, qt.IsNil)

	u, err := dbmodel.NewIdentity(petname.Generate(2, "-"+"canonical.com"))
	c.Assert(err, qt.IsNil)

	c.Assert(db.DB.Create(u).Error, qt.IsNil)

	cloud := dbmodel.Cloud{
		Name: petname.Generate(2, "-"),
		Type: "aws",
		Regions: []dbmodel.CloudRegion{{
			Name: petname.Generate(2, "-"),
		}},
	}
	c.Assert(db.DB.Create(&cloud).Error, qt.IsNil)
	id, _ := uuid.NewRandom()
	controller := dbmodel.Controller{
		Name:        petname.Generate(2, "-"),
		UUID:        id.String(),
		CloudName:   cloud.Name,
		CloudRegion: cloud.Regions[0].Name,
		CloudRegions: []dbmodel.CloudRegionControllerPriority{{
			Priority:      0,
			CloudRegionID: cloud.Regions[0].ID,
		}},
	}
	err = db.AddController(ctx, &controller)
	c.Assert(err, qt.IsNil)

	cred := dbmodel.CloudCredential{
		Name:              petname.Generate(2, "-"),
		CloudName:         cloud.Name,
		OwnerIdentityName: u.Name,
		AuthType:          "empty",
	}
	err = db.SetCloudCredential(ctx, &cred)
	c.Assert(err, qt.IsNil)

	model := dbmodel.Model{
		Name: petname.Generate(2, "-"),
		UUID: sql.NullString{
			String: id.String(),
			Valid:  true,
		},
		OwnerIdentityName: u.Name,
		ControllerID:      controller.ID,
		CloudRegionID:     cloud.Regions[0].ID,
		CloudCredentialID: cred.ID,
		Life:              state.Alive.String(),
		Status: dbmodel.Status{
			Status: "available",
			Since: sql.NullTime{
				Time:  time.Now().UTC().Truncate(time.Millisecond),
				Valid: true,
			},
		},
	}

	err = db.AddModel(ctx, &model)
	c.Assert(err, qt.IsNil)

	offerName := petname.Generate(2, "-")
	offerURL, err := crossmodel.ParseOfferURL(controller.Name + ":" + u.Name + "/" + model.Name + "." + offerName)
	c.Assert(err, qt.IsNil)

	offer := dbmodel.ApplicationOffer{
		UUID:    id.String(),
		Name:    offerName,
		ModelID: model.ID,
		URL:     offerURL.String(),
	}
	err = db.AddApplicationOffer(context.Background(), &offer)
	c.Assert(err, qt.IsNil)
	c.Assert(len(offer.UUID), qt.Equals, 36)

	role, err := db.AddRole(ctx, petname.Generate(2, "-"))
	c.Assert(err, qt.IsNil)

	return *u, group, controller, model, offer, cloud, cred, *role
}

func (s *permissionManagerSuite) TestOpenFGACleanup(c *qt.C) {
	c.Parallel()
	ctx := context.Background()

	// run cleanup on an empty authorizaton store
	err := s.manager.OpenFGACleanup(ctx)
	c.Assert(err, qt.IsNil)

	type createTagFunction func(int) *ofga.Entity

	var (
		createStringTag = func(kind openfga.Kind) createTagFunction {
			return func(i int) *ofga.Entity {
				return &ofga.Entity{
					Kind: kind,
					ID:   fmt.Sprintf("%s-%d", petname.Generate(2, "-"), i),
				}
			}
		}

		createUUIDTag = func(kind openfga.Kind) createTagFunction {
			return func(i int) *ofga.Entity {
				return &ofga.Entity{
					Kind: kind,
					ID:   uuid.NewString(),
				}
			}
		}
	)

	tagTests := []struct {
		createObjectTag createTagFunction
		relation        string
		createTargetTag createTagFunction
	}{{
		createObjectTag: createStringTag(openfga.UserType),
		relation:        "member",
		createTargetTag: createStringTag(openfga.GroupType),
	}, {
		createObjectTag: createStringTag(openfga.UserType),
		relation:        "administrator",
		createTargetTag: createUUIDTag(openfga.ControllerType),
	}, {
		createObjectTag: createStringTag(openfga.UserType),
		relation:        "reader",
		createTargetTag: createUUIDTag(openfga.ModelType),
	}, {
		createObjectTag: createStringTag(openfga.UserType),
		relation:        "administrator",
		createTargetTag: createStringTag(openfga.CloudType),
	}, {
		createObjectTag: createStringTag(openfga.UserType),
		relation:        "consumer",
		createTargetTag: createUUIDTag(openfga.ApplicationOfferType),
	}}

	orphanedTuples := []ofga.Tuple{}
	for i := 0; i < 100; i++ {
		for _, test := range tagTests {
			objectTag := test.createObjectTag(i)
			targetTag := test.createTargetTag(i)

			tuple := openfga.Tuple{
				Object:   objectTag,
				Relation: ofga.Relation(test.relation),
				Target:   targetTag,
			}
			err = s.ofgaClient.AddRelation(ctx, tuple)
			c.Assert(err, qt.IsNil)

			orphanedTuples = append(orphanedTuples, tuple)
		}
	}

	err = s.manager.OpenFGACleanup(ctx)
	c.Assert(err, qt.IsNil)

	for _, tuple := range orphanedTuples {
		c.Logf("checking relation for %+v", tuple)
		ok, err := s.ofgaClient.CheckRelation(ctx, tuple, false)
		c.Assert(err, qt.IsNil)
		c.Assert(ok, qt.IsFalse)
	}
}
