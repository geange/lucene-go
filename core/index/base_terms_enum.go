package index

// BaseTermsEnum A base TermsEnum that adds default implementations for
// * attributes()
// * termState()
// * seekExact(BytesRef)
// * seekExact(BytesRef, TermState)
// In some cases, the default implementation may be slow and consume huge memory, so subclass
// SHOULD have its own implementation if possible.
type BaseTermsEnum struct {
	termsEnum TermsEnum
}

type innerTermState struct {
}

func (i *innerTermState) CopyFrom(other TermState) {
	panic("implement me")
}
