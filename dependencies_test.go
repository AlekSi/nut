package nut_test

import (
	. "."
	. "launchpad.net/gocheck"
)

type Ds struct {
	deps Dependencies
}

var _ = Suite(&Ds{})

func (ds *Ds) SetUpTest(c *C) {
	ds.deps.Clear()
}

func (ds *Ds) TestDependenciesAdd(c *C) {
	for _, v := range []string{"1.>=1.*", "1.>=2.*", "1.*.*"} {
		dep, err := NewDependency("gonuts.io/debug/crazy", v)
		c.Check(err, IsNil)
		err = ds.deps.Add(dep)
		c.Check(err, IsNil)
	}
	c.Check(ds.deps.Get("gonuts.io/debug/crazy").String(), Equals, "gonuts.io/debug/crazy (1.>=2.*)")

	dep, err := NewDependency("gonuts.io/debug/crazy", "2.*.*")
	c.Check(err, IsNil)
	err = ds.deps.Add(dep)
	c.Check(err, FitsTypeOf, &AddDependencyError{})
	c.Check(err.Error(), Equals, "Can't add gonuts.io/debug/crazy (2.*.*) to existing dependecy gonuts.io/debug/crazy (1.>=2.*)")
	c.Check(ds.deps.Get("gonuts.io/debug/crazy").String(), Equals, "gonuts.io/debug/crazy (1.>=2.*)")
}
