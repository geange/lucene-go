package index

import (
	"errors"
	"fmt"
	"github.com/geange/gods-generic/sets/treeset"
	"github.com/geange/lucene-go/core/document"
	"sync"
)

// FieldInfos
// Collection of FieldInfos (accessible by number or by name).
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

type FieldNumbers struct {
	sync.Mutex

	numberToName map[int]string
	nameToNumber map[string]int
	indexOptions map[string]document.IndexOptions
	// We use this to enforce that a given field never
	// changes DV type, even across segments / IndexWriter
	// sessions:
	docValuesType map[string]document.DocValuesType

	dimensions map[string]*FieldDimensions

	// TODO: we should similarly catch an attempt to turn
	// norms back on after they were already committed; today
	// we silently discard the norm but this is badly trappy
	lowestUnassignedFieldNumber int

	softDeletesFieldName string
}

func NewFieldNumbers(softDeletesFieldName string) *FieldNumbers {
	return &FieldNumbers{
		numberToName:                map[int]string{},
		nameToNumber:                map[string]int{},
		indexOptions:                map[string]document.IndexOptions{},
		docValuesType:               map[string]document.DocValuesType{},
		dimensions:                  map[string]*FieldDimensions{},
		lowestUnassignedFieldNumber: -1,
		softDeletesFieldName:        softDeletesFieldName,
	}
}

// AddOrGet Returns the global field number for the given field name. If the name does not exist
// yet it tries to add it with the given preferred field number assigned if possible otherwise the
// first unassigned field number is used as the field number.
func (f *FieldNumbers) AddOrGet(fieldName string, preferredFieldNumber int,
	indexOptions document.IndexOptions, dvType document.DocValuesType,
	dimensionCount, indexDimensionCount, dimensionNumBytes int, isSoftDeletesField bool) (int, error) {

	f.Lock()
	defer f.Unlock()

	if indexOptions != document.INDEX_OPTIONS_NONE {
		currentOpts, ok := f.indexOptions[fieldName]
		if !ok {
			f.indexOptions[fieldName] = indexOptions
		} else if currentOpts != document.INDEX_OPTIONS_NONE && currentOpts != indexOptions {
			err := fmt.Errorf(`cannot change field %s from index options=%s to inconsistent index options=%s`, fieldName, currentOpts, indexOptions)
			return 0, err
		}
	}

	if dvType != document.DOC_VALUES_TYPE_NONE {
		currentDVType, ok := f.docValuesType[fieldName]
		if ok {
			f.docValuesType[fieldName] = dvType
		} else if currentDVType != document.DOC_VALUES_TYPE_NONE && currentDVType != dvType {
			return 0, fmt.Errorf(
				`cannot change DocValues type from %s to %s for field "%s"`,
				currentDVType, dvType, fieldName,
			)
		}
	}

	if dimensionCount != 0 {
		dims, ok := f.dimensions[fieldName]
		if ok {
			if dims.DimensionCount != dimensionCount {
				return 0, fmt.Errorf(
					`cannot change point dimension count from %d to %d for field="%s"`,
					dims.DimensionCount, dimensionCount, fieldName,
				)
			}

			if dims.IndexDimensionCount != indexDimensionCount {
				return 0, fmt.Errorf(
					`cannot change point index dimension count from %d to %d for field="%s"`,
					dims.IndexDimensionCount, indexDimensionCount, fieldName,
				)
			}

			if dims.DimensionNumBytes != dimensionNumBytes {
				return 0, fmt.Errorf(
					`cannot change point numBytes from %d to %d for field="%s"`,
					dims.DimensionNumBytes, dimensionNumBytes, fieldName,
				)
			}
		} else {
			f.dimensions[fieldName] = NewFieldDimensions(dimensionCount, indexDimensionCount, dimensionNumBytes)
		}
	}

	fieldNumber, ok := f.nameToNumber[fieldName]
	if !ok {
		preferredBoxed := preferredFieldNumber
		if _, ok := f.numberToName[preferredBoxed]; preferredFieldNumber != -1 && !ok {
			// cool - we can use this number globally
			fieldNumber = preferredBoxed
		} else {
			// find a new FieldNumber
			for {
				f.lowestUnassignedFieldNumber++
				if _, ok := f.numberToName[f.lowestUnassignedFieldNumber+1]; !ok {
					break
				}
			}

			fieldNumber = f.lowestUnassignedFieldNumber
		}
		//assert fieldNumber >= 0;
		f.numberToName[fieldNumber] = fieldName
		f.nameToNumber[fieldName] = fieldNumber
	}

	if isSoftDeletesField {
		if f.softDeletesFieldName == "" {
			return 0, fmt.Errorf(
				`this index has ["%s"] as soft-deletes already but soft-deletes field is not configured in IWC`,
				fieldName,
			)
		} else if fieldName != f.softDeletesFieldName {
			return 0, fmt.Errorf(
				`cannot configure ["%s"] as soft-deletes; this index uses ["%s"] as soft-deletes already`,
				f.softDeletesFieldName, fieldName,
			)
		}
	} else if fieldName == f.softDeletesFieldName {
		return 0, fmt.Errorf(
			`cannot configure ["%s"] as soft-deletes; this index uses ["%s"] as non-soft-deletes already`,
			f.softDeletesFieldName, fieldName,
		)
	}
	return fieldNumber, nil
}

