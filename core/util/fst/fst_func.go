package fst

func flag(flags, bit int) bool {
	return flags&bit != 0
}

// Gets the number of bytes required to flag the presence of each arc in the given label range, one bit per arc.
func getNumPresenceBytes(labelRange int) (int, error) {
	err := assert(labelRange >= 0)
	if err != nil {
		return 0, err
	}
	return (labelRange + 7) >> 3, nil
}
