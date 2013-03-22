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

func (d *D) TestMatchesOtherName(c *C) {
	dep := &Dependency{ImportPath: "gonuts.io/debug/crazy", Version: "0.0.1"}
	c.Check(dep.Matches("gonuts.io", d.nut), Equals, false)
}

func (d *D) TestMatchesExact(c *C) {
	dep := &Dependency{ImportPath: "gonuts.io/debug/test_nut1", Version: "0.0.1"}
	c.Check(dep.Matches("gonuts.io", d.nut), Equals, true)

	for _, v := range []string{"0.0.9", "0.9.1", "9.0.1"} {
		dep := &Dependency{ImportPath: "gonuts.io/debug/test_nut1", Version: v}
		c.Check(dep.Matches("gonuts.io", d.nut), Equals, false)
	}
}

func (d *D) TestMatchesWildcard(c *C) {
	for _, v := range []string{"*.*.*", "0.*.*", "0.0.*"} {
		dep := &Dependency{ImportPath: "gonuts.io/debug/test_nut1", Version: v}
		c.Check(dep.Matches("gonuts.io", d.nut), Equals, true, Commentf("Dependency %q should match %v", dep, d.nut))
	}

	for _, v := range []string{"9.*.*", "*.9.*", "*.*.9"} {
		dep := &Dependency{ImportPath: "gonuts.io/debug/test_nut1", Version: v}
		c.Check(dep.Matches("gonuts.io", d.nut), Equals, false, Commentf("Dependency %q should not match %v", dep, d.nut))
	}
}
