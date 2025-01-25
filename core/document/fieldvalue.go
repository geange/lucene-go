package document

import (
	"fmt"
	"io"
	"math"
)

func Bytes(obj any) ([]byte, error) {
	switch v := obj.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, ErrFieldValueType
	}
}

func Int32(obj any) (int32, error) {
	switch v := obj.(type) {
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case float32:
		return int32(v), nil
	case float64:
		return int32(v), nil
	default:
		return 0, ErrFieldValueType
	}
}

func Int64(obj any) (int64, error) {
	switch v := obj.(type) {
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, ErrFieldValueType
	}
}

func Float32(obj any) (float32, error) {
	switch v := obj.(type) {
	case uint32:
		return math.Float32frombits(v), nil
	default:
		return 0, ErrFieldValueType
	}
}

func Float64(obj any) (float64, error) {
	switch v := obj.(type) {
	case uint64:
		return math.Float64frombits(v), nil
	default:
		return 0, ErrFieldValueType
	}
}

func Str(obj any) (string, error) {
	switch v := obj.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case uint, int, int32, int64, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	default:
		return "", ErrFieldValueType
	}
}

func Reader(obj any) (io.Reader, error) {
	switch v := obj.(type) {
	case io.Reader:
		return v, nil
	default:
		return nil, ErrFieldValueType
	}
}

/*
func (r *Field) I32Value() (int32, error) {
	switch r.fieldsData.(type) {
	case int32:
		return r.fieldsData.(int32), nil
	case int64:
		return int32(r.fieldsData.(int64)), nil
	default:
		return -1, errors.New("fieldsData is not int32")
	}
}

func (r *Field) I64Value() (int64, error) {
	switch r.fieldsData.(type) {
	case int32:
		return int64(r.fieldsData.(int32)), nil
	default:
		return -1, errors.New("fieldsData is not int32")
	}
}

func (r *Field) F32Value() (float32, error) {
	switch v := r.fieldsData.(type) {
	case float32:
		return v, nil
	case int:
		return math.Float32frombits(uint32(v)), nil
	default:
		return -1, errors.New("fieldsData is not float32")
	}
}

func (r *Field) F64Value() (float64, error) {
	switch r.fieldsData.(type) {
	case float64:
		return r.fieldsData.(float64), nil
	default:
		return -1, errors.New("fieldsData is not float64")
	}
}

func (r *Field) StringValue() (string, error) {
	switch r.fieldsData.(type) {
	case string:
		return r.fieldsData.(string), nil
	case []byte:
		return string(r.fieldsData.([]byte)), nil
	default:
		return "", errors.New("fieldsData is not string")
	}
}

func (r *Field) BytesValue() ([]byte, error) {
	switch r.fieldsData.(type) {
	case string:
		return []byte(r.fieldsData.(string)), nil
	case []byte:
		return r.fieldsData.([]byte), nil
	default:
		return nil, errors.New("fieldsData is not []byte")
	}
}

func (r *Field) ReaderValue() (io.Reader, error) {
	reader, ok := r.fieldsData.(io.Reader)
	if !ok {
		return nil, errors.New("fieldsData is not io.Reader")
	}
	return reader, nil
}

*/
