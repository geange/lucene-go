package index

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/version"
)

const (
	// SegmentInfoNO
	// Used by some member fields to mean not present (e.g., norms, deletions).
	// e.g. no norms; no deletes;
	SegmentInfoNO = -1

	// SegmentInfoYES
	// Used by some member fields to mean present (e.g., norms, deletions).
	// e.g. have norms; have deletes;
	SegmentInfoYES = 1
)

var _ index.SegmentInfo = &SegmentInfo{}

// SegmentInfo
// Information about a segment such as its name, directory, and files related to the segment.
type SegmentInfo struct {
	name           string          // Unique segment name in the directory.
	dir            store.Directory // Where this segment resides.
	maxDoc         int             // number of docs in seg
	isCompoundFile bool            //
	id             []byte          // Id that uniquely identifies this segment.
	codec          index.Codec     //

	diagnosticsLock sync.RWMutex
	diagnostics     map[string]string //

	attributesLock sync.RWMutex
	attributes     map[string]string //

	indexSort index.Sort //

	// Tracks the Lucene version this segment was created with, since 3.1. Null
	// indicates an older than 3.0 index, and it's used to detect a too old index.
	// The format expected is "x.y" - "2.x" for pre-3.0 indexes (or null), and
	// specific versions afterwards ("3.0.0", "3.1.0" etc.).
	// see o.a.l.util.Version.
	version *version.Version

	// Tracks the minimum version that contributed documents to a segment. For
	// Flush segments, that is the version that wrote it. For merged segments,
	// this is the minimum minVersion of all the segments that have been merged
	// into this segment
	minVersion *version.Version

	filesLock sync.RWMutex
	setFiles  map[string]struct{}
}

func NewSegmentInfo(dir store.Directory, version, minVersion *version.Version, name string,
	maxDoc int, isCompoundFile bool, codec index.Codec, diagnostics map[string]string,
	id []byte, attributes map[string]string, indexSort index.Sort) *SegmentInfo {

	return &SegmentInfo{
		name:           name,
		dir:            dir,
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

func (s *SegmentInfo) GetID() []byte {
	return slices.Clone(s.id)
}

func (s *SegmentInfo) Name() string {
	return s.name
}

func (s *SegmentInfo) Dir() store.Directory {
	return s.dir
}

// Files
// Return all files referenced by this SegmentInfo.
func (s *SegmentInfo) Files() map[string]struct{} {
	s.filesLock.RLock()
	defer s.filesLock.RUnlock()

	return maps.Clone(s.setFiles)
}

// FilesNum
// 获取文件数量
func (s *SegmentInfo) FilesNum() int {
	return len(s.setFiles)
}

// MaxDoc
// Returns number of documents in this segment (deletions are not taken into account).
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

func (s *SegmentInfo) SetFiles(files map[string]struct{}) {
	s.filesLock.Lock()
	defer s.filesLock.Unlock()

	clear(s.setFiles)

	for fName := range files {
		s.setFiles[fName] = struct{}{}
	}
}

// AddFile
// Add this file to the set of files written for this segment.
func (s *SegmentInfo) AddFile(file string) error {
	if err := checkFileNames([]string{file}); err != nil {
		return err
	}

	fName := s.NamedForThisSegment(file)

	s.filesLock.Lock()
	defer s.filesLock.Unlock()

	s.setFiles[fName] = struct{}{}
	return nil
}

func (s *SegmentInfo) GetVersion() *version.Version {
	return s.version
}

func (s *SegmentInfo) GetMinVersion() *version.Version {
	return s.minVersion
}

// SetUseCompoundFile
// Mark whether this segment is stored as a compound file.
// Params: isCompoundFile – true if this is a compound file; else, false
func (s *SegmentInfo) SetUseCompoundFile(isCompoundFile bool) {
	s.isCompoundFile = isCompoundFile
}

// GetUseCompoundFile
// Returns true if this segment is stored as a compound file; else, false.
func (s *SegmentInfo) GetUseCompoundFile() bool {
	return s.isCompoundFile
}

func (s *SegmentInfo) SetDiagnostics(diagnostics map[string]string) {
	s.diagnostics = diagnostics
}

// GetDiagnostics
// Returns diagnostics saved into the segment when it was written. The map is immutable.
func (s *SegmentInfo) GetDiagnostics() map[string]string {
	s.diagnosticsLock.RLock()
	defer s.diagnosticsLock.RUnlock()

	return maps.Clone(s.diagnostics)
}

// PutAttribute
// Puts a codec attribute item.
// This is a key-item mapping for the field that the codec can use to store additional metadata,
// and will be available to the codec when reading the segment via getAttribute(String)
// If a item already exists for the field, it will be replaced with the new item. This method
// make a copy on write for every attribute change.
func (s *SegmentInfo) PutAttribute(key, value string) string {
	s.attributesLock.Lock()
	defer s.attributesLock.Unlock()

	old := s.attributes[key]
	s.attributes[key] = value
	return old
}

// GetAttributes
// Returns the internal codec attributes map.
// Returns: internal codec attributes map.
func (s *SegmentInfo) GetAttributes() map[string]string {
	s.attributesLock.RLock()
	defer s.attributesLock.RUnlock()

	return maps.Clone(s.attributes)
}

func (s *SegmentInfo) GetIndexSort() index.Sort {
	return s.indexSort
}

func (s *SegmentInfo) GetCodec() index.Codec {
	return s.codec
}

func checkFileNames(files []string) error {
	for _, file := range files {
		if !CODEC_FILE_PATTERN.MatchString(file) {
			return fmt.Errorf(`invalid codec filename: '%s', must match: %s`,
				file, CODEC_FILE_PATTERN.String())
		}

		if strings.HasSuffix(strings.ToLower(file), ".tmp") {
			return fmt.Errorf(`invalid codec filename: '%s', cannot end with .tmp extension`, file)
		}
	}
	return nil
}

// locates the boundary of the segment name, or -1
func indexOfSegmentName(filename string) int {
	// If it is a .del file, there's an '_' after the first character
	idx := strings.Index(filename[1:], "_")
	if idx == -1 {
		// If it's not, strip everything that's before the '.'
		idx = strings.Index(filename, ".")
	}
	return idx
}

// StripSegmentName
// Strips the segment name out of the given file name. If you used segmentFileName or
// fileNameFromGeneration to create your files, then this method simply removes whatever
// comes before the first '.', or the second '_' (excluding both).
// Returns: the filename with the segment name removed, or the given filename if it
// does not contain a '.' and '_'.
func StripSegmentName(filename string) string {
	idx := indexOfSegmentName(filename)
	if idx != -1 {
		filename = filename[idx:]
	}
	return filename
}

// NamedForThisSegment
// strips any segment name from the file, naming it with this segment this is because
// "segment names" can change,
// e.g. by addIndexes(Dir)
func (s *SegmentInfo) NamedForThisSegment(file string) string {
	return s.name + StripSegmentName(file)
}

func (s *SegmentInfo) SetCodec(codec index.Codec) {
	s.codec = codec
}
