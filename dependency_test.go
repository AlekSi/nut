package nut_test

import (
	"os"

	. "."
	. "launchpad.net/gocheck"
)

type D struct {
	nut *Nut
}

var _ = Suite(&D{})

func (d *D) SetUpTest(c *C) {
	file, err := os.Open("../test_nut1/test_nut1-0.0.1.nut")
	c.Assert(err, IsNil)
	defer file.Close()

	nf := new(NutFile)
	_, err = nf.ReadFrom(file)
	c.Assert(err, IsNil)
	d.nut = &nf.Nut
}

func (d *D) TestNew(c *C) {
	for _, v := range []string{"0.0.1", "*.*.*", "1.>=5.*", "git:e1a6adc"} {
		_, err := NewDependency("gonuts.io/debug/crazy", v)
		c.Check(err, IsNil)
	}

	for _, v := range []string{"0.1", "*", "1-2-3", "1.2.>=*", "cvs:6"} {
		_, err := NewDependency("gonuts.io/debug/crazy", v)
		c.Check(err, Not(IsNil))
	}
}

func (d *D) TestMatchesOtherName(c *C) {
	dep, err := NewDependency("gonuts.io/debug/crazy", "0.0.1")
	c.Check(err, IsNil)
	c.Check(dep.Matches("gonuts.io", d.nut), Equals, false)
}

func (d *D) TestMatchesExact(c *C) {
	dep, err := NewDependency("gonuts.io/debug/test_nut1", "0.0.1")
	c.Check(err, IsNil)
	c.Check(dep.OnNut(), Equals, true)
	c.Check(dep.IsStrict(), Equals, true)
	c.Check(dep.Matches("gonuts.io", d.nut), Equals, true)

	for _, v := range []string{"0.0.9", "0.9.1", "9.0.1"} {
		dep, err = NewDependency("gonuts.io/debug/test_nut1", v)
		c.Check(err, IsNil)
		c.Check(dep.OnNut(), Equals, true)
		c.Check(dep.IsStrict(), Equals, true)
		c.Check(dep.Matches("gonuts.io", d.nut), Equals, false)
	}
}

func (d *D) TestMatchesWildcard(c *C) {
	for _, v := range []string{"*.*.*", "0.*.*", "0.0.*"} {
		dep, err := NewDependency("gonuts.io/debug/test_nut1", v)
		c.Check(err, IsNil)
		c.Check(dep.OnNut(), Equals, true)
		c.Check(dep.IsStrict(), Equals, false)
		c.Check(dep.Matches("gonuts.io", d.nut), Equals, true, Commentf("Dependency %q should match %v", dep, d.nut))
	}

	for _, v := range []string{"9.*.*", "*.9.*", "*.*.9"} {
		dep, err := NewDependency("gonuts.io/debug/test_nut1", v)
		c.Check(err, IsNil)
		c.Check(dep.OnNut(), Equals, true)
		c.Check(dep.IsStrict(), Equals, false)
		c.Check(dep.Matches("gonuts.io", d.nut), Equals, false, Commentf("Dependency %q should not match %v", dep, d.nut))
	}
}

func (d *D) TestMatchesMoreEqual(c *C) {
	for _, v := range []string{"0.0.>=0", "0.0.>=1", ">=0.>=0.>=0"} {
		dep, err := NewDependency("gonuts.io/debug/test_nut1", v)
		c.Check(err, IsNil)
		c.Check(dep.OnNut(), Equals, true)
		c.Check(dep.IsStrict(), Equals, false)
		c.Check(dep.Matches("gonuts.io", d.nut), Equals, true, Commentf("Dependency %q should match %v", dep, d.nut))
	}

	for _, v := range []string{">=9.0.1", "0.>=9.1", "0.0.>=9"} {
		dep, err := NewDependency("gonuts.io/debug/test_nut1", v)
		c.Check(err, IsNil)
		c.Check(dep.OnNut(), Equals, true)
		c.Check(dep.IsStrict(), Equals, false)
		c.Check(dep.Matches("gonuts.io", d.nut), Equals, false, Commentf("Dependency %q should not match %v", dep, d.nut))
	}
}

func (d *D) TestDependenciesAdd(c *C) {
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
