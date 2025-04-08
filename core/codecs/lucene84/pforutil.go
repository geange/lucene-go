package lucene84

type PForUtil struct {
}

func allEqual(l []uint64) bool {
	for i := 1; i < len(l); i++ {
		if l[i] != l[0] {
			return false
		}
	}
	return true
}
