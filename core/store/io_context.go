package store

// IOContext holds additional details on the merge/search context. A IOContext object can never be initialized as null as passed as a parameter to either Directory.openInput(String, IOContext) or Directory.createOutput(String, IOContext)
type IOContext struct {
}

type Context int
