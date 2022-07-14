package search

// Scorer Expert: Common scoring functionality for different types of queries.
// 不同类型查询的通用评分功能。
//
// A Scorer exposes an iterator() over documents matching a query in increasing order of doc Id.
// 计分器暴露一个迭代器，这个迭代器按照文档id递增顺序
//
// Document scores are computed using a given Similarity implementation.
// NOTE: The values Float.Nan, Float.NEGATIVE_INFINITY and Float.POSITIVE_INFINITY are not valid scores.
// Certain collectors (eg TopScoreDocCollector) will not properly collect hits with these scores.
type Scorer interface {
	Scorable
}
