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

func TestAddRelation(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().AddRelation(gomock.Any()).Return(nil)
	mockClient.EXPECT().Close().Return(nil)

	// Create command with mocked dependencies
	command := &addRelationCommand{
		store: mockStore,
		tuples: []params.RelationshipTuple{
			{
				Object:       "user-alice",
				Relation:     "member",
				TargetObject: "group-test",
			},
		},
		addRelationAPIFunc: func() (JIMMAPI, error) {
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

func TestRemoveRelation(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().RemoveRelation(gomock.Any()).Return(nil)
	mockClient.EXPECT().Close().Return(nil)

	command := &removeRelationCommand{
		store: mockStore,
		tuples: []params.RelationshipTuple{
			{
				Object:       "user-alice",
				Relation:     "member",
				TargetObject: "group-test",
			},
		},
		removeRelationAPIFunc: func() (JIMMAPI, error) {
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

func TestCheckRelation(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().CheckRelation(gomock.Any()).Return(params.CheckRelationResponse{
		Allowed: true,
	}, nil)
	mockClient.EXPECT().Close().Return(nil)

	command := &checkRelationCommand{
		store:        mockStore,
		object:       "user-alice",
		relation:     "member",
		targetObject: "group-test",
		checkRelationAPIFunc: func() (JIMMAPI, error) {
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

func TestListRelations(t *testing.T) {
	c := qt.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockJIMMAPI(ctrl)
	mockStore := mocks.NewMockClientStore(ctrl)

	// Setup expectations
	mockStore.EXPECT().CurrentController().Return("test-controller", nil)
	mockClient.EXPECT().ListRelationshipTuples(gomock.Any()).Return(&params.ListRelationshipTuplesResponse{
		Tuples: []params.RelationshipTuple{
			{
				Object:       "user-alice",
				Relation:     "member",
				TargetObject: "group-test",
			},
		},
	}, nil)
	mockClient.EXPECT().Close().Return(nil)

	command := &listRelationsCommand{
		store:        mockStore,
		object:       "user-alice",
		relation:     "member",
		targetObject: "group-test",
		listRelationsAPIFunc: func() (JIMMAPI, error) {
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
