package index

import (
	"errors"
	"io"
	"math"
)

// IndexSorter Handles how documents should be sorted in an index, both within a segment and
// between segments. Implementers must provide the following methods:
// getDocComparator(LeafReader, int) - an object that determines how documents within a segment
// are to be sorted getComparableProviders(List) - an array of objects that return a sortable
// long value per document and segment getProviderName() - the SPI-registered name of a
// SortFieldProvider to serialize the sort The companion SortFieldProvider should be
// registered with SPI via META-INF/services
type IndexSorter interface {

	// GetComparableProviders Get an array of IndexSorter.ComparableProvider, one per segment,
	// for merge sorting documents in different segments
	// Params: readers – the readers to be merged
	GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error)

	// GetDocComparator Get a comparator that determines the sort order of docs within a single Reader.
	// NB We cannot simply use the FieldComparator API because it requires docIDs to be sent in-order.
	// The default implementations allocate array[maxDoc] to hold native values for comparison, but 1)
	// they are transient (only alive while sorting this one segment) and 2) in the typical index
	// sorting case, they are only used to sort newly flushed segments, which will be smaller than
	// merged segments
	//
	// Params: reader – the Reader to sort
	//		   maxDoc – the number of documents in the Reader
	GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error)

	// GetProviderName The SPI-registered name of a SortFieldProvider that will deserialize the parent SortField
	GetProviderName() string
}

// ComparableProvider Used for sorting documents across segments
// Returns a long so that the natural ordering of long values matches the ordering of doc IDs for the given comparator
type ComparableProvider func(docID int) (int64, error)

// DocComparator A comparator of doc IDs, used for sorting documents within a segment
// Compare docID1 against docID2. The contract for the return value is the same as Comparator.compare(Object, Object).
type DocComparator func(docID1, docID2 int) int

// NumericDocValuesProvider Provide a NumericDocValues instance for a LeafReader
type NumericDocValuesProvider func(reader LeafReader) (NumericDocValues, error)

// SortedDocValuesProvider Provide a SortedDocValues instance for a LeafReader
type SortedDocValuesProvider interface {
	GetNumber(reader LeafReader) (SortedDocValues, error)
}

var _ IndexSorter = &IntSorter{}

// IntSorter Sorts documents based on integer values from a NumericDocValues instance
type IntSorter struct {
	missingValue   *int32
	reverseMul     int
	valuesProvider NumericDocValuesProvider
	providerName   string
}

func NewIntSorter(providerName string, missingValue int32, reverse bool, valuesProvider NumericDocValuesProvider) *IntSorter {
	reverseMul := 1
	if reverse {
		reverseMul = -1
	}

	return &IntSorter{
		missingValue:   &missingValue,
		reverseMul:     reverseMul,
		valuesProvider: valuesProvider,
		providerName:   providerName,
	}
}

func (i *IntSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	providers := make([]ComparableProvider, 0)
	missingValue := int32(0)
	if i.missingValue != nil {
		missingValue = *i.missingValue
	}

	for readerIndex, reader := range readers {
		values, err := i.valuesProvider(reader)
		if err != nil {
			return nil, err
		}
		providers[readerIndex] = func(docID int) (int64, error) {
			ok, err := values.AdvanceExact(docID)
			if err != nil {
				return 0, err
			}
			if ok {
				return values.LongValue()
			}
			return int64(missingValue), nil
		}
	}
	return providers, nil
}

func (i *IntSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	dvs, err := i.valuesProvider(reader)
	if err != nil {
		return nil, err
	}
	values := make([]int32, maxDoc)
	if i.missingValue != nil {
		for idx := range values {
			values[idx] = *i.missingValue
		}
	}

	for {
		docID, err := dvs.NextDoc()
		if err != nil {
			return nil, err
		}
		if docID == NO_MORE_DOCS {
			break
		}
		value, err := dvs.LongValue()
		if err != nil {
			return nil, err
		}

		values[docID] = int32(value)
	}

	return func(docID1, docID2 int) int {
		return i.reverseMul * compare(values[docID1], values[docID2])
	}, nil
}

func (i *IntSorter) GetProviderName() string {
	return i.providerName
}

var _ IndexSorter = &LongSorter{}

// LongSorter Sorts documents based on long values from a NumericDocValues instance
type LongSorter struct {
	missingValue   *int64
	reverseMul     int
	valuesProvider NumericDocValuesProvider
	providerName   string
}

func NewLongSorter(providerName string, missingValue int64,
	reverse bool, valuesProvider NumericDocValuesProvider) *LongSorter {
	reverseMul := 1
	if reverse {
		reverseMul = -1
	}

	return &LongSorter{
		missingValue:   &missingValue,
		reverseMul:     reverseMul,
		valuesProvider: valuesProvider,
		providerName:   providerName,
	}
}

func (i *LongSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	providers := make([]ComparableProvider, 0)
	missingValue := int64(0)
	if i.missingValue != nil {
		missingValue = *i.missingValue
	}

	for readerIndex, reader := range readers {
		values, err := i.valuesProvider(reader)
		if err != nil {
			return nil, err
		}
		providers[readerIndex] = func(docID int) (int64, error) {
			ok, err := values.AdvanceExact(docID)
			if err != nil {
				return 0, err
			}
			if ok {
				return values.LongValue()
			}
			return int64(missingValue), nil
		}
	}
	return providers, nil
}

