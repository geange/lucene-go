package selector

type Selector interface {
	SelectK(from, to, k int)
}
