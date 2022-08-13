package store

// IOContext holds additional details on the merge/search context. A IOContext object can never be initialized as null as passed as a parameter to either Directory.openInput(String, IOContext) or Directory.createOutput(String, IOContext)
type IOContext struct {
	//An object of a enumerator Context type
	context   *Context
	mergeInfo *MergeInfo
	flushInfo *FlushInfo
	readOnce  bool
}

// Context is a enumerator which specifies the context in which the Directory is being used for.
type Context int

const (
	MERGE = Context(iota)
	READ
	FLUSH
	DEFAULT
)
