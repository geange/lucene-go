package store

// TrackingDirectoryWrapper A delegating Directory that records which files were written to and deleted.
type TrackingDirectoryWrapper struct {
	*FilterDirectory

	createdFileNames map[string]struct{}
}

func (t *TrackingDirectoryWrapper) DeleteFile(name string) error {
	if err := t.in.DeleteFile(name); err != nil {
		return err
	}

	delete(t.createdFileNames, name)

	return nil
}

func (t *TrackingDirectoryWrapper) CreateOutput(name string, context *IOContext) (IndexOutput, error) {
	output, err := t.in.CreateOutput(name, context)
	if err != nil {
		return nil, err
	}
	t.createdFileNames[name] = struct{}{}
	return output, nil
}

func (t *TrackingDirectoryWrapper) CreateTempOutput(prefix, suffix string, context *IOContext) (IndexOutput, error) {
	tempOutput, err := t.in.CreateTempOutput(prefix, suffix, context)
	if err != nil {
		return nil, err
	}
	t.createdFileNames[tempOutput.GetName()] = struct{}{}
	return tempOutput, nil
}

func (t *TrackingDirectoryWrapper) CopyFrom(from Directory, src, dest string, context *IOContext) error {
	err := t.FilterDirectory.CopyFrom(from, src, dest, context)
	if err != nil {
		return err
	}
	t.createdFileNames[dest] = struct{}{}
	return nil
}

func (t *TrackingDirectoryWrapper) Rename(source, dest string) error {
	if err := t.in.Rename(source, dest); err != nil {
		return err
	}
	t.createdFileNames[dest] = struct{}{}
	t.createdFileNames[source] = struct{}{}
	return nil
}
