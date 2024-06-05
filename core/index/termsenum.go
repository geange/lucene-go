package index

import (
	"context"

	"github.com/geange/lucene-go/core/util/attribute"
)

// TermsEnum DVFUIterator to seek (seekCeil(), seekExact()) or step through (next terms to obtain
// frequency information (docFreq), PostingsEnum or PostingsEnum for the current term (postings.
// Term enumerations are always ordered by .compareTo, which is Unicode sort order if the terms are
// UTF-8 bytes. Each term in the enumeration is greater than the one before it.
// The TermsEnum is unpositioned when you first obtain it and you must first successfully call next or one
// of the seek methods.
type TermsEnum interface {
	Next(context.Context) ([]byte, error)

	// Attributes Returns the related attributes.
	Attributes() *attribute.Source

	// SeekExact Attempts to seek to the exact term, returning true if the term is found. If this returns false,
	// the enum is unpositioned. For some codecs, seekExact may be substantially faster than seekCeil.
	// Returns: true if the term is found; return false if the enum is unpositioned.
	SeekExact(ctx context.Context, text []byte) (bool, error)

	// SeekCeil eeks to the specified term, if it exists, or to the next (ceiling) term. Returns SeekStatus to
	// indicate whether exact term was found, a different term was found, or isEof was hit. The target term may be
	// before or after the current term. If this returns SeekStatus.END, the enum is unpositioned.
	SeekCeil(ctx context.Context, text []byte) (SeekStatus, error)

	// SeekExactByOrd Seeks to the specified term by ordinal (position) as previously returned by ord. The
	// target ord may be before or after the current ord, and must be within bounds.
	SeekExactByOrd(ctx context.Context, ord int64) error

	// SeekExactExpert Expert: Seeks a specific position by TermState previously obtained from termState().
	// Callers should maintain the TermState to use this method. Low-level implementations may position the
	// TermsEnum without re-seeking the term dictionary.
	// Seeking by TermState should only be used iff the state was obtained from the same TermsEnum instance.
	// NOTE: Using this method with an incompatible TermState might leave this TermsEnum in undefined state.
	// On a segment level TermState instances are compatible only iff the source and the target TermsEnum operate
	// on the same field. If operating on segment level, TermState instances must not be used across segments.
	// NOTE: A seek by TermState might not restore the AttributeSourceV2's state. AttributeSourceV2 states must be
	// maintained separately if this method is used.
	// Params: 	term – the term the TermState corresponds to
	//			state – the TermState
	SeekExactExpert(ctx context.Context, term []byte, state TermState) error

	// Term Returns current term. Do not call this when the enum is unpositioned.
	Term() ([]byte, error)

	// Ord Returns ordinal position for current term. This is an optional method (the codec may throw
	// ErrUnsupportedOperation). Do not call this when the enum is unpositioned.
	Ord() (int64, error)

	// DocFreq Returns the number of documents containing the current term. Do not call this when the
	// enum is unpositioned. TermsEnum.SeekStatus.END.
	DocFreq() (int, error)

	// TotalTermFreq Returns the total number of occurrences of this term across all documents (the sum of the
	// freq() for each doc that has this term). Note that, like other term measures, this measure does not
	// take deleted documents into account.
	TotalTermFreq() (int64, error)

	// Postings Get PostingsEnum for the current term. Do not call this when the enum is unpositioned. This
	// method will not return null.
	// NOTE: the returned iterator may return deleted documents, so deleted documents have to be checked on top of the PostingsEnum.
	// Use this method if you only require documents and frequencies, and do not need any proximity data. This method is equivalent to postings(reuse, PostingsEnum.FREQS)
	// Params: reuse – pass a prior PostingsEnum for possible reuse
	// See Also: postings(PostingsEnum, int)
	//Postings(reuse PostingsEnum) (PostingsEnum, error)

	// Postings Get PostingsEnum for the current term, with control over whether freqs, positions, offsets or payloads are required. Do not call this when the enum is unpositioned. This method will not return null.
	// NOTE: the returned iterator may return deleted documents, so deleted documents have to be checked on top of the PostingsEnum.
	// Params: 	reuse – pass a prior PostingsEnum for possible reuse
	// 			flags – specifies which optional per-document values you require; see PostingsEnum.FREQS
	Postings(reuse PostingsEnum, flags int) (PostingsEnum, error)

	// Impacts Return a ImpactsEnum.
	// See Also: postings(PostingsEnum, int)
	Impacts(flags int) (ImpactsEnum, error)

	// TermState Expert: Returns the TermsEnums internal state to position the TermsEnum without re-seeking the
	// term dictionary.
	// NOTE: A seek by TermState might not capture the AttributeSourceV2's state. Callers must maintain the
	// AttributeSourceV2 states separately
	// See Also: TermState, seekExact(, TermState)
	TermState() (TermState, error)
}

