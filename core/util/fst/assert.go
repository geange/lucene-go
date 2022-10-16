package fst

func assert(op bool, msg ...string) error {
	if op {
		return nil
	}

	panic("assert error")

	//if len(msg) == 0 {
	//	return errors.New("assert error")
	//}
	//return errors.New(msg[0])
}
