package core

type IndexOptions int

const (
	// INDEX_OPTIONS_NONE Not indexed
	INDEX_OPTIONS_NONE = IndexOptions(iota)

	// INDEX_OPTIONS_DOCS Only documents are indexed: term frequencies and positions are omitted. Phrase
	// and other positional queries on the field will throw an exception, and scoring will behave as if any
	// term in the document appears only once.
	INDEX_OPTIONS_DOCS

	// INDEX_OPTIONS_DOCS_AND_FREQS Only documents and term frequencies are indexed: positions are omitted.
	// This enables normal scoring, except Phrase and other positional queries will throw an exception.
	INDEX_OPTIONS_DOCS_AND_FREQS

	// INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS Indexes documents, frequencies and positions. This is a
	// typical default for full-text search: full scoring is enabled and positional queries are supported.
	INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS

	// INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS Indexes documents, frequencies, positions and
	// offsets. Character offsets are encoded alongside the positions.
	INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
)
