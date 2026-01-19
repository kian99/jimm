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

func TestListControllers(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().ListControllers().Return([]params.ControllerInfo{
		{
			Name: "controller-1",
			UUID: "deadbeef-1bad-500d-9000-4b1d0d06f00d",
		},
	}, nil)
	mockClient.EXPECT().Close().Return(nil)

	// Create command with mocked dependencies
	command := &listControllersCommand{
		store: mockStore,
		listAPIFunc: func() (JIMMAPI, error) {
			return mockClient, nil
		},
	}

	// Set up controller command base
	command.SetClientStore(mockStore)

	// Run the command
	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNil)
}

func TestListControllersError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations - client creation fails
	expectedErr := errors.New("failed to get controller")

	// Create command with mocked dependencies
	command := &listControllersCommand{
		store: mockStore,
		listAPIFunc: func() (JIMMAPI, error) {
			return nil, expectedErr
		},
	}

	// Set up controller command base
	command.SetClientStore(mockStore)

	// Run the command - should fail
	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNotNil)
}

