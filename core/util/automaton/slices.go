package automaton

func grow[T any](s []T, size int) []T {
	if len(s) >= size {
		return s
	}
	var empty T
	add := size - len(s)
	for i := 0; i < add; i++ {
		s = append(s, empty)
	}
	return s
}
