package fst

import (
	"errors"
	"fmt"
)

func assert(op bool, msg ...string) error {
	if op {
		return nil
	}
	if len(msg) == 0 {
		return errors.New("assert error")
	}
	return fmt.Errorf("%+v", msg)
}
