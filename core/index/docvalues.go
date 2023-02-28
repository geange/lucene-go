package index

type DocValues struct {
}

func GetNumeric(reader LeafReader, field string) (NumericDocValues, error) {
	dv, err := reader.GetNumericDocValues(field)
	if err != nil {
		return nil, err
	}
	return dv, nil
}

func GetSorted(reader LeafReader, field string) (SortedDocValues, error) {
	dv, err := reader.GetSortedDocValues(field)
	if err != nil {
		return nil, err
	}
	if dv != nil {
		return dv, nil
	}

	panic("")
}
