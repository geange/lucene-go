package index

import "bytes"

var _ FilteredTermsEnum = &SingleTermsEnum{}

// SingleTermsEnum Subclass of FilteredTermsEnum for enumerating a single term.
// For example, this can be used by MultiTermQuerys that need only visit one term,
// but want to preserve MultiTermQuery semantics such as MultiTermQuery.getRewriteMethod.
type SingleTermsEnum struct {
	singleRef []byte
	*FilteredTermsEnumBase
}

func NewSingleTermsEnum(tenum TermsEnum, termText []byte) *SingleTermsEnum {
	enum := &SingleTermsEnum{
		singleRef: termText,
	}
	enum.FilteredTermsEnumBase = NewFilteredTermsEnumDefault(&FilteredTermsEnumDefaultConfig{
		Accept:        enum.Accept,
		NextSeekTerm:  nil,
		Tenum:         tenum,
		StartWithSeek: true,
	})
	enum.setInitialSeekTerm(termText)
	return enum
}

func (s *SingleTermsEnum) Accept(term []byte) (AcceptStatus, error) {
	if bytes.Equal(term, s.singleRef) {
		return ACCEPT_STATUS_YES, nil
	}
	return ACCEPT_STATUS_END, nil
}
