package nut

import (
	"archive/zip"
	"bytes"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Check package for errors and return them.
func CheckPackage(pack *build.Package) (errors []string) {
	// check name
	if strings.ToLower(pack.Name) != pack.Name {
		errors = append(errors, `Package name should be lower case.`)
	}
	if strings.HasPrefix(pack.Name, "_") {
		errors = append(errors, `Package name should not starts with "_".`)
	}
	if strings.HasSuffix(pack.Name, "_") {
		errors = append(errors, `Package name should not ends with "_".`)
	}
	if strings.HasSuffix(pack.Name, "_test") {
		errors = append(errors, `Package name should not ends with "_test".`)
	}

	// check doc summary
	r := regexp.MustCompile(fmt.Sprintf(`Package %s .+\.`, pack.Name))
	if !r.MatchString(pack.Doc) {
		errors = append(errors, fmt.Sprintf(`Package summary in code should be in form "Package %s ... ."`, pack.Name))
	}

	return
}

// Describes nut – a Go package with associated meta-information.
// It embeds Spec and build.Package to provide easy access to properties:
// Nut.Name instead of Nut.Package.Name, Nut.Version instead of Nut.Spec.Version.
type Nut struct {
	Spec
	build.Package
}

// Check nut for errors and return them. Calls Spec.Check() and CheckPackage().
func (nut *Nut) Check() (errors []string) {
	errors = nut.Spec.Check()
	errors = append(errors, CheckPackage(&nut.Package)...)
	return
}

// Returns canonical filename in format <name>-<version>.nut
func (nut *Nut) FileName() string {
	return fmt.Sprintf("%s-%s.nut", nut.Name, nut.Version)
}

// Returns canonical filepath in format <prefix>/<vendor>/<name>-<version>.nut
// (with "\" instead of "/" on Windows).
func (nut *Nut) FilePath(prefix string) string {
	return filepath.Join(prefix, nut.Vendor, nut.FileName())
}

// Returns canonical import path in format <prefix>/<vendor>/<name>
func (nut *Nut) ImportPath(prefix string) string {
	return fmt.Sprintf("%s/%s/%s", prefix, nut.Vendor, nut.Name)
}

// Read nut from directory: package from <dir> and spec from <dir>/<SpecFileName>.
func (nut *Nut) ReadFrom(dir string) (err error) {
	// This method is called ReadFrom to prevent code n.ReadFrom(r) from calling n.Spec.ReadFrom(r).

	// read package
	pack, err := build.ImportDir(dir, 0)
	if err != nil {
		return
	}
	nut.Package = *pack

	// read spec
	f, err := os.Open(filepath.Join(dir, SpecFileName))
	if err != nil {
		return
	}
	defer f.Close()
	_, err = nut.Spec.ReadFrom(f)
	return
}

// Describes .nut file (a ZIP archive).
type NutFile struct {
	Nut
	Reader *zip.Reader
}

// check interface
var (
	_ io.ReaderFrom = &NutFile{}
)

// Reads nut from specified file.
func (nf *NutFile) ReadFile(fileName string) (err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = nf.ReadFrom(f)
	return
}

// ReadFrom reads nut from r until EOF.
// The return value n is the number of bytes read.
// Any error except io.EOF encountered during the read is also returned.
// Implements io.ReaderFrom.
func (nf *NutFile) ReadFrom(r io.Reader) (n int64, err error) {
	var b []byte
	b, err = ioutil.ReadAll(r)
	n = int64(len(b))
	if err != nil {
		return
	}

	nf.Reader, err = zip.NewReader(bytes.NewReader(b), n)
	if err != nil {
		return
	}

	// read spec (typically the last file)
	var specReader io.ReadCloser
	for i := len(nf.Reader.File) - 1; i >= 0; i-- {
		file := nf.Reader.File[i]
		if file.Name == SpecFileName {
			specReader, err = file.Open()
			if err != nil {
				return
			}
			defer func() {
				e := specReader.Close()
				if err == nil { // don't hide original error
					err = e
				}
			}()
			break
		}
	}
	if specReader == nil {
		err = fmt.Errorf("NutFile.ReadFrom: %q not found", SpecFileName)
		return
	}
	spec := &nf.Spec
	_, err = spec.ReadFrom(specReader)
	if err != nil {
		return
	}

	// read package
	pack, err := nf.context().ImportDir(".", 0)
	if err != nil {
		return
	}
	nf.Package = *pack
	return
}

// byName implements sort.Interface.
type byName []os.FileInfo

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// Returns build.Context for given nut.
// Returned value may be used to call normal context methods Import and ImportDir
// to extract information about package in nut without unpacking it: name, doc, dependencies.
func (nf *NutFile) context() (ctxt *build.Context) {
	ctxt = new(build.Context)
	*ctxt = build.Default

	// FIXME path is ignored (multi-package nuts are not supported yet)
	ctxt.ReadDir = func(path string) ([]os.FileInfo, error) {
		// log.Printf("nf.ReadDir %q", path)

		fi := make([]os.FileInfo, len(nf.Reader.File))
		for i, f := range nf.Reader.File {
			fi[i] = f.FileInfo()
		}
		sort.Sort(byName(fi))
		return fi, nil
	}

	ctxt.OpenFile = func(path string) (io.ReadCloser, error) {
		// log.Printf("nf.OpenFile %q", path)

		for _, f := range nf.Reader.File {
			if f.Name == path {
				return f.Open()
			}
		}

		return nil, fmt.Errorf("NutFile.Context.OpenFile: %q not found", path)
	}

	return
}
