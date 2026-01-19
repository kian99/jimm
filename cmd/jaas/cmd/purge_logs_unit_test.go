// Copyright 2025 Canonical.

package cmd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/juju/cmd/v3"
	qt "github.com/frankban/quicktest"
	"go.uber.org/mock/gomock"

	"github.com/canonical/jimm/v3/cmd/jaas/cmd/mocks"
	"github.com/canonical/jimm/v3/pkg/api/params"
)

func TestPurgeLogs(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().PurgeLogs(gomock.Any()).Return(&params.PurgeLogsResponse{
		DeletedCount: 100,
	}, nil)
	mockClient.EXPECT().Close().Return(nil)

	// Create command with mocked dependencies
	command := &purgeLogsCommand{
		store: mockStore,
		date:  time.Date(2021, 2, 3, 0, 0, 0, 0, time.UTC),
		purgeLogsAPIFunc: func() (JIMMAPI, error) {
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

func TestPurgeLogsAPIError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("failed to connect")

	command := &purgeLogsCommand{
		store: mockStore,
		date:  time.Date(2021, 2, 3, 0, 0, 0, 0, time.UTC),
		purgeLogsAPIFunc: func() (JIMMAPI, error) {
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

func TestPurgeLogsError(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	expectedErr := errors.New("purge failed")

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().PurgeLogs(gomock.Any()).Return(nil, expectedErr)
	mockClient.EXPECT().Close().Return(nil)

	command := &purgeLogsCommand{
		store: mockStore,
		date:  time.Date(2021, 2, 3, 0, 0, 0, 0, time.UTC),
		purgeLogsAPIFunc: func() (JIMMAPI, error) {
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
