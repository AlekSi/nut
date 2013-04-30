package nut

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Current format.
var DependencyRegexp = regexp.MustCompile(`^(\d+|\*).(\d+|\*).(\d+|\*)$`)

// Describes dependency information.
type Dependency struct {
	ImportPath string
	Version    string
}

func NewDependency(importPath, version string) (d *Dependency, err error) {
	d = &Dependency{importPath, version}
	if d.OnNut() && !DependencyRegexp.MatchString(version) {
		err = fmt.Errorf("Bad format for nut dependency %q.", version)
	}
	return
}

func (d Dependency) String() string {
	return fmt.Sprintf("%s (%s)", d.ImportPath, d.Version)
}

func (d *Dependency) OnNut() bool {
	return !strings.Contains(d.Version, ":")
}

func (d *Dependency) Matches(prefix string, nut *Nut) bool {
	if !d.OnNut() {
		panic(fmt.Errorf("Not a nut: (%#v).Matches(%#v)", d, nut))
	}

	// check import path
	if d.ImportPath != nut.ImportPath(prefix) {
		return false
	}

	// check exact matching and wildcards
	match := DependencyRegexp.FindAllStringSubmatch(d.Version, -1)
	if match != nil {
		major_s := match[0][1]
		minor_s := match[0][2]
		patch_s := match[0][3]

		major, _ := strconv.Atoi(major_s)
		minor, _ := strconv.Atoi(minor_s)
		patch, _ := strconv.Atoi(patch_s)

		if (major_s == "*" || major == nut.Version.Major) &&
			(minor_s == "*" || minor == nut.Version.Minor) &&
			(patch_s == "*" || patch == nut.Version.Patch) {
			return true
		}
	}

	return false
}

type Dependencies []Dependency

// check interface
var (
	_ sort.Interface   = &Dependencies{}
	_ json.Marshaler   = &Dependencies{}
	_ json.Unmarshaler = &Dependencies{}
)

func (deps Dependencies) Len() int           { return len(deps) }
func (deps Dependencies) Less(i, j int) bool { return deps[i].ImportPath < deps[j].ImportPath }
func (deps Dependencies) Swap(i, j int)      { deps[i], deps[j] = deps[j], deps[i] }

func (deps *Dependencies) MarshalJSON() (b []byte, err error) {
	sort.Sort(deps)
	b = make([]byte, 0, 50*len(*deps))
	b = append(b, '{')
	for _, d := range *deps {
		b = append(b, '"')
		b = append(b, d.ImportPath...)
		b = append(b, `":"`...)
		b = append(b, d.Version...)
		b = append(b, `",`...)
	}
	b[len(b)-1] = '}'
	return
}

func (deps *Dependencies) UnmarshalJSON(b []byte) (err error) {
	m := make(map[string]string)
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	for p, v := range m {
		*deps = append(*deps, Dependency{p, v})
	}
	sort.Sort(*deps)
	return
}
