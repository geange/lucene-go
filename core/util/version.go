package util

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

var (
	VersionLast = NewVersion(8, 11, 1)
)
