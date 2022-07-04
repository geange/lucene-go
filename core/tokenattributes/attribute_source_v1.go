package tokenattributes

type AttributeSourceV1 struct {
	data *PackedTokenAttributeIMP
}

func NewAttributeSourceV1() *AttributeSourceV1 {
	return &AttributeSourceV1{data: NewPackedTokenAttributeIMP()}
}

func (r *AttributeSourceV1) PackedTokenAttribute() *PackedTokenAttributeIMP {
	return r.data
}

func (r *AttributeSourceV1) Clear() error {
	return r.data.Clear()
}
