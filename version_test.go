package nut_test

import (
	"encoding/json"

	. "."
	. "launchpad.net/gocheck"
)

type V struct {
	versions []string
}

var _ = Suite(&V{})

func (f *V) SetUpTest(c *C) {
	f.versions = []string{
		"0.0.0", "0.0.1", "0.0.2",
		"0.1.0", "0.1.1", "0.1.2",
		"1.0.0", "1.0.1", "1.0.2",
		"1.1.0", "1.1.1", "1.1.2",
		"1.1.10", "1.10.1", "10.1.1", "10.10.10",
	}
}

func (f *V) TestNew(c *C) {
	for _, vs := range f.versions {
		v, err := NewVersion(vs)
		c.Check(err, Equals, nil)
		c.Check(v.String(), Equals, vs)
	}
}

func (f *V) TestLess(c *C) {
	for i, vi := range f.versions {
		left, err := NewVersion(vi)
		c.Assert(err, Equals, nil)

		for _, vj := range f.versions[:i] {
			right, err := NewVersion(vj)
			c.Assert(err, Equals, nil)
			c.Check(left.Less(right), Equals, false, Commentf("Expected %s >= %s", left, right))
		}
		for _, vj := range f.versions[i+1:] {
			right, err := NewVersion(vj)
			c.Assert(err, Equals, nil)
			c.Check(left.Less(right), Equals, true, Commentf("Expected %s < %s", left, right))
		}
	}
}

func (f *V) TestJSON(c *C) {
	for _, vs := range f.versions {
		v, err := NewVersion(vs)
		c.Assert(err, Equals, nil)

		b, err := json.Marshal(v)
		c.Check(string(b), Equals, `"`+vs+`"`)
		c.Assert(err, Equals, nil)

		v2 := new(Version)
		err = json.Unmarshal(b, v2)
		c.Check(v2, DeepEquals, v)
		c.Assert(err, Equals, nil)
	}
}
