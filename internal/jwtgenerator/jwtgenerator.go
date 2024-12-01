package jwtgenerator

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/jimmjwx"
	"github.com/canonical/jimm/v3/internal/openfga"
	ofganames "github.com/canonical/jimm/v3/internal/openfga/names"

	"github.com/juju/juju/core/permission"
	"github.com/juju/names/v5"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
)

// jwtGeneratorDatabase specifies the database interface used by the
// JWT generator.
type jwtGeneratorDatabase interface {
	GetController(ctx context.Context, controller *dbmodel.Controller) error
}

// jwtGeneratorAccessChecker specifies the access checker used by the JWT
// generator to obtain user's access rights to various entities.
type jwtGeneratorAccessChecker interface {
	GetUserModelAccess(context.Context, *openfga.User, names.ModelTag) (openfga.Relation, error)
	GetUserControllerAccess(context.Context, *openfga.User, names.ControllerTag) (openfga.Relation, error)
	GetUserCloudAccess(context.Context, *openfga.User, names.CloudTag) (openfga.Relation, error)
	CheckBatchPermissions(ctx context.Context, user *openfga.User, desiredPerms map[names.Tag]openfga.Relation) (map[names.Tag]openfga.Relation, error)
}

// jwtService specifies the service JWT generator uses to generate JWTs.
type jwtService interface {
	NewJWT(context.Context, jimmjwx.JWTParams) ([]byte, error)
}

// JWTGenerator provides the necessary state and methods to authorize a user and generate JWT tokens.
type JWTGenerator struct {
	database      jwtGeneratorDatabase
	accessChecker jwtGeneratorAccessChecker
	jwtService    jwtService

	user      *openfga.User
	mt        names.ModelTag
	ct        names.ControllerTag
	callCount atomic.Int32
}

// NewJWTGenerator returns a new JwtAuthorizer struct
func NewJWTGenerator(database jwtGeneratorDatabase, accessChecker jwtGeneratorAccessChecker, jwtService jwtService) JWTGenerator {
	return JWTGenerator{
		database:      database,
		accessChecker: accessChecker,
		jwtService:    jwtService,
	}
}

// SetTags implements TokenGenerator
func (auth *JWTGenerator) SetTags(mt names.ModelTag, ct names.ControllerTag) {
	auth.mt = mt
	auth.ct = ct
}

// SetTags implements TokenGenerator
func (auth *JWTGenerator) GetUser() names.UserTag {
	if auth.user != nil {
		return auth.user.ResourceTag()
	}
	return names.UserTag{}
}

// MakeLoginToken authorizes the user based on the provided login requests and returns
// a JWT containing claims about user's access to the controller, model (if applicable)
// and all clouds that the controller knows about.
func (auth *JWTGenerator) MakeLoginToken(ctx context.Context, user *openfga.User) ([]byte, error) {
	const op = errors.Op("jimm.MakeLoginToken")

	if user == nil {
		return nil, errors.E(op, "user not specified")
	}
	auth.user = user

	// Recreate the accessMapCache to prevent leaking permissions across multiple login requests.
	accessMap := make(map[names.Tag]openfga.Relation)
	var authErr error

	var modelAccess openfga.Relation
	if auth.mt.Id() == "" {
		return nil, errors.E(op, "model not set")
	}
	modelAccess, authErr = auth.accessChecker.GetUserModelAccess(ctx, auth.user, auth.mt)
	if authErr != nil {
		zapctx.Error(ctx, "model access check failed", zap.Error(authErr))
		return nil, authErr
	}
	accessMap[auth.mt] = modelAccess

	if auth.ct.Id() == "" {
		return nil, errors.E(op, "controller not set")
	}
	var controllerAccess openfga.Relation
	controllerAccess, authErr = auth.accessChecker.GetUserControllerAccess(ctx, auth.user, auth.ct)
	if authErr != nil {
		return nil, authErr
	}
	accessMap[auth.ct] = controllerAccess

	var ctl dbmodel.Controller
	ctl.SetTag(auth.ct)
	err := auth.database.GetController(ctx, &ctl)
	if err != nil {
		zapctx.Error(ctx, "failed to fetch controller", zap.Error(err))
		return nil, errors.E(op, "failed to fetch controller", err)
	}
	clouds := make(map[names.CloudTag]bool)
	for _, cloudRegion := range ctl.CloudRegions {
		clouds[cloudRegion.CloudRegion.Cloud.ResourceTag()] = true
	}
	for cloudTag := range clouds {
		cloudAccess, err := auth.accessChecker.GetUserCloudAccess(ctx, auth.user, cloudTag)
		if err != nil {
			zapctx.Error(ctx, "cloud access check failed", zap.Error(err))
			return nil, errors.E(op, "failed to check user's cloud access", err)
		}
		accessMap[cloudTag] = cloudAccess
	}

	return auth.jwtService.NewJWT(ctx, jimmjwx.JWTParams{
		Controller: auth.ct.Id(),
		User:       auth.user.Tag().String(),
		Access:     toJujuPermissionMap(accessMap),
	})
}

// MakeToken assumes MakeLoginToken has already been called and checks the permissions
// specified in the permissionMap. If the logged in user has all the desired permissions,
// a JWT will be returned with assertions confirming all those permissions.
func (auth *JWTGenerator) MakeToken(ctx context.Context, permissionMap map[string]interface{}) ([]byte, error) {
	const op = errors.Op("jimm.MakeToken")

	if auth.callCount.Load() >= 10 {
		return nil, errors.E(op, "Permission check limit exceeded")
	}
	auth.callCount.Add(1)
	if auth.user == nil {
		return nil, errors.E(op, "User authorization missing.")
	}

	accessMap := make(map[names.Tag]openfga.Relation)

	if permissionMap != nil {
		desiredPerms, err := auth.validatePermissions(permissionMap)
		if err != nil {
			return nil, err
		}
		accessMap, err = auth.accessChecker.CheckBatchPermissions(ctx, auth.user, desiredPerms)
		if err != nil {
			return nil, err
		}
	}

	jwt, err := auth.jwtService.NewJWT(ctx, jimmjwx.JWTParams{
		Controller: auth.ct.Id(),
		User:       auth.user.Tag().String(),
		Access:     toJujuPermissionMap(accessMap),
	})
	if err != nil {
		return nil, err
	}

	return jwt, nil
}

func toJujuPermissionMap(in map[names.Tag]openfga.Relation) map[names.Tag]permission.Access {
	permissions := make(map[names.Tag]permission.Access)
	for key, value := range in {
		permissions[key] = ofganames.ToJujuPermission(value)
	}
	return permissions
}

func (auth *JWTGenerator) validatePermissions(permissionMap map[string]interface{}) (map[names.Tag]openfga.Relation, error) {
	desiredPermissions := make(map[names.Tag]openfga.Relation)
	for key, val := range permissionMap {
		stringVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("failed to get permission assertion: expected %T, got %T", stringVal, val)
		}
		tag, err := names.ParseTag(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tag %s", key)
		}
		relation, err := ofganames.ToOpenFGARelation(stringVal)
		if err != nil {
			return nil, fmt.Errorf("failed to parse relation %s: %s", stringVal, err)
		}
		desiredPermissions[tag] = relation
	}
	return desiredPermissions, nil
}
