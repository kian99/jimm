package login_test

import (
	"context"
	"time"

	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/jimm"
	ofganames "github.com/canonical/jimm/v3/internal/openfga/names"
	"github.com/canonical/jimm/v3/internal/testutils/jimmtest"
	"github.com/juju/names/v5"

	qt "github.com/frankban/quicktest"
)

func (s *loginManagerSuite) TestGetUser(c *qt.C) {
	c.Parallel()

	ctx := context.Background()
	client, _, _, err := jimmtest.SetupTestOFGAClient(c.Name())
	c.Assert(err, qt.IsNil)
	db := &db.Database{
		DB: jimmtest.PostgresDB(c, time.Now),
	}

	j := &jimm.JIMM{
		UUID:          "test",
		Database:      *db,
		OpenFGAClient: client,
	}

	err = j.Database.Migrate(ctx, false)
	c.Assert(err, qt.IsNil)

	ofgaUser, err := s.manager.GetUser(ctx, "bob@canonical.com.com")
	c.Assert(err, qt.IsNil)
	// Username -> email
	c.Assert(ofgaUser.Name, qt.Equals, "bob@canonical.com.com")
	// As no display name was set for this user as they're being created this time over
	c.Assert(ofgaUser.DisplayName, qt.Equals, "bob")
	// This user SHOULD NOT be an admin, so ensure admin check is OK
	c.Assert(ofgaUser.JimmAdmin, qt.IsFalse)

	// Next we'll update this user to an admin of JIMM and run the same tests.
	c.Assert(
		ofgaUser.SetControllerAccess(
			context.Background(),
			names.NewControllerTag(j.UUID),
			ofganames.AdministratorRelation,
		),
		qt.IsNil,
	)

	ofgaUser, err = s.manager.GetUser(ctx, "bob@canonical.com.com")
	c.Assert(err, qt.IsNil)

	c.Assert(ofgaUser.Name, qt.Equals, "bob@canonical.com.com")
	c.Assert(ofgaUser.DisplayName, qt.Equals, "bob")
	// This user SHOULD be an admin, so ensure admin check is OK
	c.Assert(ofgaUser.JimmAdmin, qt.IsTrue)
}

func (s *loginManagerSuite) TestUpdateUserLastLogin(c *qt.C) {
	c.Parallel()

	ctx := context.Background()

	err := s.manager.UpdateUserLastLogin(ctx, "bob@canonical.com.com")
	c.Assert(err, qt.IsNil)
	user := dbmodel.Identity{Name: "bob@canonical.com.com"}
	err = s.db.GetIdentity(ctx, &user)
	c.Assert(err, qt.IsNil)
	c.Assert(user.DisplayName, qt.Equals, "bob")
	c.Assert(user.LastLogin.Time, qt.Not(qt.Equals), time.Time{})
	c.Assert(user.LastLogin.Valid, qt.IsTrue)
}