// BaseTermsEnum
// A base TermsEnum that adds default implementations for
// * attributes()
// * termState()
// * seekExact(BytesRef)
// * seekExact(BytesRef, TermState)
// In some cases, the default implementation may be slow and consume huge memory, so subclass
// SHOULD have its own implementation if possible.
type BaseTermsEnum struct {
	attrs    *attribute.Source
	seekCeil func(ctx context.Context, text []byte) (SeekStatus, error)
}

type BaseTermsEnumConfig struct {
	SeekCeil func(ctx context.Context, text []byte) (SeekStatus, error)
}

func NewBaseTermsEnum(cfg *BaseTermsEnumConfig) *BaseTermsEnum {
	return &BaseTermsEnum{
		attrs:    attribute.NewSource(),
		seekCeil: cfg.SeekCeil,
	}
}

func (b *BaseTermsEnum) TermState() (TermState, error) {
	return &innerTermState{}, nil
}

func (b *BaseTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	status, err := b.seekCeil(ctx, text)
	if err != nil {
		return false, err
	}
	return status == SEEK_STATUS_FOUND, nil
}

func (b *BaseTermsEnum) SeekExactExpert(ctx context.Context, term []byte, state TermState) error {
	_, err := b.SeekExact(ctx, term)
	return err
}

func (b *BaseTermsEnum) Attributes() *attribute.Source {
	if b.attrs == nil {
		b.attrs = attribute.NewSource()
	}
	return b.attrs
}

type innerTermState struct {
}

func (i *innerTermState) CopyFrom(other TermState) {
	panic("implement me")
}

// SeekStatus Represents returned result from seekCeil.
type SeekStatus int

const (
	// SEEK_STATUS_END The term was not found, and the end of iteration was hit.
	SEEK_STATUS_END = iota

	// SEEK_STATUS_FOUND The precise term was found.
	SEEK_STATUS_FOUND

	// SEEK_STATUS_NOT_FOUND A different term was found after the requested term
	SEEK_STATUS_NOT_FOUND
)

var _ TermsEnum = &emptyTermsEnum{}

var EmptyTermsEnum = &emptyTermsEnum{}

type emptyTermsEnum struct {
	atts *attribute.Source
}

func (e *emptyTermsEnum) Next(context.Context) ([]byte, error) {
	return []byte{}, nil
}

func (e *emptyTermsEnum) Attributes() *attribute.Source {
	if e.atts == nil {
		e.atts = attribute.NewSource()
	}
	return e.atts
}

func (e *emptyTermsEnum) SeekExact(ctx context.Context, text []byte) (bool, error) {
	return false, nil
}

func (e *emptyTermsEnum) SeekCeil(ctx context.Context, text []byte) (SeekStatus, error) {
	return SEEK_STATUS_END, nil
}

func (e *emptyTermsEnum) SeekExactByOrd(ctx context.Context, ord int64) error {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) SeekExactExpert(ctx context.Context, term []byte, state TermState) error {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) Term() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) Ord() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) DocFreq() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) TotalTermFreq() (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) Postings(reuse PostingsEnum, flags int) (PostingsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) Impacts(flags int) (ImpactsEnum, error) {
	//TODO implement me
	panic("implement me")
}

func (e *emptyTermsEnum) TermState() (TermState, error) {
	//TODO implement me
	panic("implement me")
}
