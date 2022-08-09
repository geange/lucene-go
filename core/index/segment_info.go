package index

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

const (
	// SegmentInfoNO Used by some member fields to mean not present (e.g., norms, deletions).
	// e.g. no norms; no deletes;
	SegmentInfoNO = -1

	// SegmentInfoYES Used by some member fields to mean present (e.g., norms, deletions).
	// e.g. have norms; have deletes;
	SegmentInfoYES = 1
)

// SegmentInfo Information about a segment such as its name, directory, and files related to the segment.
type SegmentInfo struct {
	name           string          // Unique segment name in the directory.
	maxDoc         int             // number of docs in seg
	dir            store.Directory // Where this segment resides.
	isCompoundFile bool            //
	id             []byte          // Id that uniquely identifies this segment.
	codec          Codec

	diagnostics map[string]string
	attributes  map[string]string
	indexSort   *types.Sort

	// Tracks the Lucene version this segment was created with, since 3.1. Null
	// indicates an older than 3.0 index, and it's used to detect a too old index.
	// The format expected is "x.y" - "2.x" for pre-3.0 indexes (or null), and
	// specific versions afterwards ("3.0.0", "3.1.0" etc.).
	// see o.a.l.util.Version.
	version *util.Version

	// Tracks the minimum version that contributed documents to a segment. For
	// flush segments, that is the version that wrote it. For merged segments,
	// this is the minimum minVersion of all the segments that have been merged
	// into this segment
	minVersion *util.Version
}
