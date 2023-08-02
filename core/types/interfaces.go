package types

// DocMap A map of doc IDs.
type DocMap interface {
	Get(docId int) int
}
