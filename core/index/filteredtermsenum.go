package index

import (
	"context"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/util/attribute"
)

type FilteredTermsEnum interface {
	index.TermsEnum

	// Accept Return if term is accepted, not accepted or the iteration should ended (and possibly seek).
	Accept(term []byte) (AcceptStatus, error)
}

// AcceptStatus Return item, if term should be accepted or the iteration should END.
// The *_SEEK values denote, that after handling the current term the enum should
// call nextSeekTerm and step forward.
// See Also: accept(BytesRef)
type AcceptStatus int

const (
	// ACCEPT_STATUS_YES Accept the term and position the enum at the next term.
	ACCEPT_STATUS_YES = AcceptStatus(iota)

	// ACCEPT_STATUS_YES_AND_SEEK Accept the term and advance (nextSeekTerm(BytesRef)) to the next term.
	ACCEPT_STATUS_YES_AND_SEEK

	// ACCEPT_STATUS_NO Reject the term and position the enum at the next term.
	ACCEPT_STATUS_NO

	// ACCEPT_STATUS_NO_AND_SEEK Reject the term and advance (nextSeekTerm(BytesRef)) to the next term.
	ACCEPT_STATUS_NO_AND_SEEK

	// ACCEPT_STATUS_END Reject the term and stop enumerating.
	ACCEPT_STATUS_END
)

type FilteredTermsEnumDefaultConfig struct {
	Accept        func(term []byte) (AcceptStatus, error)
	NextSeekTerm  func(currentTerm []byte) ([]byte, error)
	Tenum         index.TermsEnum
	StartWithSeek bool
}

type FilteredTermsEnumBase struct {
	initialSeekTerm []byte
	doSeek          bool
	actualTerm      []byte          // Which term the enum is currently positioned to.
	tenum           index.TermsEnum // The delegate TermsEnum.

	Accept       func(term []byte) (AcceptStatus, error)
	NextSeekTerm func(currentTerm []byte) ([]byte, error)
}

func NewFilteredTermsEnumDefault(cfg *FilteredTermsEnumDefaultConfig) *FilteredTermsEnumBase {
	return &FilteredTermsEnumBase{
		initialSeekTerm: nil,
		doSeek:          cfg.StartWithSeek,
		actualTerm:      nil,
		tenum:           cfg.Tenum,
		Accept:          cfg.Accept,
	}
}

// Use this method to set the initial BytesRef to seek before iterating.
// This is a convenience method for subclasses that do not override nextSeekTerm.
// If the initial seek term is null (default), the enum is empty.
// You can only use this method, if you keep the default implementation of nextSeekTerm.
func (f *FilteredTermsEnumBase) setInitialSeekTerm(term []byte) {
	f.initialSeekTerm = term
}

// On the first call to next or if accept returns FilteredTermsEnum.AcceptStatus.YES_AND_SEEK or
// FilteredTermsEnum.AcceptStatus.NO_AND_SEEK, this method will be called to eventually seek the
// underlying TermsEnum to a new position. On the first call, currentTerm will be null, later
// calls will provide the term the underlying enum is positioned at. This method returns per
// default only one time the initial seek term and then null, so no repositioning is ever done.
// Override this method, if you want a more sophisticated TermsEnum, that repositions the iterator
// during enumeration. If this method always returns null the enum is empty.
// Please note: This method should always provide a greater term than the last enumerated term,
// else the behaviour of this enum violates the contract for TermsEnums.
func (f *FilteredTermsEnumBase) nextSeekTerm(currentTerm []byte) ([]byte, error) {
	if f.NextSeekTerm != nil {
		return f.NextSeekTerm(currentTerm)
	}
	t := f.initialSeekTerm
	f.initialSeekTerm = nil
	return t, nil
}

// Attributes Returns the related attributes, the returned AttributeSource is shared with the delegate TermsEnum.
func (f *FilteredTermsEnumBase) Attributes() *attribute.Source {
	return f.tenum.Attributes()
}

func (f *FilteredTermsEnumBase) Term() ([]byte, error) {
	return f.tenum.Term()
}

func (f *FilteredTermsEnumBase) DocFreq() (int, error) {
	return f.tenum.DocFreq()
}

func (f *FilteredTermsEnumBase) TotalTermFreq() (int64, error) {
	return f.tenum.TotalTermFreq()
}

func (f *FilteredTermsEnumBase) SeekExact(ctx context.Context, text []byte) (bool, error) {
	return f.tenum.SeekExact(ctx, text)
}

func (f *FilteredTermsEnumBase) SeekCeil(ctx context.Context, text []byte) (index.SeekStatus, error) {
	return f.tenum.SeekCeil(ctx, text)
}

func (f *FilteredTermsEnumBase) SeekExactByOrd(ctx context.Context, ord int64) error {
	return f.tenum.SeekExactByOrd(nil, ord)
}

func (f *FilteredTermsEnumBase) Ord() (int64, error) {
	return f.tenum.Ord()
}

func (f *FilteredTermsEnumBase) Postings(reuse index.PostingsEnum, flags int) (index.PostingsEnum, error) {
	return f.tenum.Postings(reuse, flags)
}

func (f *FilteredTermsEnumBase) Impacts(flags int) (index.ImpactsEnum, error) {
	return f.tenum.Impacts(flags)
}

// SeekExactExpert This enum does not support seeking!
// Throws: ErrUnsupportedOperation – In general, subclasses do not support seeking.
func (f *FilteredTermsEnumBase) SeekExactExpert(ctx context.Context, term []byte, state index.TermState) error {
	return f.tenum.SeekExactExpert(ctx, term, state)
}

// TermState Returns the filtered enums term state
func (f *FilteredTermsEnumBase) TermState() (index.TermState, error) {
	return f.tenum.TermState()
}

func (f *FilteredTermsEnumBase) Next(context.Context) ([]byte, error) {
	// System.out.println("FTE.next doSeek=" + doSeek);
	// new Throwable().printStackTrace(System.out);
	var err error
	for {
		// Seek or forward the iterator
		if f.doSeek {
			f.doSeek = false
			t, err := f.nextSeekTerm(f.actualTerm)
			if err != nil {
				return nil, err
			}

			if len(t) == 0 {
				return nil, nil
			}

			if v, err := f.tenum.SeekCeil(nil, t); err != nil {
				return nil, err
			} else if v == index.SEEK_STATUS_END {
				return nil, nil
			}

			f.actualTerm, err = f.tenum.Term()
			if err != nil {
				return nil, err
			}
		} else {
			f.actualTerm, err = f.tenum.Term()
			if err != nil {
				return nil, err
			}
			if f.actualTerm == nil {
				return nil, nil
			}
		}

		status, err := f.Accept(f.actualTerm)
		if err != nil {
			return nil, err
		}
		switch status {
		case ACCEPT_STATUS_YES_AND_SEEK:
			f.doSeek = true
		case ACCEPT_STATUS_YES:
			return f.actualTerm, nil
		case ACCEPT_STATUS_NO_AND_SEEK:
			f.doSeek = true
			return nil, nil
		case ACCEPT_STATUS_END:
			return nil, nil
		case ACCEPT_STATUS_NO:
			return nil, nil
		}
	}
}
