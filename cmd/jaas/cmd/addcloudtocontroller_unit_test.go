// Copyright 2025 Canonical.

package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/juju/cmd/v3"
	qt "github.com/frankban/quicktest"
	"github.com/juju/juju/cloud"
	"go.uber.org/mock/gomock"

	"github.com/canonical/jimm/v3/cmd/jaas/cmd/mocks"
	"github.com/canonical/jimm/v3/pkg/api/params"
)

func TestAddCloudToController(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().AddCloudToController(gomock.Any()).Return(nil)
	mockClient.EXPECT().Close().Return(nil)

	// Mock cloud by name function
	cloudByNameFunc := func(cloudName string) (*cloud.Cloud, error) {
		return &cloud.Cloud{
			Name:      "test-cloud",
			Type:      "kubernetes",
			AuthTypes: []cloud.AuthType{"certificate"},
			Regions: []cloud.Region{
				{Name: "default"},
			},
		}, nil
	}

	// Create command with mocked dependencies
	command := &addCloudToControllerCommand{
		store:                       mockStore,
		cloudByNameFunc:             cloudByNameFunc,
		dstControllerName:           "controller-1",
		cloudName:                   "test-cloud",
		addCloudToControllerAPIFunc: func() (JIMMAPI, error) {
			return mockClient, nil
		},
	}

	command.SetClientStore(mockStore)

	// Run the command
	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNil)
}

func TestAddCloudToControllerAPIError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to connect")

	// Create command with mocked dependencies
	command := &addCloudToControllerCommand{
		store:             mockStore,
		dstControllerName: "controller-1",
		cloudName:         "test-cloud",
		cloudByNameFunc: func(cloudName string) (*cloud.Cloud, error) {
			return &cloud.Cloud{
				Name:      "test-cloud",
				Type:      "kubernetes",
				AuthTypes: []cloud.AuthType{"certificate"},
				Regions: []cloud.Region{
					{Name: "default"},
				},
			}, nil
		},
		addCloudToControllerAPIFunc: func() (JIMMAPI, error) {
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

func TestAddCloudToControllerCloudNotFound(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockClientStore(ctrl)

	// Create command with cloud lookup failure
	command := &addCloudToControllerCommand{
		store:             mockStore,
		dstControllerName: "controller-1",
		cloudName:         "test-cloud",
		cloudByNameFunc: func(cloudName string) (*cloud.Cloud, error) {
			return nil, errors.New("cloud not found")
		},
		addCloudToControllerAPIFunc: func() (JIMMAPI, error) {
			return nil, errors.New("should not be called")
		},
	}

	command.SetClientStore(mockStore)

	ctx := &cmd.Context{
		Context: context.Background(),
	}

	err := command.Run(ctx)
	c.Assert(err, qt.IsNotNil)
	c.Assert(err.Error(), qt.Contains, "could not find existing cloud")
}

func TestAddCloudToControllerAddError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to add cloud")

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().AddCloudToController(gomock.Any()).Return(expectedErr)
	mockClient.EXPECT().Close().Return(nil)

	command := &addCloudToControllerCommand{
		store:             mockStore,
		dstControllerName: "controller-1",
		cloudName:         "test-cloud",
		cloudByNameFunc: func(cloudName string) (*cloud.Cloud, error) {
			return &cloud.Cloud{
				Name:      "test-cloud",
				Type:      "kubernetes",
				AuthTypes: []cloud.AuthType{"certificate"},
				Regions: []cloud.Region{
					{Name: "default"},
				},
			}, nil
		},
		addCloudToControllerAPIFunc: func() (JIMMAPI, error) {
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
