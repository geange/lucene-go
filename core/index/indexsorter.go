package index

import (
	"errors"
	"github.com/geange/lucene-go/core/types"
	"io"
	"math"
)

// IndexSorter Handles how documents should be sorted in an index, both within a segment and
// between segments. Implementers must provide the following methods:
// getDocComparator(LeafReader, int) - an object that determines how documents within a segment
// are to be sorted getComparableProviders(List) - an array of objects that return a sortable
// long item per document and segment getProviderName() - the SPI-registered name of a
// SortFieldProvider to serialize the sort The companion SortFieldProvider should be
// registered with SPI via META-INF/services
type IndexSorter interface {

	// GetComparableProviders
	// Get an array of IndexSorter.ComparableProvider, one per segment,
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
// 用于跨多个段（segment）进行文档排序
type ComparableProvider interface {
	// GetAsComparableLong Returns a long so that the natural ordering of long values
	// matches the ordering of doc IDs for the given comparator
	GetAsComparableLong(docID int) (int64, error)
}

// DocComparator A comparator of doc IDs, used for sorting documents within a segment
// 用于段内文档的排序
type DocComparator interface {
	// Compare docID1 against docID2. The contract for the return item is
	// the same as cmp.Compare(Object, Object).
	Compare(docID1, docID2 int) int
}

// NumericDocValuesProvider Provide a NumericDocValues instance for a LeafReader
type NumericDocValuesProvider interface {
	Get(reader LeafReader) (NumericDocValues, error)
}

var _ NumericDocValuesProvider = &EmptyNumericDocValuesProvider{}

type EmptyNumericDocValuesProvider struct {
	FnGet func(reader LeafReader) (NumericDocValues, error)
}

func (e *EmptyNumericDocValuesProvider) Get(reader LeafReader) (NumericDocValues, error) {
	return e.FnGet(reader)
}

// SortedDocValuesProvider Provide a SortedDocValues instance for a LeafReader
type SortedDocValuesProvider interface {
	Get(reader LeafReader) (SortedDocValues, error)
}

var _ SortedDocValuesProvider = &EmptySortedDocValuesProvider{}

type EmptySortedDocValuesProvider struct {
	FnGet func(reader LeafReader) (SortedDocValues, error)
}

func (e *EmptySortedDocValuesProvider) Get(reader LeafReader) (SortedDocValues, error) {
	return e.FnGet(reader)
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

var _ ComparableProvider = &IntComparableProvider{}

type IntComparableProvider struct {
	values       NumericDocValues
	missingValue int64
}

func (r *IntComparableProvider) GetAsComparableLong(docID int) (int64, error) {
	ok, err := r.values.AdvanceExact(docID)
	if err != nil {
		return 0, err
	}
	if ok {
		return r.values.LongValue()
	}
	return r.missingValue, nil
}

func (i *IntSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	providers := make([]ComparableProvider, 0)
	missingValue := int64(0)
	if i.missingValue != nil {
		missingValue = int64(*i.missingValue)
	}

	for readerIndex, reader := range readers {
		values, err := i.valuesProvider.Get(reader)
		if err != nil {
			return nil, err
		}
		providers[readerIndex] = &IntComparableProvider{
			values:       values,
			missingValue: missingValue,
		}
	}
	return providers, nil
}

var _ DocComparator = &IntDocComparator{}

type IntDocComparator struct {
	values     []int32
	reverseMul int
}

func (r *IntDocComparator) Compare(docID1, docID2 int) int {
	return r.reverseMul * Compare(r.values[docID1], r.values[docID2])
}

func (i *IntSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	dvs, err := i.valuesProvider.Get(reader)
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
		if docID == types.NO_MORE_DOCS {
			break
		}
		value, err := dvs.LongValue()
		if err != nil {
			return nil, err
		}

		values[docID] = int32(value)
	}

	return &IntDocComparator{
		values:     values,
		reverseMul: i.reverseMul,
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

var _ ComparableProvider = &LongComparableProvider{}

type LongComparableProvider struct {
	values       NumericDocValues
	missingValue int64
}

func (r *LongComparableProvider) GetAsComparableLong(docID int) (int64, error) {
	ok, err := r.values.AdvanceExact(docID)
	if err != nil {
		return 0, err
	}
	if ok {
		return r.values.LongValue()
	}
	return r.missingValue, nil
}

func (i *LongSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	providers := make([]ComparableProvider, 0)
	missingValue := int64(0)
	if i.missingValue != nil {
		missingValue = *i.missingValue
	}

	for readerIndex, reader := range readers {
		values, err := i.valuesProvider.Get(reader)
		if err != nil {
			return nil, err
		}
		providers[readerIndex] = &LongComparableProvider{
			values:       values,
			missingValue: missingValue,
		}
	}
	return providers, nil
}

var _ DocComparator = &LongDocComparator{}

type LongDocComparator struct {
	values     []int64
	reverseMul int
}

func (r *LongDocComparator) Compare(docID1, docID2 int) int {
	//TODO implement me
	panic("implement me")
}

func (i *LongSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	dvs, err := i.valuesProvider.Get(reader)
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
		if docID == types.NO_MORE_DOCS {
			break
		}
		value, err := dvs.LongValue()
		if err != nil {
			return nil, err
		}

		values[docID] = value
	}

	return &LongDocComparator{
		values:     values,
		reverseMul: i.reverseMul,
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

var _ ComparableProvider = &FloatComparableProvider{}

type FloatComparableProvider struct {
	values       NumericDocValues
	missingValue float32
}

func (r *FloatComparableProvider) GetAsComparableLong(docID int) (int64, error) {
	value := r.missingValue
	ok, err := r.values.AdvanceExact(docID)
	if err != nil {
		return 0, err
	}
	if ok {
		v, err := r.values.LongValue()
		if err != nil {
			return 0, err
		}
		value = math.Float32frombits(uint32(v))
	}
	return int64(math.Float32bits(value)), nil
}

func (f *FloatSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	providers := make([]ComparableProvider, 0)
	missingValue := float32(0)
	if f.missingValue != nil {
		missingValue = *f.missingValue
	}

	for readerIndex, reader := range readers {
		values, err := f.valuesProvider.Get(reader)
		if err != nil {
			return nil, err
		}
		providers[readerIndex] = &FloatComparableProvider{
			values:       values,
			missingValue: missingValue,
		}
	}
	return providers, nil
}

var _ DocComparator = &FloatDocComparator{}

type FloatDocComparator struct {
	values     []float32
	reverseMul int
}

func (f *FloatDocComparator) Compare(docID1, docID2 int) int {
	return f.reverseMul * Compare(f.values[docID1], f.values[docID2])
}

func (f *FloatSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	dvs, err := f.valuesProvider.Get(reader)
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
		if docID == types.NO_MORE_DOCS {
			break
		}
		value, err := dvs.LongValue()
		if err != nil {
			return nil, err
		}
		values[docID] = math.Float32frombits(uint32(value))
	}

	return &FloatDocComparator{
		values:     values,
		reverseMul: f.reverseMul,
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

var _ ComparableProvider = &DoubleComparableProvider{}

type DoubleComparableProvider struct {
	values       NumericDocValues
	missingValue float64
}

func (d *DoubleComparableProvider) GetAsComparableLong(docID int) (int64, error) {
	value := d.missingValue
	ok, err := d.values.AdvanceExact(docID)
	if err != nil {
		return 0, err
	}
	if ok {
		v, err := d.values.LongValue()
		if err != nil {
			return 0, err
		}
		value = math.Float64frombits(uint64(v))
	}
	return int64(math.Float64bits(value)), nil
}

func (d *DoubleSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	providers := make([]ComparableProvider, 0)
	missingValue := float64(0)
	if d.missingValue != nil {
		missingValue = *d.missingValue
	}

	for readerIndex, reader := range readers {
		values, err := d.valuesProvider.Get(reader)
		if err != nil {
			return nil, err
		}
		providers[readerIndex] = &DoubleComparableProvider{
			values:       values,
			missingValue: missingValue,
		}
	}
	return providers, nil
}

var _ DocComparator = &DoubleDocComparator{}

type DoubleDocComparator struct {
	values     []float64
	reverseMul int
}

func (d *DoubleDocComparator) Compare(docID1, docID2 int) int {
	return d.reverseMul * Compare(d.values[docID1], d.values[docID2])
}

func (d *DoubleSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	dvs, err := d.valuesProvider.Get(reader)
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
		if docID == types.NO_MORE_DOCS {
			break
		}
		value, err := dvs.LongValue()
		if err != nil {
			return nil, err
		}
		values[docID] = math.Float64frombits(uint64(value))
	}

	return &DoubleDocComparator{
		values:     values,
		reverseMul: d.reverseMul,
	}, nil
}

func (d *DoubleSorter) GetProviderName() string {
	return d.providerName
}

var _ IndexSorter = &StringSorter{}

type StringSorter struct {
	providerName   string
	missingValue   string
	reverseMul     int
	valuesProvider SortedDocValuesProvider
}

func (s *StringSorter) GetComparableProviders(readers []LeafReader) ([]ComparableProvider, error) {
	//readersNum := len(readers)
	//
	//providers := make([]ComparableProvider, readersNum)
	//values := make([]SortedDocValues, readersNum)
	//
	//for i, reader := range readers {
	//	sorted, err := s.valuesProvider.Get(reader)
	//	if err != nil {
	//		return nil, err
	//	}
	//	values[i] = sorted
	//}

	// TODO
	panic("")
}

func NewStringSorter(providerName, missingValue string, reverse bool, valuesProvider SortedDocValuesProvider) *StringSorter {
	return &StringSorter{
		providerName: providerName,
		missingValue: missingValue,
		reverseMul: func() int {
			if reverse {
				return -1
			}
			return 1
		}(),
		valuesProvider: valuesProvider,
	}
}

var _ DocComparator = &StringDocComparator{}

type StringDocComparator struct {
	ords       []int
	reverseMul int
}

func (s *StringDocComparator) Compare(docID1, docID2 int) int {
	return s.reverseMul * Compare(s.ords[docID1], s.ords[docID2])
}

func (s *StringSorter) GetDocComparator(reader LeafReader, maxDoc int) (DocComparator, error) {
	sorted, err := s.valuesProvider.Get(reader)
	if err != nil {
		return nil, err
	}

	missingOrd := math.MinInt32
	if s.missingValue == "STRING_LAST" {
		missingOrd = math.MaxInt32
	}

	ords := make([]int, maxDoc)
	for i := range ords {
		ords[i] = missingOrd
	}

	for {
		docID, err := sorted.NextDoc()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		value, err := sorted.OrdValue()
		if err != nil {
			return nil, err
		}
		ords[docID] = value
	}

	return &StringDocComparator{
		ords:       nil,
		reverseMul: 0,
	}, nil
}

func (s *StringSorter) GetProviderName() string {
	return s.providerName
}

func Compare[T int | int32 | int64 | float32 | float64](x, y T) int {
	if x < y {
		return -1
	} else if x == y {
		return 0
	} else {
		return 1
	}
}
