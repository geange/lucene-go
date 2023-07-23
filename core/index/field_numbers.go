package index

import (
	"fmt"
	"sync"

	"github.com/geange/lucene-go/core/document"
)

type FieldNumbers struct {
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

	sync.Mutex
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
			return 0, fmt.Errorf(
				`cannot change field %s from index options=%s to inconsistent index options=%s`,
				fieldName, currentOpts, indexOptions,
			)
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
				_, ok := f.numberToName[f.lowestUnassignedFieldNumber+1]
				if !ok {
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
