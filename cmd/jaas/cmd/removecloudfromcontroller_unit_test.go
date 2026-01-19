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

func TestRemoveCloudFromController(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().RemoveCloudFromController(gomock.Any()).Return(nil)
	mockClient.EXPECT().Close().Return(nil)

	// Create command with mocked dependencies
	command := &removeCloudFromControllerCommand{
		store:                mockStore,
		targetControllerName: "controller-1",
		cloudName:            "test-cloud",
		removeCloudFromControllerAPIFunc: func() (JIMMAPI, error) {
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

func TestRemoveCloudFromControllerAPIError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to connect")

	command := &removeCloudFromControllerCommand{
		store:                mockStore,
		targetControllerName: "controller-1",
		cloudName:            "test-cloud",
		removeCloudFromControllerAPIFunc: func() (JIMMAPI, error) {
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

func TestRemoveCloudFromControllerRemoveError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to remove cloud")

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().RemoveCloudFromController(gomock.Any()).Return(expectedErr)
	mockClient.EXPECT().Close().Return(nil)

	command := &removeCloudFromControllerCommand{
		store:                mockStore,
		targetControllerName: "controller-1",
		cloudName:            "test-cloud",
		removeCloudFromControllerAPIFunc: func() (JIMMAPI, error) {
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
