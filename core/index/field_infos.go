package index

import "github.com/geange/lucene-go/core/types"

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
	byNumber []types.FieldInfo

	byName map[string]*types.FieldInfo
	values []*types.FieldInfo // for an unmodifiable iterator
}

func NewFieldInfos(infos []types.FieldInfo) *FieldInfos {
	return nil
}

func (f *FieldInfos) FieldInfo(fieldName string) *types.FieldInfo {
	return f.byName[fieldName]
}
