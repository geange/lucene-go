package core

import "github.com/geange/lucene-go/core/util"

// PostingsEnum Iterates through the postings. NOTE: you must first call nextDoc before using any of the
// per-doc methods.
type PostingsEnum interface {
	// Freq Returns term frequency in the current document, or 1 if the field was indexed with IndexOptions.DOCS. Do not call this before nextDoc is first called, nor after nextDoc returns DocIdSetIterator.NO_MORE_DOCS.
	// NOTE: if the PostingsEnum was obtain with NONE, the result of this method is undefined.
	Freq() (int, error)

	// NextPosition Returns the next position, or -1 if positions were not indexed. Calling this more than freq() times is undefined.
	NextPosition() (int, error)

	// StartOffset Returns start offset for the current position, or -1 if offsets were not indexed.
	StartOffset() (int, error)

	// EndOffset Returns end offset for the current position, or -1 if offsets were not indexed.
	EndOffset() (int, error)

	// GetPayload Returns the payload at this position, or null if no payload was indexed. You should not
	// modify anything (neither members of the returned BytesRef nor bytes in the byte[]).
	GetPayload() (*util.BytesRef, error)
}

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
