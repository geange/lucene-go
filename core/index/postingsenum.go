package index

const (
	POSTINGS_ENUM_NONE      = 0
	POSTINGS_ENUM_FREQS     = 1 << 3
	POSTINGS_ENUM_POSITIONS = POSTINGS_ENUM_FREQS | 1<<4
	POSTINGS_ENUM_OFFSETS   = POSTINGS_ENUM_POSITIONS | 1<<5
	POSTINGS_ENUM_PAYLOADS  = POSTINGS_ENUM_POSITIONS | 1<<6
	POSTINGS_ENUM_ALL       = POSTINGS_ENUM_OFFSETS | POSTINGS_ENUM_PAYLOADS
)

// FeatureRequested Returns true if the given feature is requested in the flags, false otherwise.
func FeatureRequested(flags, feature int) bool {
	return (flags & feature) == feature
}
