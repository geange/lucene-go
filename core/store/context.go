package store

var (
	DEFAULT  = NewIOContext(WithContextType(CONTEXT_DEFAULT))
	READONCE = NewIOContext(WithReadOnce(true))
	READ     = NewIOContext(WithReadOnce(true))
)

// IOContext holds additional details on the merge/search context.
// A IOContext object can never be initialized as null as passed as a parameter to either
// Directory.openInput(String, IOContext) or Directory.createOutput(String, IOContext)
type IOContext struct {
	Type      ContextType
	MergeInfo *MergeInfo
	FlushInfo *FlushInfo
	ReadOnce  bool
}

// ContextType is a enumerator which specifies the context in which the Directory is being used for.
type ContextType int

const (
	CONTEXT_MERGE = ContextType(iota)
	CONTEXT_READ
	CONTEXT_FLUSH
	CONTEXT_DEFAULT
)

type IOContextOption func(ctx *IOContext)

func WithContextType(cType ContextType) IOContextOption {
	return func(ctx *IOContext) {
		ctx.Type = cType
		ctx.ReadOnce = false

	}
}

func WithReadOnce(readOnce bool) IOContextOption {
	return func(ctx *IOContext) {
		ctx.Type = CONTEXT_READ
		ctx.ReadOnce = readOnce
		ctx.FlushInfo = nil
		ctx.MergeInfo = nil
	}
}

func WithFlushInfo(flushInfo *FlushInfo) IOContextOption {
	return func(ctx *IOContext) {
		ctx.FlushInfo = flushInfo
		ctx.Type = CONTEXT_FLUSH
		ctx.MergeInfo = nil
		ctx.ReadOnce = false
	}
}

func WithMergeInfo(mergeInfo *MergeInfo) IOContextOption {
	return func(ctx *IOContext) {
		ctx.FlushInfo = nil
		ctx.Type = CONTEXT_MERGE
		ctx.MergeInfo = mergeInfo
		ctx.ReadOnce = false
	}
}

func NewIOContext(option IOContextOption) *IOContext {
	ctx := newIOContext()
	option(ctx)
	return ctx
}

func newIOContext() *IOContext {
	return &IOContext{
		Type:      CONTEXT_READ,
		MergeInfo: nil,
		FlushInfo: nil,
		ReadOnce:  false,
	}
}
