package util

// AttributeImpl Base class for Attributes that can be added to a AttributeSource.
// Attributes are used to add data in a dynamic, yet type-safe way to a source of usually streamed objects,
// e. g. a org.apache.lucene.analysis.TokenStream.
type AttributeImpl interface {
	Clear() error
	End() error
	CopyTo(target AttributeImpl) error
	Clone() error
}
