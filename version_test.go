package nut_test

import (
	"encoding/json"

	. "."
	. "launchpad.net/gocheck"
)

type V struct {
	versions    []string
	badVersions []string
}

var _ = Suite(&V{})

func (f *V) SetUpSuite(c *C) {
	f.versions = []string{
		"0.0.0", "0.0.1", "0.0.2",
		"0.1.0", "0.1.1", "0.1.2",
		"1.0.0", "1.0.1", "1.0.2",
		"1.1.0", "1.1.1", "1.1.2",
		"1.1.10", "1.10.1", "10.1.1", "10.10.10",
	}
	f.badVersions = []string{
		"1.0.-1",
	}
}

func (f *V) TestNew(c *C) {
	for _, vs := range f.versions {
		v, err := NewVersion(vs)
		c.Check(err, IsNil)
		c.Check(v.String(), Equals, vs)
	}

	for _, vs := range f.badVersions {
		v, err := NewVersion(vs)
		c.Check(err, Not(IsNil))
		c.Check(v.String(), Equals, "0.0.0")
	}
}

func (f *V) TestLess(c *C) {
	for i, vi := range f.versions {
		left, err := NewVersion(vi)
		c.Assert(err, IsNil)

		for _, vj := range f.versions[:i] {
			right, err := NewVersion(vj)
			c.Assert(err, IsNil)
			c.Check(left.Less(right), Equals, false, Commentf("Expected %s >= %s", left, right))
		}
		for _, vj := range f.versions[i+1:] {
			right, err := NewVersion(vj)
			c.Assert(err, IsNil)
			c.Check(left.Less(right), Equals, true, Commentf("Expected %s < %s", left, right))
		}
	}
}

func (f *V) TestJSON(c *C) {
	for _, vs := range f.versions {
		v, err := NewVersion(vs)
		c.Assert(err, IsNil)

		b, err := json.Marshal(v)
		c.Check(string(b), Equals, `"`+vs+`"`)
		c.Assert(err, IsNil)

		v2 := new(Version)
		err = json.Unmarshal(b, v2)
		c.Check(v2, DeepEquals, v)
		c.Assert(err, IsNil)
	}
}
