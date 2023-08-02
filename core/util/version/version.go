package version

import (
	"fmt"
	"strings"
)

// Version
// Use by certain classes to match version compatibility across releases of Lucene.
// WARNING: When changing the version parameter that you supply to components in Lucene, do not simply change
// the version at search-time, but instead also adjust your indexing code to match, and re-index.
type Version struct {
	// major version, the difference between stable and trunk
	major uint8

	// minor version, incremented within the stable branch
	minor uint8

	// bugfix number, incremented on release branches
	bugfix uint8

	// Prerelease version, currently 0 (alpha), 1 (beta), or 2 (final)
	prerelease uint8

	// stores the version pieces, with most significant pieces in high bits
	// ie:  | 1 byte | 1 byte | 1 byte |   2 bits   |
	//         major   minor    bugfix   prerelease
	encodedValue uint32
}

func New(options ...Option) (*Version, error) {
	op := &option{}
	for _, fn := range options {
		fn(op)
	}

	if op.prerelease > 2 || op.prerelease < 0 {
		return nil, fmt.Errorf("illegal prerelease version: %d", op.prerelease)
	}

	if op.prerelease != 0 && (op.minor != 0 || op.bugfix != 0) {
		format := "prerelease version only supported with major release (got prerelease: %d, minor: %d, bugfix: %d)"
		return nil, fmt.Errorf(format, op.prerelease, op.minor, op.bugfix)
	}

	return newVersion(op.major, op.minor, op.bugfix, op.prerelease), nil
}

func newVersion(major, minor, bugfix, prerelease uint8) *Version {
	return &Version{
		major:        major,
		minor:        minor,
		bugfix:       bugfix,
		prerelease:   prerelease,
		encodedValue: uint32(major)<<18 | uint32(minor)<<10 | uint32(bugfix)<<2 | uint32(prerelease),
	}
}

func Parse(version string) (*Version, error) {
	tokens := strings.Split(version, ".")

	switch len(tokens) {
	case 3:
		var major, minor, bugfix uint8
		if _, err := fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &bugfix); err != nil {
			return nil, err
		}

		return New(
			WithMajor(major),
			WithMinor(minor),
			WithBugfix(bugfix),
		)
	case 4:
		var major, minor, bugfix, prerelease uint8
		if _, err := fmt.Sscanf(version, "%d.%d.%d.%d", &major, &minor, &bugfix, &prerelease); err != nil {
			return nil, err
		}
		return New(
			WithMajor(major),
			WithMinor(minor),
			WithBugfix(bugfix),
			WithPrerelease(prerelease),
		)
	}
	return nil, fmt.Errorf("parse '%s' error", version)
}

func (v *Version) Major() uint8 {
	return v.major
}

func (v *Version) Minor() uint8 {
	return v.minor
}

func (v *Version) Bugfix() uint8 {
	return v.bugfix
}

func (v *Version) Prerelease() uint8 {
	return v.prerelease
}

func (v *Version) String() string {
	if v.prerelease == 0 {
		return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.bugfix)
	}
	return fmt.Sprintf("%d.%d.%d.%d", v.major, v.minor, v.bugfix, v.prerelease)
}

func (v *Version) OnOrAfter(other *Version) bool {
	return v.encodedValue >= other.encodedValue
}

func (v *Version) Clone() *Version {
	if v == nil {
		return nil
	}

	return &Version{
		major:        v.major,
		minor:        v.minor,
		bugfix:       v.bugfix,
		prerelease:   v.prerelease,
		encodedValue: v.encodedValue,
	}
}

var (
	Last = newVersion(8, 11, 1, 0)
)
