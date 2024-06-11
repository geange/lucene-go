package index

import "github.com/geange/lucene-go/core/interface/index"

type Query interface {
	// Rewrite
	// Expert: called to re-write queries into primitive queries. For example, a PrefixQuery will be
	// rewritten into a BooleanQuery that consists of TermQuerys.
	Rewrite(reader index.IndexReader) (Query, error)

	// String
	// Convert a query to a string, with field assumed to be the default field and omitted.
	String(field string) string
}
