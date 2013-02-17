package nut_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	. "."
	. "launchpad.net/gocheck"
)

type N struct {
	f *os.File
	b *bytes.Buffer
}

var _ = Suite(&N{})

func (f *N) SetUpTest(c *C) {
	file, err := os.Open("../test_nut1/test_nut1-0.0.1.nut")
	c.Assert(err, IsNil)
	f.f = file

	b, err := ioutil.ReadAll(f.f)
	c.Assert(err, IsNil)
	f.b = bytes.NewBuffer(b)

	_, err = file.Seek(0, 0)
	c.Assert(err, IsNil)
}

func (f *N) TestNutFile(c *C) {
	nf := new(NutFile)
	_, err := nf.ReadFrom(f.f)
	c.Assert(err, IsNil)

	c.Check(nf.Spec.Version.String(), Equals, "0.0.1")
	c.Check(nf.Version.String(), Equals, "0.0.1")
	c.Check(nf.Package.Name, Equals, "test_nut1")
	c.Check(nf.Name, Equals, "test_nut1")
	c.Check(nf.FileName(), Equals, "test_nut1-0.0.1.nut")
	c.Check(nf.FilePath("prefix"), Equals, filepath.FromSlash("prefix/test_nut1-0.0.1.nut"))
	c.Check(nf.ImportPath("prefix"), Equals, "prefix/test_nut1/0.0.1")
	c.Check(nf.Doc, Equals, "Package test_nut1 is used to test nut.")
	c.Check(nf.GoFiles, DeepEquals, []string{"test_nut1.go", fmt.Sprintf("test_nut1_%s.go", runtime.GOOS)})

	c.Check(len(nf.Reader.File), Equals, 11)
	names := make([]string, 0, 11)
	for _, f := range nf.Reader.File {
		names = append(names, f.Name)
	}
	c.Check([]string{"test_nut1.go", "test_nut1_darwin.go", "test_nut1_freebsd.go", "test_nut1_linux.go", "test_nut1_netbsd.go",
		"test_nut1_openbsd.go", "test_nut1_plan9.go", "test_nut1_windows.go", "README", "LICENSE", "nut.json"},
		DeepEquals, names)
}
