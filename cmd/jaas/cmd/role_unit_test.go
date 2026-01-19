// Copyright 2025 Canonical.

package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/juju/cmd/v3"
	qt "github.com/frankban/quicktest"
	"go.uber.org/mock/gomock"

	"github.com/canonical/jimm/v3/cmd/jaas/cmd/mocks"
	"github.com/canonical/jimm/v3/pkg/api/params"
)

func TestAddRole(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().AddRole(gomock.Any()).Return(params.AddRoleResponse{}, nil)
	mockClient.EXPECT().Close().Return(nil)

	// Create command with mocked dependencies
	command := &addRoleCommand{
		store: mockStore,
		name:  "test-role",
		addAPIFunc: func() (JIMMAPI, error) {
			return mockClient, nil
		},
	}

	command.SetClientStore(mockStore)

	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNil)
}

func TestAddRoleAPIError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to connect")

	command := &addRoleCommand{
		store: mockStore,
		name:  "test-role",
		addAPIFunc: func() (JIMMAPI, error) {
			return nil, expectedErr
		},
	}

	command.SetClientStore(mockStore)

	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNotNil)
}

func TestRenameRole(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().RenameRole(gomock.Any()).Return(nil)
	mockClient.EXPECT().Close().Return(nil)

	command := &renameRoleCommand{
		store:          mockStore,
		oldName:        "old-role",
		newName:        "new-role",
		renameAPIFunc: func() (JIMMAPI, error) {
			return mockClient, nil
		},
	}

	command.SetClientStore(mockStore)

	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNil)
}

func TestRemoveRole(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().RemoveRole(gomock.Any()).Return(nil)
	mockClient.EXPECT().Close().Return(nil)

	command := &removeRoleCommand{
		store: mockStore,
		name:  "test-role",
		removeAPIFunc: func() (JIMMAPI, error) {
			return mockClient, nil
		},
	}

	command.SetClientStore(mockStore)

	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNil)
}

func TestListRoles(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().ListRoles(gomock.Any()).Return([]params.Role{
		{Name: "role-1", UUID: "uuid-1"},
	}, nil)
	mockClient.EXPECT().Close().Return(nil)

	command := &listRolesCommand{
		store: mockStore,
		listAPIFunc: func() (JIMMAPI, error) {
			return mockClient, nil
		},
	}

	command.SetClientStore(mockStore)

	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNil)
}
