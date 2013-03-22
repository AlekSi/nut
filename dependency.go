package nut

import (
	"fmt"
	"regexp"
	"strconv"
)

// Current format.
var DependencyRegexp = regexp.MustCompile(`^(\d+|\*).(\d+|\*).(\d+|\*)$`)

// Describes dependency information.
type Dependency struct {
	Name    string // Nut name
	Version string // Nut version expression
}

func (d Dependency) String() string {
	return fmt.Sprintf("%s (%s)", d.Name, d.Version)
}

func (d *Dependency) Matches(nut *Nut) bool {
	// check name
	if d.Name != nut.Name {
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
