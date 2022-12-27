package index

import (
	"errors"
	"fmt"
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
	Name string          // Unique segment name in the directory.
	Dir  store.Directory // Where this segment resides.

	maxDoc         int    // number of docs in seg
	isCompoundFile bool   //
	id             []byte // Id that uniquely identifies this segment.
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

func NewSegmentInfo(dir store.Directory, version, minVersion *util.Version, name string,
	maxDoc int, isCompoundFile bool, codec Codec, diagnostics map[string]string,
	id []byte, attributes map[string]string, indexSort *types.Sort) *SegmentInfo {

	return &SegmentInfo{
		Name:           name,
		Dir:            dir,
		maxDoc:         maxDoc,
		isCompoundFile: isCompoundFile,
		id:             id,
		codec:          codec,
		diagnostics:    diagnostics,
		attributes:     attributes,
		indexSort:      indexSort,
		version:        version,
		minVersion:     minVersion,
	}
}

// MaxDoc Returns number of documents in this segment (deletions are not taken into account).
func (s *SegmentInfo) MaxDoc() (int, error) {
	if s.maxDoc == -1 {
		return 0, errors.New("maxDoc isn't set yet")
	}
	return s.maxDoc, nil
}

func (s *SegmentInfo) SetMaxDoc(maxDoc int) error {
	if s.maxDoc != -1 {
		return fmt.Errorf("maxDoc was already set: this.maxDoc=%d vs maxDoc=%d", s.maxDoc, maxDoc)
	}
	s.maxDoc = maxDoc
	return nil
}
