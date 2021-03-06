// Copyright 2017 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package common_test

import (
	"github.com/juju/errors"
	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/environs"
	"github.com/juju/juju/provider/common"
)

type ErrorsSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&ErrorsSuite{})

func (*ErrorsSuite) TestWrapZoneIndependentError(c *gc.C) {
	err1 := errors.New("foo")
	err2 := errors.Annotate(err1, "bar")
	wrapped := common.ZoneIndependentError(err2)
	c.Assert(wrapped, jc.Satisfies, environs.IsAvailabilityZoneIndependent)
	c.Assert(wrapped, gc.ErrorMatches, "bar: foo")

	stack := errors.ErrorStack(wrapped)
	c.Assert(stack, gc.Matches, `
github.com/juju/juju/provider/common/errors_test.go:.*: foo
github.com/juju/juju/provider/common/errors_test.go:.*: bar
github.com/juju/juju/provider/common/errors_test.go:.*: bar: foo`[1:])
}
