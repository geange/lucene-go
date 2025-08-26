package util

import (
	"errors"
	"io"
)

func Close(closers ...io.Closer) error {
	errs := make([]error, 0)
	for _, closer := range closers {
		if closer == nil {
			continue
		}
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
