package nut

import (
	"fmt"
	"regexp"
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

type parsedNutDependency struct {
	majorMin, majorMax int
	minorMin, minorMax int
	patchMin, patchMax int
}

// check interface
var (
	_ fmt.Stringer = &parsedNutDependency{}
)

func (p *parsedNutDependency) valid() bool {
	return p.majorMin <= p.majorMax &&
		p.minorMin <= p.minorMax &&
		p.patchMin <= p.patchMax
}

func (p *parsedNutDependency) String() (v string) {
	if !p.valid() {
		panic(fmt.Errorf("%#v is not valid", p))
	}

	s := make([]string, 3)
	for i, mm := range [][2]int{{p.majorMin, p.majorMax}, {p.minorMin, p.minorMax}, {p.patchMin, p.patchMax}} {
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
	NutDependencyRegexp = regexp.MustCompile(`^((?:>=)?\d+|\*)\.((?:>=)?\d+|\*)\.((?:>=)?\d+|\*)$`)

	// Current format for fixed nut dependency.
	NutFixedDependencyRegexp = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)

	// Current format for fixed VCS dependency.
	VcsFixedDependencyRegexp = regexp.MustCompile(`^(bzr|git|hg):(\S+)$`)
)

// Describes dependency information.
type Dependency struct {
	ImportPath string
	Version    string
	parsed     *parsedNutDependency
}

// check interface
var (
	_ fmt.Stringer = &Dependency{}
)

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
		d.parsed = &parsedNutDependency{}
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
	return VcsFixedDependencyRegexp.MatchString(d.Version)
}

func (d *Dependency) IsFixed() bool {
	return NutFixedDependencyRegexp.MatchString(d.Version) || VcsFixedDependencyRegexp.MatchString(d.Version)
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
