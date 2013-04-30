package nut

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// Describes part of nut meta-information, stored in file nut.json.
type Spec struct {
	Version    Version
	Vendor     string
	Authors    []Person
	ExtraFiles []string
	Homepage   string
}

// Describes nut author.
type Person struct {
	FullName string
	Email    string `json:",omitempty"`
}

const (
	ExampleFullName = "Crazy Nutter"
	ExampleEmail    = "crazy.nutter@gonuts.io"
	SpecFileName    = "nut.json"
)

var VendorRegexp = regexp.MustCompile(`^[0-9a-z][0-9a-z_-]*$`)

// check interface
var (
	_ io.ReaderFrom = &Spec{}
	_ io.WriterTo   = &Spec{}
)

// Reads spec from specified file.
func (spec *Spec) ReadFile(fileName string) (err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = spec.ReadFrom(f)
	return
}

// ReadFrom reads spec from r until EOF.
// The return value n is the number of bytes read.
// Any error except io.EOF encountered during the read is also returned.
// Implements io.ReaderFrom.
func (spec *Spec) ReadFrom(r io.Reader) (n int64, err error) {
	var b []byte
	b, err = ioutil.ReadAll(r)
	n = int64(len(b))
	if err != nil {
		return
	}

	err = json.Unmarshal(b, spec)
	return
}

// WriteTo writes spec to w.
// The return value n is the number of bytes written.
// Any error encountered during the write is also returned.
// Implements io.WriterTo.
func (spec *Spec) WriteTo(w io.Writer) (n int64, err error) {
	var b []byte
	b, err = json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return
	}

	b = append(b, '\n')
	n1, err := w.Write(b)
	n = int64(n1)
	return
}

// Checks spec for errors and return them.
func (spec *Spec) Check() (errors []string) {
	// check version
	if spec.Version.String() == "0.0.0" {
		errors = append(errors, fmt.Sprintf("Version %q is invalid.", spec.Version))
	}

	// check vendor
	if !VendorRegexp.MatchString(spec.Vendor) {
		errors = append(errors, fmt.Sprintf(`Vendor should contain only lower word characters (match "%s").`, VendorRegexp))
	}

	// author should be specified
	if len(spec.Authors) == 0 {
		errors = append(errors, "No authors given.")
	} else {
		for _, a := range spec.Authors {
			if a.FullName == ExampleFullName {
				errors = append(errors, fmt.Sprintf("%q is not a real person.", a.FullName))
			}
		}
	}

	// check license
	licenseFound := false
	for _, f := range spec.ExtraFiles {
		f = strings.ToLower(f)
		if strings.Contains(f, "license") || strings.Contains(f, "licence") || strings.Contains(f, "copying") {
			licenseFound = true
			break
		}
	}
	if !licenseFound {
		errors = append(errors, "Spec should include license file in ExtraFiles. "+
			"See http://gonuts.io/-/doc/faq for reasoning.")
	}

	// check homepage
	if spec.Homepage != "" {
		u, err := url.Parse(spec.Homepage)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Can't parse homepage: %s", err))
		} else {
			if !u.IsAbs() || u.Opaque != "" || (u.Scheme != "http" && u.Scheme != "https") {
				errors = append(errors, "Homepage should be absolute http:// or https:// URL.")
			}
		}
	}

	return
}
