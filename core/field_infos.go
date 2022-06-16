package core

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
	byNumber []FieldInfo

	byName map[string]*FieldInfo
	values []*FieldInfo // for an unmodifiable iterator
}
