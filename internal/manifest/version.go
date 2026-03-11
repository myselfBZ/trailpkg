package manifest

import (
	"strconv"
	"strings"
)

type Version struct {
    Major, Minor, Patch int
}

func ParseVersion(s string) (Version, error) {
	parts := strings.Split(s, ".")
	var v Version
	var err error

	if len(parts) > 0 {
		v.Major, err = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 && err == nil {
		v.Minor, err = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 && err == nil {
		v.Patch, err = strconv.Atoi(parts[2])
	}

	return v, err
}

func (v Version) IsAtLeast(other Version) bool {
    if v.Major != other.Major {
        return v.Major > other.Major
    }
    if v.Minor != other.Minor {
        return v.Minor > other.Minor
    }
    return v.Patch >= other.Patch
}
