package utils

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (version *Version) Parse(raw string) error {
	parts := strings.Split(raw, ".")
	if len(parts) == 3 {
		version.Major, _ = strconv.Atoi(parts[0])
		version.Minor, _ = strconv.Atoi(parts[1])
		version.Patch, _ = strconv.Atoi(parts[2])
		return nil
	}
	return fmt.Errorf("could not parse %s into version", raw)
}

func (version *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch)
}

func ParseVersion(raw string) (v *Version, e error) {
	v = &Version{}
	e = v.Parse(raw)
	return v, e
}

func (version *Version) Smaller(other *Version) bool {
	if version.Major < other.Major {
		return true
	}
	if version.Minor < other.Minor {
		return true
	}
	if version.Patch < other.Patch {
		return true
	}
	return false
}
