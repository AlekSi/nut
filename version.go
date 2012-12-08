package nut

import (
	"fmt"
	"regexp"
	"strconv"
)

// Current format for nut version.
var VersionRegexp = regexp.MustCompile(`^(\d+).(\d+).(\d+)$`)

// Describes nut version. See http://gonuts.io/-/doc/versioning for explanation of version specification.
type Version struct {
	Major int
	Minor int
	Patch int
}

// Parse and set version.
func NewVersion(version string) (v *Version, err error) {
	v = new(Version)
	err = v.setVersion(version)
	return
}

// Return version as string in current format.
func (v Version) String() string {
	res := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if !VersionRegexp.MatchString(res) { // sanity check
		panic(fmt.Sprintf("%s not matches %s", res, VersionRegexp))
	}
	return res
}

// Returns true if left < right, false otherwise.
func (left *Version) Less(right *Version) bool {
	if left.Major < right.Major {
		return true
	} else if left.Major > right.Major {
		return false
	}

	if left.Minor < right.Minor {
		return true
	} else if left.Minor > right.Minor {
		return false
	}

	if left.Patch < right.Patch {
		return true
	} else if left.Patch > right.Patch {
		return false
	}

	// left == right => "left < right" is false
	return false
}

// Marshal to JSON.
func (v *Version) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, v)), nil
}

// Unmarshal from JSON.
func (v *Version) UnmarshalJSON(b []byte) error {
	return v.setVersion(string(b[1 : len(b)-1]))
}

func (v *Version) setVersion(version string) (err error) {
	parsed := VersionRegexp.FindAllStringSubmatch(version, -1)
	if (parsed == nil) || (len(parsed[0]) != 4) {
		err = fmt.Errorf("Bad format for version %q: parsed as %#v", version, parsed)
		return
	}

	v.Major, _ = strconv.Atoi(parsed[0][1])
	v.Minor, _ = strconv.Atoi(parsed[0][2])
	v.Patch, _ = strconv.Atoi(parsed[0][3])
	return
}
