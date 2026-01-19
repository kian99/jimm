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

func TestUnregisterController(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().RemoveController(gomock.Any()).Return(params.ControllerInfo{
		Name: "old-controller",
	}, nil)
	mockClient.EXPECT().Close().Return(nil)

	// Create command with mocked dependencies
	command := &unregisterControllerCommand{
		store: mockStore,
		params: params.RemoveControllerRequest{
			Name:  "old-controller",
			Force: false,
		},
		unregisterControllerAPIFunc: func() (JIMMAPI, error) {
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

func TestUnregisterControllerAPIError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to connect")

	command := &unregisterControllerCommand{
		store: mockStore,
		params: params.RemoveControllerRequest{
			Name:  "old-controller",
			Force: false,
		},
		unregisterControllerAPIFunc: func() (JIMMAPI, error) {
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

func TestUnregisterControllerRemoveError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to remove controller")

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().RemoveController(gomock.Any()).Return(params.ControllerInfo{}, expectedErr)
	mockClient.EXPECT().Close().Return(nil)

	command := &unregisterControllerCommand{
		store: mockStore,
		params: params.RemoveControllerRequest{
			Name:  "old-controller",
			Force: false,
		},
		unregisterControllerAPIFunc: func() (JIMMAPI, error) {
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