func (f *FieldNumbers) verifyConsistentIndexOptions(number int, name string, indexOptions document.IndexOptions) error {
	if f.numberToName[number] != name {
		return fmt.Errorf(`field number %d is already mapped to field name "%s" not "%s"`,
			number, f.numberToName[number], name)
	}

	if f.nameToNumber[name] != number {
		return fmt.Errorf(`field name "%s" is already mapped to field number "%d" not "%d"`,
			name, f.nameToNumber[name], number)
	}

	currentIndexOptions, ok := f.indexOptions[name]
	if indexOptions != document.INDEX_OPTIONS_NONE && ok && currentIndexOptions != document.INDEX_OPTIONS_NONE && indexOptions != currentIndexOptions {
		return fmt.Errorf(`cannot change field "%s" from index options=%s  to inconsistent index options=%s`,
			name, currentIndexOptions, indexOptions,
		)
	}
	return nil
}

func (f *FieldNumbers) verifyConsistentDocValuesType(number int, name string, dvType document.DocValuesType) error {
	if f.numberToName[number] != name {
		return fmt.Errorf(`field number %d is already mapped to field name "%s" not "%s"`,
			number, f.numberToName[number], name)
	}

	if f.nameToNumber[name] != number {
		return fmt.Errorf(`field name "%s" is already mapped to field number "%d" not "%d"`,
			name, f.nameToNumber[name], number)
	}

	currentDVType, ok := f.docValuesType[name]
	if dvType != document.DOC_VALUES_TYPE_NONE && ok && currentDVType != document.DOC_VALUES_TYPE_NONE && dvType != currentDVType {
		return fmt.Errorf(`cannot change DocValues type from %s to %d for field "%s"`,
			currentDVType, dvType, name,
		)
	}

	return nil
}

func (f *FieldNumbers) setIndexOptions(number int, name string, indexOptions document.IndexOptions) error {
	if err := f.verifyConsistentIndexOptions(number, name, indexOptions); err != nil {
		return err
	}
	f.indexOptions[name] = indexOptions
	return nil
}

func (f *FieldNumbers) setDocValuesType(number int, name string, dvType document.DocValuesType) error {
	if err := f.verifyConsistentDocValuesType(number, name, dvType); err != nil {
		return err
	}
	f.docValuesType[name] = dvType
	return nil
}

func (f *FieldNumbers) SetDimensions(number int, name string, dimensionCount, indexDimensionCount, dimensionNumBytes int) {
	//f.verifyConsistentDimensions(number, name, dimensionCount, indexDimensionCount, dimensionNumBytes);
	f.dimensions[name] = NewFieldDimensions(dimensionCount, indexDimensionCount, dimensionNumBytes)
}

func (f *FieldNumbers) contains(fieldName string, dvType document.DocValuesType) bool {
	if _, ok := f.nameToNumber[fieldName]; !ok {
		return false
	}
	return dvType == f.docValuesType[fieldName]
}

type FieldDimensions struct {
	DimensionCount      int
	IndexDimensionCount int
	DimensionNumBytes   int
}

func NewFieldDimensions(dimensionCount, indexDimensionCount, dimensionNumBytes int) *FieldDimensions {
	return &FieldDimensions{
		DimensionCount:      dimensionCount,
		IndexDimensionCount: indexDimensionCount,
		DimensionNumBytes:   dimensionNumBytes,
	}
}
