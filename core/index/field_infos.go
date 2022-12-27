package index

import (
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/geange/lucene-go/core/types"
)

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
	softDeletesField string

	// used only by fieldInfo(int)
	byNumber []types.FieldInfo

	byName map[string]*types.FieldInfo
	values []*types.FieldInfo // for an unmodifiable iterator
}

func NewFieldInfos(infos []*types.FieldInfo) *FieldInfos {
	hasVectors := false
	hasProx := false
	hasPayloads := false
	hasOffsets := false
	hasFreq := false
	hasNorms := false
	hasDocValues := false
	hasPointValues := false
	softDeletesField := ""

	tmap := treeset.NewWith(func(a, b interface{}) int {
		info1 := a.(*types.FieldInfo)
		info2 := b.(*types.FieldInfo)
		if info1.Number == info2.Number {
			return 0
		} else if info1.Number > info2.Number {
			return 1
		} else {
			return -1
		}
	})

	max := 0
	for _, info := range infos {
		if info.Number > max {
			max = info.Number
		}
	}

	this := &FieldInfos{
		byName: map[string]*types.FieldInfo{},
	}

	for _, info := range infos {
		if info.Number < 0 {
			panic("")
		}

		if tmap.Contains(info) {
			panic("")
		}

		tmap.Add(info)

		if _, ok := this.byName[info.Name]; ok {
			panic("")
		} else {
			this.byName[info.Name] = info
		}

		hasVectors = hasVectors || info.HasVectors()
		hasProx = hasProx || info.GetIndexOptions() >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
		hasFreq = hasFreq || info.GetIndexOptions() != types.INDEX_OPTIONS_DOCS
		hasOffsets = hasOffsets || info.GetIndexOptions() >= types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		hasNorms = hasNorms || info.HasNorms()
		hasDocValues = hasDocValues || info.GetDocValuesType() != types.DOC_VALUES_TYPE_NONE
		hasPayloads = hasPayloads || info.HasPayloads()
		hasPointValues = hasPointValues || info.GetPointDimensionCount() != 0

		if info.IsSoftDeletesField() {
			if softDeletesField == info.Name {
				panic("")
			}
			softDeletesField = info.Name
		}
	}

	this.hasVectors = hasVectors
	this.hasProx = hasProx
	this.hasPayloads = hasPayloads
	this.hasOffsets = hasOffsets
	this.hasFreq = hasFreq
	this.hasNorms = hasNorms
	this.hasDocValues = hasDocValues
	this.hasPointValues = hasPointValues
	this.softDeletesField = softDeletesField

	values := tmap.Values()
	items := make([]*types.FieldInfo, 0, len(values))
	for _, value := range values {
		info := value.(*types.FieldInfo)
		items = append(items, info)
	}
	this.values = items

	return this
}

func (f *FieldInfos) FieldInfo(fieldName string) *types.FieldInfo {
	return f.byName[fieldName]
}
