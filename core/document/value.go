package document

type Value interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~uintptr | ~float32 | ~float64 | string | []byte
}

type FieldValueType int

const (
	FieldValueString = FieldValueType(iota)
	FieldValueBytes
	FieldValueI32
	FieldValueI64
	FieldValueF32
	FieldValueF64
	FieldValueReader
	FieldValueOther
)
