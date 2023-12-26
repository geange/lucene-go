package index

import (
	"errors"
	"github.com/geange/gods-generic/sets/treeset"
	"github.com/geange/lucene-go/core/document"
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
	byNumber []*document.FieldInfo

	byName map[string]*document.FieldInfo
	values []*document.FieldInfo // for an unmodifiable iterator
}

func NewFieldInfos(infos []*document.FieldInfo) *FieldInfos {
	hasVectors := false
	hasProx := false
	hasPayloads := false
	hasOffsets := false
	hasFreq := false
	hasNorms := false
	hasDocValues := false
	hasPointValues := false
	softDeletesField := ""

	tmap := treeset.NewWith[*document.FieldInfo](func(info1, info2 *document.FieldInfo) int {
		if info1.Number() == info2.Number() {
			return 0
		} else if info1.Number() > info2.Number() {
			return 1
		} else {
			return -1
		}
	})

	maxNum := 0
	for _, info := range infos {
		if info.Number() > maxNum {
			maxNum = info.Number()
		}
	}

	this := &FieldInfos{
		byName:   map[string]*document.FieldInfo{},
		byNumber: []*document.FieldInfo{},
	}

	for _, info := range infos {
		if info.Number() < 0 {
			panic("")
		}

		if tmap.Contains(info) {
			panic("")
		}

		tmap.Add(info)

		if _, ok := this.byName[info.Name()]; ok {
			panic("")
		} else {
			this.byName[info.Name()] = info
		}

		hasVectors = hasVectors || info.HasVectors()
		hasProx = hasProx || info.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
		hasFreq = hasFreq || info.GetIndexOptions() != document.INDEX_OPTIONS_DOCS
		hasOffsets = hasOffsets || info.GetIndexOptions() >= document.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
		hasNorms = hasNorms || info.HasNorms()
		hasDocValues = hasDocValues || info.GetDocValuesType() != document.DOC_VALUES_TYPE_NONE
		hasPayloads = hasPayloads || info.HasPayloads()
		hasPointValues = hasPointValues || info.GetPointDimensionCount() != 0

		if info.IsSoftDeletesField() {
			if softDeletesField == info.Name() {
				panic("")
			}
			softDeletesField = info.Name()
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
	items := make([]*document.FieldInfo, 0, len(values))
	for _, value := range values {
		items = append(items, value)
	}
	this.byNumber = items
	this.values = items

	return this
}

func (f *FieldInfos) FieldInfo(fieldName string) *document.FieldInfo {
	return f.byName[fieldName]
}

func (f *FieldInfos) FieldInfoByNumber(fieldNumber int) *document.FieldInfo {
	return f.byNumber[fieldNumber]
}

func (f *FieldInfos) Size() int {
	return len(f.byName)
}

func (f *FieldInfos) List() []*document.FieldInfo {
	return f.values
}

func (f *FieldInfos) HasNorms() bool {
	return f.hasNorms
}

func (f *FieldInfos) HasDocValues() bool {
	return f.hasDocValues
}

func (f *FieldInfos) HasVectors() bool {
	return f.hasVectors
}

func (f *FieldInfos) HasPointValues() bool {
	return f.hasPointValues
}

type FieldInfosBuilder struct {
	byName             map[string]*document.FieldInfo
	globalFieldNumbers *FieldNumbers
	finished           bool
}

func NewFieldInfosBuilder(globalFieldNumbers *FieldNumbers) *FieldInfosBuilder {
	return &FieldInfosBuilder{
		byName:             make(map[string]*document.FieldInfo),
		globalFieldNumbers: globalFieldNumbers,
		finished:           false,
	}
}

func (f *FieldInfosBuilder) Add(other *FieldInfos) error {
	if f.assertNotFinished() != nil {
		return nil
	}

	for _, fieldInfo := range other.values {
		if _, err := f.AddFieldInfo(fieldInfo); err != nil {
			return err
		}
	}
	return nil
}

// GetOrAdd Create a new field, or return existing one.
func (f *FieldInfosBuilder) GetOrAdd(name string) (*document.FieldInfo, error) {
	fi, ok := f.byName[name]
	if !ok {
		if err := f.assertNotFinished(); err != nil {
			return nil, err
		}
		// This field wasn't yet added to this in-RAM
		// segment's FieldInfo, so now we get a global
		// number for this field.  If the field was seen
		// before then we'll get the same name and number,
		// else we'll allocate a new one:
		isSoftDeletesField := name == f.globalFieldNumbers.softDeletesFieldName
		fieldNumber, err := f.globalFieldNumbers.AddOrGet(name, -1,
			document.INDEX_OPTIONS_NONE, document.DOC_VALUES_TYPE_NONE,
			0, 0, 0, isSoftDeletesField)
		if err != nil {
			return nil, err
		}
		fi = document.NewFieldInfo(name, fieldNumber, false, false, false,
			document.INDEX_OPTIONS_NONE, document.DOC_VALUES_TYPE_NONE,
			-1, map[string]string{},
			0, 0, 0, isSoftDeletesField)
		//assert !byName.containsKey(fi.name);
		if err := f.globalFieldNumbers.verifyConsistentDocValuesType(
			fi.Number(), fi.Name(), document.DOC_VALUES_TYPE_NONE); err != nil {
			return nil, err
		}
		f.byName[fi.Name()] = fi
	}

	return fi, nil
}

func (f *FieldInfosBuilder) AddFieldInfo(fi *document.FieldInfo) (*document.FieldInfo, error) {
	return f.AddFieldInfoV(fi, -1)
}

func (f *FieldInfosBuilder) AddFieldInfoV(fi *document.FieldInfo, dvGen int64) (*document.FieldInfo, error) {
	// IMPORTANT - reuse the field number if possible for consistent field numbers across segments
	return f.addOrUpdateInternal(fi.Name(), fi.Number(), fi.HasVectors(),
		fi.OmitsNorms(), fi.HasPayloads(),
		fi.GetIndexOptions(), fi.GetDocValuesType(), dvGen,
		fi.Attributes(),
		fi.GetPointDimensionCount(), fi.GetPointIndexDimensionCount(), fi.GetPointNumBytes(),
		fi.IsSoftDeletesField())
}

func (f *FieldInfosBuilder) addOrUpdateInternal(name string, preferredFieldNumber int,
	storeTermVector, omitNorms, storePayloads bool,
	indexOptions document.IndexOptions, docValues document.DocValuesType,
	dvGen int64, attributes map[string]string,
	dataDimensionCount, indexDimensionCount, dimensionNumBytes int,
	isSoftDeletesField bool) (*document.FieldInfo, error) {

	if err := f.assertNotFinished(); err != nil {
		return nil, err
	}

	fi, ok := f.byName[name]
	if !ok {
		// This field wasn't yet added to this in-RAM
		// segment's FieldInfo, so now we get a global
		// number for this field.  If the field was seen
		// before then we'll get the same name and number,
		// else we'll allocate a new one:
		fieldNumber, err := f.globalFieldNumbers.AddOrGet(
			name, preferredFieldNumber, indexOptions, docValues, dataDimensionCount,
			indexDimensionCount, dimensionNumBytes, isSoftDeletesField)
		if err != nil {
			return nil, err
		}

		fi = document.NewFieldInfo(name, fieldNumber, storeTermVector, omitNorms,
			storePayloads, indexOptions, docValues, dvGen, attributes, dataDimensionCount,
			indexDimensionCount, dimensionNumBytes, isSoftDeletesField)
		//assert !byName.containsKey(fi.name);
		if err := f.globalFieldNumbers.verifyConsistentDocValuesType(fi.Number(), fi.Name(), fi.GetDocValuesType()); err != nil {
			return nil, err
		}
		f.byName[fi.Name()] = fi
	} else {
		if err := fi.Update(storeTermVector, omitNorms, storePayloads,
			indexOptions, attributes, dataDimensionCount,
			indexDimensionCount, dimensionNumBytes); err != nil {
			return nil, err
		}

		if docValues != document.DOC_VALUES_TYPE_NONE {
			// Only pay the synchronization cost if fi does not already have a DVType
			updateGlobal := fi.GetDocValuesType() == document.DOC_VALUES_TYPE_NONE
			if updateGlobal {
				// Must also update docValuesType map so it's
				// aware of this field's DocValuesType.  This will throw IllegalArgumentException if
				// an illegal type change was attempted.
				if err := f.globalFieldNumbers.setDocValuesType(fi.Number(), name, docValues); err != nil {
					return nil, err
				}
			}

			if err := fi.SetDocValuesType(docValues); err != nil { // this will also perform the consistency check.
				return nil, err
			}
			if err := fi.SetDocValuesGen(dvGen); err != nil {
				return nil, err
			}
		}
	}
	return fi, nil
}

func (f *FieldInfosBuilder) fieldInfo(fieldName string) *document.FieldInfo {
	return f.byName[fieldName]
}

func (f *FieldInfosBuilder) assertNotFinished() error {
	if f.finished {
		return errors.New("builder is finished")
	}
	return nil
}

func (f *FieldInfosBuilder) Finish() *FieldInfos {
	f.finished = true

	list := make([]*document.FieldInfo, 0)
	for _, v := range f.byName {
		list = append(list, v)
	}
	return NewFieldInfos(list)
}
