package util

import "github.com/geange/lucene-go/core/tokenattributes"

type AttributeSourceV1 struct {
	data *tokenattributes.PackedTokenAttributeIMP
}

func (r *AttributeSourceV1) PackedTokenAttribute() *tokenattributes.PackedTokenAttributeIMP {
	return r.data
}

func (r *AttributeSourceV1) Clear() error {
	return r.data.Clear()
}
