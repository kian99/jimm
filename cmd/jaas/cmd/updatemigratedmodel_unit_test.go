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

func TestUpdateMigratedModel(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().UpdateMigratedModel(gomock.Any()).Return(nil)
	mockClient.EXPECT().Close().Return(nil)

	// Create command with mocked dependencies
	command := &updateMigratedModelCommand{
		store: mockStore,
		req: params.UpdateMigratedModelRequest{
			TargetController: "controller-2",
			ModelTag:         "model-e0bf3abf-7029-4e48-9c26-68a7b6e02947",
		},
		updateFunc: func() (JIMMAPI, error) {
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

func TestUpdateMigratedModelAPIError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to connect")

	command := &updateMigratedModelCommand{
		store: mockStore,
		req: params.UpdateMigratedModelRequest{
			TargetController: "controller-2",
			ModelTag:         "model-e0bf3abf-7029-4e48-9c26-68a7b6e02947",
		},
		updateFunc: func() (JIMMAPI, error) {
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

func TestUpdateMigratedModelError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to update")

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().UpdateMigratedModel(gomock.Any()).Return(expectedErr)
	mockClient.EXPECT().Close().Return(nil)

	command := &updateMigratedModelCommand{
		store: mockStore,
		req: params.UpdateMigratedModelRequest{
			TargetController: "controller-2",
			ModelTag:         "model-e0bf3abf-7029-4e48-9c26-68a7b6e02947",
		},
		updateFunc: func() (JIMMAPI, error) {
			return mockClient, nil
		},
	}

	command.SetClientStore(mockStore)

	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNotNil)
}
