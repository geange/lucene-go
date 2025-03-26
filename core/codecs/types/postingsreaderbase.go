package types

import (
	"context"
	"io"

	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

type PostingsReaderBase interface {
	io.Closer

	// Init
	// Performs any initialization, such as reading and verifying the header from the provided terms dictionary IndexInput.
	Init(ctx context.Context, termsIn store.IndexInput, state *index.SegmentReadState) error

	// NewTermState
	// Return a newly created empty TermState
	NewTermState() (index.TermState, error)

	// DecodeTerm
	// Actually decode metadata for next term
	DecodeTerm(ctx context.Context, in store.DataInput, fieldInfo *document.FieldInfo, state index.TermState, absolute bool) error

	// Postings
	// Must fully consume state, since after this call that TermState may be reused.
	Postings(ctx context.Context, fieldInfo *document.FieldInfo, state index.TermState, reuse index.PostingsEnum, flags int) (index.PostingsEnum, error)

	// Impacts
	// Return a ImpactsEnum that computes impacts with scorer.
	Impacts(ctx context.Context, fieldInfo *document.FieldInfo, state index.TermState, flags int) (index.ImpactsEnum, error)

	CheckIntegrity() error
}
