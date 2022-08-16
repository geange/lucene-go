package fst

type Outputs interface {
	Merge(first, second any) any
}
