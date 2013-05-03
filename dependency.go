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
var DependencyRegexp = regexp.MustCompile(`^((?:>=)?\d+|\*).((?:>=)?\d+|\*).((?:>=)?\d+|\*)$`)

// Describes dependency information.
type Dependency struct {
	ImportPath string
	Version    string
	parsed     *parsedDependency
}

type parsedDependency struct {
	majorMin, majorMax int
	minorMin, minorMax int
	patchMin, patchMax int
}

func NewDependency(importPath, version string) (d *Dependency, err error) {
	d = &Dependency{ImportPath: importPath, Version: version}
	if d.OnNut() && !DependencyRegexp.MatchString(version) {
		err = fmt.Errorf("Bad format for nut dependency %q.", version)
	}
	return
}

func (d *Dependency) parseSection(s string) (min, max int) {
	if s == "*" {
		return MinSectionValue, MaxSectionValue
	}

	var moreEqual bool
	if strings.HasPrefix(s, ">=") {
		moreEqual = true
		s = s[2:]
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("%#v: failed to parse section %q: %s", d, s, err))
	}
	min = n
	max = n

	if moreEqual {
		max = MaxSectionValue
	}
	return
}

func (d *Dependency) parse() {
	if d.parsed == nil {
		if !d.OnNut() {
			panic(fmt.Errorf("Not a nut: %#v", d))
		}

		match := DependencyRegexp.FindAllStringSubmatch(d.Version, -1)
		d.parsed = &parsedDependency{}
		d.parsed.majorMin, d.parsed.majorMax = d.parseSection(match[0][1])
		d.parsed.minorMin, d.parsed.minorMax = d.parseSection(match[0][2])
		d.parsed.patchMin, d.parsed.patchMax = d.parseSection(match[0][3])
	}
}

func (d Dependency) String() string {
	return fmt.Sprintf("%s (%s)", d.ImportPath, d.Version)
}

func (d *Dependency) OnNut() bool {
	return !strings.Contains(d.Version, ":")
}

func (d *Dependency) MajorMin() int {
	d.parse()
	return d.parsed.majorMin
}

func (d *Dependency) MajorMax() int {
	d.parse()
	return d.parsed.majorMax
}

func (d *Dependency) MinorMin() int {
	d.parse()
	return d.parsed.minorMin
}

func (d *Dependency) MinorMax() int {
	d.parse()
	return d.parsed.minorMax
}

func (d *Dependency) PatchMin() int {
	d.parse()
	return d.parsed.patchMin
}

func (d *Dependency) PatchMax() int {
	d.parse()
	return d.parsed.patchMax
}

func (d *Dependency) Matches(prefix string, nut *Nut) bool {
	if !d.OnNut() {
		panic(fmt.Errorf("Not a nut: (%#v).Matches(%#v)", d, nut))
	}

	// parse early to check for panic
	d.parse()

	// check import path
	if d.ImportPath != nut.ImportPath(prefix) {
		return false
	}

	// check version
	if d.MajorMin() > nut.Version.Major || d.MajorMax() < nut.Version.Major {
		return false
	}
	if d.MinorMin() > nut.Version.Minor || d.MinorMax() < nut.Version.Minor {
		return false
	}
	if d.PatchMin() > nut.Version.Patch || d.PatchMax() < nut.Version.Patch {
		return false
	}

	return true
}

type Dependencies struct {
	d map[string]string // import path to version
}

// check interface
var (
	_ json.Marshaler   = &Dependencies{}
	_ json.Unmarshaler = &Dependencies{}
)

func NewDependencies() *Dependencies {
	return &Dependencies{d: make(map[string]string)}
}

func (deps *Dependencies) Clear() {
	deps.d = make(map[string]string)
}

// TODO not replace, make narrow
func (deps *Dependencies) Add(d *Dependency) {
	deps.d[d.ImportPath] = d.Version
}

func (deps *Dependencies) Get(importPath string) (dep *Dependency) {
	v, ok := deps.d[importPath]
	if ok {
		dep = &Dependency{ImportPath: importPath, Version: v}
	}
	return
}

func (deps *Dependencies) Del(d *Dependency) {
	delete(deps.d, d.ImportPath)
}

func (deps *Dependencies) Len() int {
	return len(deps.d)
}

func (deps *Dependencies) ImportPaths() (paths []string) {
	paths = make([]string, 0, deps.Len())
	for i := range deps.d {
		paths = append(paths, i)
	}
	sort.Strings(paths)
	return
}

func (deps *Dependencies) MarshalJSON() (b []byte, err error) {
	b = make([]byte, 0, 50*deps.Len())
	b = append(b, '{')
	for _, p := range deps.ImportPaths() {
		d := deps.Get(p)
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
	deps.Clear() // also make
	var d *Dependency
	for p, v := range m {
		d, err = NewDependency(p, v)
		if err != nil {
			return
		}
		deps.Add(d)
	}
	return
}
