package document

// IndexAbleField Represents a single field for indexing. IndexWriter consumes
// []IndexAbleField as a document.
// IndexAbleField代表一个可以被索引的field，每一个Document都是由多个IndexAbleField组成
type IndexAbleField interface {
	// Name 获取Field name
	Name() string

	// FieldType 获取field的属性
	FieldType() IndexAbleFieldType

	// FType 获取Value的类型信息
	FType() FieldValueType

	// Value 内容信息
	Value() interface{}
}

type FieldValueType int

const (
	FVBinary = FieldValueType(iota)
	FVString
	FVReader
	FVNumeric
)
