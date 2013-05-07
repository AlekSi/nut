package nut_test

import (
	. "."
	. "launchpad.net/gocheck"
)

type Ds struct{}

var _ = Suite(&Ds{})

func (*Ds) TestDependenciesAdd(c *C) {
	deps := NewDependencies()
	for _, v := range []string{"1.>=1.*", "1.>=2.*", "1.*.*"} {
		dep, err := NewDependency("gonuts.io/debug/crazy", v)
		c.Check(err, IsNil)
		err = deps.Add(dep)
		c.Check(err, IsNil)
	}
	c.Check(deps.Get("gonuts.io/debug/crazy").String(), Equals, "gonuts.io/debug/crazy (1.>=2.*)")

	dep, err := NewDependency("gonuts.io/debug/crazy", "2.*.*")
	c.Check(err, IsNil)
	err = deps.Add(dep)
	c.Check(err, FitsTypeOf, &AddDependencyError{})
	c.Check(err.Error(), Equals, "Can't add gonuts.io/debug/crazy (2.*.*) to existing dependecy gonuts.io/debug/crazy (1.>=2.*)")
	c.Check(deps.Get("gonuts.io/debug/crazy").String(), Equals, "gonuts.io/debug/crazy (1.>=2.*)")
}