func (i *LongSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	dvs, err := i.valuesProvider(reader)
	if err != nil {
		return nil, err
	}
	values := make([]int64, maxDoc)
	if i.missingValue != nil {
		for idx := range values {
			values[idx] = *i.missingValue
		}
	}

	for {
		docID, err := dvs.NextDoc()
		if err != nil {
			return nil, err
		}
		if docID == NO_MORE_DOCS {
			break
		}
		value, err := dvs.LongValue()
		if err != nil {
			return nil, err
		}

		values[docID] = value
	}

	return func(docID1, docID2 int) int {
		return i.reverseMul * compare(values[docID1], values[docID2])
	}, nil
}

func (i *LongSorter) GetProviderName() string {
	return i.providerName
}

var _ IndexSorter = &FloatSorter{}

// FloatSorter Sorts documents based on float values from a NumericDocValues instance
type FloatSorter struct {
	missingValue   *float32
	reverseMul     int
	valuesProvider NumericDocValuesProvider
	providerName   string
}

func NewFloatSorter(providerName string, missingValue float32,
	reverse bool, valuesProvider NumericDocValuesProvider) *FloatSorter {
	reverseMul := 1
	if reverse {
		reverseMul = -1
	}

	return &FloatSorter{
		missingValue:   &missingValue,
		reverseMul:     reverseMul,
		valuesProvider: valuesProvider,
		providerName:   providerName,
	}
}

func (f *FloatSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	providers := make([]ComparableProvider, 0)
	missingValue := float32(0)
	if f.missingValue != nil {
		missingValue = *f.missingValue
	}

	for readerIndex, reader := range readers {
		values, err := f.valuesProvider(reader)
		if err != nil {
			return nil, err
		}
		providers[readerIndex] = func(docID int) (int64, error) {
			value := missingValue
			ok, err := values.AdvanceExact(docID)
			if err != nil {
				return 0, err
			}
			if ok {
				v, err := values.LongValue()
				if err != nil {
					return 0, err
				}
				value = math.Float32frombits(uint32(v))
			}
			return int64(math.Float32bits(value)), nil
		}
	}
	return providers, nil
}

func (f *FloatSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	dvs, err := f.valuesProvider(reader)
	if err != nil {
		return nil, err
	}
	values := make([]float32, maxDoc)
	if f.missingValue != nil {
		for idx := range values {
			values[idx] = *f.missingValue
		}
	}

	for {
		docID, err := dvs.NextDoc()
		if err != nil {
			return nil, err
		}
		if docID == NO_MORE_DOCS {
			break
		}
		value, err := dvs.LongValue()
		if err != nil {
			return nil, err
		}
		values[docID] = math.Float32frombits(uint32(value))
	}

	return func(docID1, docID2 int) int {
		return f.reverseMul * compare(values[docID1], values[docID2])
	}, nil
}

func (f *FloatSorter) GetProviderName() string {
	return f.providerName
}

var _ IndexSorter = &DoubleSorter{}

// DoubleSorter Sorts documents based on double values from a NumericDocValues instance
type DoubleSorter struct {
	missingValue   *float64
	reverseMul     int
	valuesProvider NumericDocValuesProvider
	providerName   string
}

func NewDoubleSorter(providerName string, missingValue float64,
	reverse bool, valuesProvider NumericDocValuesProvider) *DoubleSorter {
	reverseMul := 1
	if reverse {
		reverseMul = -1
	}

	return &DoubleSorter{
		missingValue:   &missingValue,
		reverseMul:     reverseMul,
		valuesProvider: valuesProvider,
		providerName:   providerName,
	}
}

func (d *DoubleSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	providers := make([]ComparableProvider, 0)
	missingValue := float64(0)
	if d.missingValue != nil {
		missingValue = *d.missingValue
	}

	for readerIndex, reader := range readers {
		values, err := d.valuesProvider(reader)
		if err != nil {
			return nil, err
		}
		providers[readerIndex] = func(docID int) (int64, error) {
			value := missingValue
			ok, err := values.AdvanceExact(docID)
			if err != nil {
				return 0, err
			}
			if ok {
				v, err := values.LongValue()
				if err != nil {
					return 0, err
				}
				value = math.Float64frombits(uint64(v))
			}
			return int64(math.Float64bits(value)), nil
		}
	}
	return providers, nil
}

func (d *DoubleSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	dvs, err := d.valuesProvider(reader)
	if err != nil {
		return nil, err
	}
	values := make([]float64, maxDoc)
	if d.missingValue != nil {
		for idx := range values {
			values[idx] = *d.missingValue
		}
	}

	for {
		docID, err := dvs.NextDoc()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if docID == NO_MORE_DOCS {
			break
		}
		value, err := dvs.LongValue()
		if err != nil {
			return nil, err
		}
		values[docID] = math.Float64frombits(uint64(value))
	}

	return func(docID1, docID2 int) int {
		return d.reverseMul * compare(values[docID1], values[docID2])
	}, nil
}

func (d *DoubleSorter) GetProviderName() string {
	return d.providerName
}

func compare[T int32 | int64 | float32 | float64](x, y T) int {
	if x < y {
		return -1
	} else if x == y {
		return 0
	} else {
		return 1
	}
}
