package nut

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func minSection(ints ...int) (min int) {
	min = MaxSectionValue
	for _, i := range ints {
		if min > i {
			min = i
		}
	}
	return
}

func maxSection(ints ...int) (max int) {
	max = MinSectionValue
	for _, i := range ints {
		if max < i {
			max = i
		}
	}
	return
}

type parsedDependency struct {
	majorMin, majorMax int
	minorMin, minorMax int
	patchMin, patchMax int
}

func (pd *parsedDependency) valid() bool {
	return pd.majorMin <= pd.majorMax &&
		pd.minorMin <= pd.minorMax &&
		pd.patchMin <= pd.patchMax
}

func (pd *parsedDependency) String() (v string) {
	if !pd.valid() {
		panic(fmt.Errorf("%#v is not valid", pd))
	}

	s := make([]string, 3)
	for i, mm := range [][2]int{{pd.majorMin, pd.majorMax}, {pd.minorMin, pd.minorMax}, {pd.patchMin, pd.patchMax}} {
		min, max := mm[0], mm[1]
		if min == MinSectionValue && max == MaxSectionValue {
			s[i] = "*"
		} else if min == max {
			s[i] = strconv.Itoa(min)
		} else if max == MaxSectionValue {
			s[i] = ">=" + strconv.Itoa(min)
		}
	}

	v = strings.Join(s, ".")
	if !NutDependencyRegexp.MatchString(v) { // sanity check
		panic(fmt.Errorf("%s not matches %s", v, NutDependencyRegexp))
	}
	return
}

var (
	// Current format for nut dependency.
	NutDependencyRegexp      = regexp.MustCompile(`^((?:>=)?\d+|\*)\.((?:>=)?\d+|\*)\.((?:>=)?\d+|\*)$`)
	NutFixedDependencyRegexp = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	VcsDependencyRegexp      = regexp.MustCompile(`^(bzr|git|hg|svn):(\S+)$`)
)

// Describes dependency information.
type Dependency struct {
	ImportPath string
	Version    string
	parsed     *parsedDependency
}

func NewDependency(importPath, version string) (d *Dependency, err error) {
	d = &Dependency{ImportPath: importPath, Version: version}
	if !d.OnNut() && !d.OnVcs() {
		err = fmt.Errorf("Bad format for dependency %q.", version)
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

		match := NutDependencyRegexp.FindAllStringSubmatch(d.Version, -1)
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
	return NutDependencyRegexp.MatchString(d.Version)
}

func (d *Dependency) OnVcs() bool {
	return VcsDependencyRegexp.MatchString(d.Version)
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

type AddDependencyError struct {
	ex, add *Dependency
}

func (e *AddDependencyError) Error() string {
	return fmt.Sprintf("Can't add %s to existing dependecy %s", e.add, e.ex)
}

type Dependencies struct {
	d map[string]Dependency // import path to dependency
}

// check interface
var (
	_ error            = &AddDependencyError{}
	_ json.Marshaler   = &Dependencies{}
	_ json.Unmarshaler = &Dependencies{}
)

func NewDependencies() (deps *Dependencies) {
	deps = new(Dependencies)
	deps.Clear()
	return
}

func (deps *Dependencies) Clear() {
	deps.d = make(map[string]Dependency)
}

func (deps *Dependencies) Add(d *Dependency) (err error) {
	e := deps.Get(d.ImportPath)
	if e == nil {
		deps.d[d.ImportPath] = *d
		return
	}

	e.parse()
	e.parsed.majorMin = maxSection(e.parsed.majorMin, d.MajorMin())
	e.parsed.majorMax = minSection(e.parsed.majorMax, d.MajorMax())
	e.parsed.minorMin = maxSection(e.parsed.minorMin, d.MinorMin())
	e.parsed.minorMax = minSection(e.parsed.minorMax, d.MinorMax())
	e.parsed.patchMin = maxSection(e.parsed.patchMin, d.PatchMin())
	e.parsed.patchMax = minSection(e.parsed.patchMax, d.PatchMax())

	if !e.parsed.valid() {
		err = &AddDependencyError{ex: deps.Get(d.ImportPath), add: d}
		return
	}

	e.Version = e.parsed.String()
	deps.d[d.ImportPath] = *e
	return
}

func (deps *Dependencies) Get(importPath string) (dep *Dependency) {
	d, ok := deps.d[importPath]
	if ok {
		dep = &d
	}
	return
}

func (deps *Dependencies) ImportPaths() (paths []string) {
	paths = make([]string, 0, len(deps.d))
	for i := range deps.d {
		paths = append(paths, i)
	}
	sort.Strings(paths)
	return
}

func (deps *Dependencies) MarshalJSON() (b []byte, err error) {
	b = make([]byte, 0, 50*len(deps.d))
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
		err = deps.Add(d)
		if err != nil {
			return
		}
	}
	return
}
