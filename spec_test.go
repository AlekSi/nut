package nut_test

import (
	"bytes"
	"io/ioutil"
	"os"

	. "."
	. "launchpad.net/gocheck"
)

type S struct {
	f *os.File
	b *bytes.Buffer
}

var _ = Suite(&S{})

func (f *S) SetUpTest(c *C) {
	file, err := os.Open("../test_nut1/nut.json")
	c.Assert(err, IsNil)
	f.f = file

	b, err := ioutil.ReadAll(f.f)
	c.Assert(err, IsNil)
	f.b = bytes.NewBuffer(b)

	_, err = file.Seek(0, 0)
	c.Assert(err, IsNil)
}

func (f *S) TestReadWrite(c *C) {
	spec := new(Spec)

	n, err := spec.ReadFrom(f.f)
	c.Check(n, Equals, int64(f.b.Len()))
	c.Assert(err, IsNil)

	c.Check(spec.Version.String(), Equals, "0.0.1")
	c.Check(spec.Vendor, Equals, "AlekSi")
	c.Check(len(spec.Authors), Equals, 1)
	c.Check(spec.Authors[0], Equals, Person{FullName: "Alexey Palazhchenko", Email: "alexey.palazhchenko@gmail.com"})
	c.Check(len(spec.ExtraFiles), Equals, 2)
	c.Check(spec.ExtraFiles[0], Equals, "README")
	c.Check(spec.ExtraFiles[1], Equals, "LICENSE")

	buf := new(bytes.Buffer)
	n2, err := spec.WriteTo(buf)
	c.Check(n, Equals, n2)
	c.Check(buf.String(), Equals, f.b.String())
	c.Assert(err, IsNil)
}
