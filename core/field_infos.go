package core

import "github.com/geange/lucene-go/core/index"

// FieldInfos Collection of FieldInfos (accessible by number or by name).
type FieldInfos struct {
	hasFreq          bool
	hasProx          bool
	hasPayloads      bool
	hasOffsets       bool
	hasVectors       bool
	hasNorms         bool
	hasDocValues     bool
	hasPointValues   bool
	softDeletesField bool

	// used only by fieldInfo(int)
	byNumber []index.FieldInfo

	byName map[string]*index.FieldInfo
	values []*index.FieldInfo // for an unmodifiable iterator
}
