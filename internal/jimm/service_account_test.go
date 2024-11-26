// Copyright 2024 Canonical.

package jimm_test

import (
	"context"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"
	jujuparams "github.com/juju/juju/rpc/params"

	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/jimm"
	"github.com/canonical/jimm/v3/internal/openfga"
	ofganames "github.com/canonical/jimm/v3/internal/openfga/names"
	"github.com/canonical/jimm/v3/internal/testutils/jimmtest"
)

func TestAddServiceAccount(t *testing.T) {
	c := qt.New(t)

	ctx := context.Background()
	client, _, _, err := jimmtest.SetupTestOFGAClient(c.Name())
	c.Assert(err, qt.IsNil)
	j := &jimm.JIMM{
		OpenFGAClient: client,
	}
	c.Assert(err, qt.IsNil)

	bob, err := dbmodel.NewIdentity("bob@canonical.com")
	c.Assert(err, qt.IsNil)
	user := openfga.NewUser(
		bob,
		client,
	)
	clientID := "39caae91-b914-41ae-83f8-c7b86ca5ad5a@serviceaccount"
	err = j.AddServiceAccount(ctx, user, clientID)
	c.Assert(err, qt.IsNil)
	err = j.AddServiceAccount(ctx, user, clientID)
	c.Assert(err, qt.IsNil)

	alive, err := dbmodel.NewIdentity("alive@canonical.com")
	c.Assert(err, qt.IsNil)
	userAlice := openfga.NewUser(
		alive,
		client,
	)
	err = j.AddServiceAccount(ctx, userAlice, clientID)
	c.Assert(err, qt.ErrorMatches, "service account already owned")
}

func TestCopyServiceAccountCredential(t *testing.T) {
	c := qt.New(t)

	ctx := context.Background()
	client, _, _, err := jimmtest.SetupTestOFGAClient(c.Name())
	c.Assert(err, qt.IsNil)
	api := &jimmtest.API{
		CheckCredentialModels_: func(context.Context, jujuparams.TaggedCredential) ([]jujuparams.UpdateCredentialModelResult, error) {
			return []jujuparams.UpdateCredentialModelResult{}, nil
		},
		UpdateCredential_: func(context.Context, jujuparams.TaggedCredential) ([]jujuparams.UpdateCredentialModelResult, error) {
			return []jujuparams.UpdateCredentialModelResult{}, nil
		},
	}
	j := &jimm.JIMM{
		UUID: uuid.NewString(),
		Database: db.Database{
			DB: jimmtest.PostgresDB(c, func() time.Time { return now }),
		},
		Dialer: &jimmtest.Dialer{
			API: api,
		},
		OpenFGAClient: client,
	}

	err = j.Database.Migrate(ctx, false)
	c.Assert(err, qt.IsNil)
	svcAccId, err := dbmodel.NewIdentity("39caae91-b914-41ae-83f8-c7b86ca5ad5a@serviceaccount")
	c.Assert(err, qt.IsNil)
	c.Assert(j.Database.DB.Create(&svcAccId).Error, qt.IsNil)
	svcAcc := openfga.NewUser(svcAccId, client)
	u, err := dbmodel.NewIdentity("alice@canonical.com")
	c.Assert(err, qt.IsNil)

	c.Assert(j.Database.DB.Create(&u).Error, qt.IsNil)

	user := openfga.NewUser(u, client)

	err = user.SetControllerAccess(context.Background(), j.ResourceTag(), ofganames.AdministratorRelation)
	c.Assert(err, qt.IsNil)

	// Create cloud, controller and cloud-credential as setup for test.
	cloud := dbmodel.Cloud{
		Name: "test-cloud",
		Type: "test-provider",
		Regions: []dbmodel.CloudRegion{{
			Name: "test-region-1",
		}},
	}
	c.Assert(j.Database.DB.Create(&cloud).Error, qt.IsNil)

	err = user.SetCloudAccess(context.Background(), cloud.ResourceTag(), ofganames.AdministratorRelation)
	c.Assert(err, qt.IsNil)

	controller1 := dbmodel.Controller{
		Name:        "test-controller-1",
		UUID:        "00000000-0000-0000-0000-0000-0000000000001",
		CloudName:   "test-cloud",
		CloudRegion: "test-region-1",
		CloudRegions: []dbmodel.CloudRegionControllerPriority{{
			Priority:      0,
			CloudRegionID: cloud.Regions[0].ID,
		}},
	}
	err = j.Database.AddController(context.Background(), &controller1)
	c.Assert(err, qt.Equals, nil)

	cred := dbmodel.CloudCredential{
		Name:              "test-credential-1",
		CloudName:         cloud.Name,
		OwnerIdentityName: u.Name,
		AuthType:          "empty",
	}
	err = j.Database.SetCloudCredential(context.Background(), &cred)
	c.Assert(err, qt.Equals, nil)

	_, res, err := j.CopyServiceAccountCredential(ctx, user, svcAcc, cred.ResourceTag())
	c.Assert(err, qt.Equals, nil)
	newCred := dbmodel.CloudCredential{
		Name:              "test-credential-1",
		CloudName:         cloud.Name,
		OwnerIdentityName: svcAcc.Name,
	}
	c.Assert(len(res), qt.Equals, 0)
	err = j.Database.GetCloudCredential(context.Background(), &newCred)
	c.Assert(err, qt.Equals, nil)
}

func TestCopyServiceAccountCredentialWithMissingCredential(t *testing.T) {
	c := qt.New(t)

	ctx := context.Background()
	client, _, _, err := jimmtest.SetupTestOFGAClient(c.Name())
	c.Assert(err, qt.IsNil)
	j := &jimm.JIMM{
		UUID: uuid.NewString(),
		Database: db.Database{
			DB: jimmtest.PostgresDB(c, func() time.Time { return now }),
		},
		OpenFGAClient: client,
	}

	err = j.Database.Migrate(ctx, false)
	c.Assert(err, qt.IsNil)
	svcAccId, err := dbmodel.NewIdentity("39caae91-b914-41ae-83f8-c7b86ca5ad5a@serviceaccount")
	c.Assert(err, qt.IsNil)
	c.Assert(j.Database.DB.Create(&svcAccId).Error, qt.IsNil)
	svcAcc := openfga.NewUser(svcAccId, client)
	u, err := dbmodel.NewIdentity("alice@canonical.com")
	c.Assert(err, qt.IsNil)
	c.Assert(j.Database.DB.Create(&u).Error, qt.IsNil)
	user := openfga.NewUser(u, client)

	cred := dbmodel.CloudCredential{
		Name:              "test-credential-1",
		CloudName:         "fake-cloud",
		OwnerIdentityName: u.Name,
		AuthType:          "empty",
	}
	_, _, err = j.CopyServiceAccountCredential(ctx, user, svcAcc, cred.ResourceTag())
	c.Assert(err, qt.ErrorMatches, "cloudcredential .* not found")
}
