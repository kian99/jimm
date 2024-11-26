// Copyright 2024 Canonical.

package jujuapi

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/canonical/jimm/v3/internal/jimm"
	qt "github.com/frankban/quicktest"
)

func TestControllerPing(t *testing.T) {
	c := qt.New(t)
	j := jimm.JIMM{}
	r := newControllerRoot(&j, Params{}, "")
	defer r.cleanup()
	var calls uint32
	r.setPingF(func() { atomic.AddUint32(&calls, 1) })
	m, err := r.FindMethod("Pinger", 1, "Ping")
	c.Assert(err, qt.IsNil)
	_, err = m.Call(context.Background(), "", reflect.Value{})
	c.Assert(err, qt.IsNil)
	c.Check(atomic.LoadUint32(&calls), qt.Equals, uint32(1))
}
