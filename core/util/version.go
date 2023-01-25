package util

import (
	"fmt"
	"strconv"
	"strings"
)

// Version Use by certain classes to match version compatibility across releases of Lucene.
// WARNING: When changing the version parameter that you supply to components in Lucene, do not simply change
// the version at search-time, but instead also adjust your indexing code to match, and re-index.
type Version struct {
	// Major version, the difference between stable and trunk
	Major int

	// Minor version, incremented within the stable branch
	Minor int

	// Bugfix number, incremented on release branches
	Bugfix int

	// Prerelease version, currently 0 (alpha), 1 (beta), or 2 (final)
	Prerelease int

	// stores the version pieces, with most significant pieces in high bits
	// ie:  | 1 byte | 1 byte | 1 byte |   2 bits   |
	//         major   minor    bugfix   prerelease
	encodedValue int
}

func NewVersion(major, minor, bugfix int) *Version {
	return &Version{
		Major:        major,
		Minor:        minor,
		Bugfix:       bugfix,
		Prerelease:   0,
		encodedValue: 0,
	}
}

func ParseVersion(version string) (*Version, error) {
	tokens := strings.Split(version, ".")

	switch len(tokens) {
	case 3:
		major, err := strconv.Atoi(tokens[0])
		if err != nil {
			return nil, err
		}
		minor, err := strconv.Atoi(tokens[1])
		if err != nil {
			return nil, err
		}
		bugfix, err := strconv.Atoi(tokens[2])
		if err != nil {
			return nil, err
		}
		return &Version{
			Major:        major,
			Minor:        minor,
			Bugfix:       bugfix,
			Prerelease:   0,
			encodedValue: 0,
		}, nil
	case 4:
		major, err := strconv.Atoi(tokens[0])
		if err != nil {
			return nil, err
		}
		minor, err := strconv.Atoi(tokens[1])
		if err != nil {
			return nil, err
		}
		bugfix, err := strconv.Atoi(tokens[2])
		if err != nil {
			return nil, err
		}
		prerelease, err := strconv.Atoi(tokens[3])
		if err != nil {
			return nil, err
		}
		return &Version{
			Major:        major,
			Minor:        minor,
			Bugfix:       bugfix,
			Prerelease:   prerelease,
			encodedValue: 0,
		}, nil
	}
	return nil, fmt.Errorf("parse '%s' error", version)
}

func (v *Version) String() string {
	if v.Prerelease == 0 {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Bugfix)
	}
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Bugfix, v.Prerelease)
}

var (
	VersionLast = NewVersion(8, 11, 1)
)
