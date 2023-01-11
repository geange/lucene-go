package index

type DocValues struct {
}

func (DocValues) GetNumeric(reader LeafReader, field string) (NumericDocValues, error) {
	dv, err := reader.GetNumericDocValues(field)
	if err != nil {
		return nil, err
	}
	return dv, nil
}
