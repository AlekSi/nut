package nut_test

import (
	"bytes"
	"io/ioutil"
	"os"

	. "."
	. "launchpad.net/gocheck"
)

type S struct {
	s *Spec
	b *bytes.Buffer
}

var _ = Suite(&S{})

func (f *S) SetUpTest(c *C) {
	file, err := os.Open("../test_nut1/nut.json")
	c.Assert(err, IsNil)
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	c.Assert(err, IsNil)
	f.b = bytes.NewBuffer(b)
	file.Seek(0, 0)

	s := new(Spec)
	n, err := s.ReadFrom(file)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, int64(f.b.Len()))
	f.s = s
}

func (f *S) TestReadFromWriteTo(c *C) {
	c.Check(f.s.Version.String(), Equals, "0.0.1")
	c.Check(f.s.Vendor, Equals, "debug")
	c.Check(len(f.s.Authors), Equals, 1)
	c.Check(f.s.Authors[0], Equals, Person{FullName: "Alexey Palazhchenko", Email: "alexey.palazhchenko@gmail.com"})
	c.Check(len(f.s.ExtraFiles), Equals, 2)
	c.Check(f.s.ExtraFiles[0], Equals, "README")
	c.Check(f.s.ExtraFiles[1], Equals, "LICENSE")

	buf := new(bytes.Buffer)
	n, err := f.s.WriteTo(buf)
	c.Assert(err, IsNil)
	c.Check(n, Equals, int64(f.b.Len()))
	c.Check(buf.String(), Equals, f.b.String())
}

func (f *S) TestReadFile(c *C) {
	s := new(Spec)
	err := s.ReadFile("../test_nut1/nut.json")
	c.Check(err, IsNil)
	c.Check(s, DeepEquals, f.s)
}
