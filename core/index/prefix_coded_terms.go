package index

import "bytes"

// Prefix codes term instances (prefixes are shared).
// This is expected to be faster to build than a FST and might also be more compact
// if there are no common suffixes.
// lucene.internal
type PrefixCodedTerms struct {
	buffer *bytes.Buffer
}
