package nut_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	. "."
	. "launchpad.net/gocheck"
)

type N struct {
	nf *NutFile
}

var _ = Suite(&N{})

func (f *N) SetUpTest(c *C) {
	file, err := os.Open("../test_nut1/test_nut1-0.0.1.nut")
	c.Assert(err, IsNil)
	defer file.Close()

	fi, err := file.Stat()
	c.Assert(err, IsNil)

	nf := new(NutFile)
	n, err := nf.ReadFrom(file)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, fi.Size())
	f.nf = nf
}

func (f *N) TestNutFile(c *C) {
	c.Check(f.nf.Spec.Version.String(), Equals, "0.0.1")
	c.Check(f.nf.Version.String(), Equals, "0.0.1")
	c.Check(f.nf.Vendor, Equals, "debug")
	c.Check(f.nf.Package.Name, Equals, "test_nut1")
	c.Check(f.nf.Name, Equals, "test_nut1")
	c.Check(f.nf.FileName(), Equals, "test_nut1-0.0.1.nut")
	c.Check(f.nf.FilePath("prefix"), Equals, filepath.FromSlash("prefix/debug/test_nut1-0.0.1.nut"))
	c.Check(f.nf.ImportPath("prefix"), Equals, "prefix/debug/test_nut1")
	c.Check(f.nf.Doc, Equals, "Package test_nut1 is used to test nut.")
	c.Check(f.nf.GoFiles, DeepEquals, []string{"test_nut1.go", fmt.Sprintf("test_nut1_%s.go", runtime.GOOS)})

	c.Check(len(f.nf.Reader.File), Equals, 11)
	names := make([]string, 0, 11)
	for _, f := range f.nf.Reader.File {
		names = append(names, f.Name)
	}
	c.Check([]string{"test_nut1.go", "test_nut1_darwin.go", "test_nut1_freebsd.go", "test_nut1_linux.go", "test_nut1_netbsd.go",
		"test_nut1_openbsd.go", "test_nut1_plan9.go", "test_nut1_windows.go", "README", "LICENSE", "nut.json"},
		DeepEquals, names)
}

func (f *N) TestNutDir(c *C) {
	pwd, err := os.Getwd()
	c.Assert(err, IsNil)
	c.Assert(os.Chdir("../test_nut1"), IsNil)
	defer func() {
		c.Assert(os.Chdir(pwd), IsNil)
	}()

	nut := new(Nut)
	err = nut.ReadFrom(".")
	c.Check(err, IsNil)
	c.Check(nut.Spec, DeepEquals, f.nf.Spec)
	c.Check(nut.Package, DeepEquals, f.nf.Package)
}
