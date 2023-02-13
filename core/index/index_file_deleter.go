package index

type IndexFileDeleter struct {
}

type RefCount struct {
	fileName string
	initDone bool
	count    int
}
