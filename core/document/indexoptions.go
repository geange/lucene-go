package document

type IndexOptions int

const (
	// INDEX_OPTIONS_NONE
	// Not indexed
	INDEX_OPTIONS_NONE = IndexOptions(iota)

	// INDEX_OPTIONS_DOCS
	// Only documents are indexed: term frequencies and positions are omitted. Phrase
	// and other positional queries on the field will throw an exception, and scoring will behave as if any
	// term in the document appears only once.
	INDEX_OPTIONS_DOCS

	// INDEX_OPTIONS_DOCS_AND_FREQS
	// Only documents and term frequencies are indexed: positions are omitted.
	// This enables normal scoring, except Phrase and other positional queries will throw an exception.
	INDEX_OPTIONS_DOCS_AND_FREQS

	// INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	// Indexes documents, frequencies and positions.
	// This is a typical default for full-text search: full scoring is enabled and positional queries are supported.
	INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS

	// INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	// Indexes documents, frequencies, positions and offsets. Character offsets are encoded alongside the positions.
	INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
)

func (i IndexOptions) String() string {
	switch i {
	case INDEX_OPTIONS_NONE:
		return "NONE"
	case INDEX_OPTIONS_DOCS:
		return "DOCS"
	case INDEX_OPTIONS_DOCS_AND_FREQS:
		return "DOCS_AND_FREQS"
	case INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS:
		return "DOCS_AND_FREQS_AND_POSITIONS"
	case INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS:
		return "DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS"
	default:
		return "NONE"
	}
}

func StringToIndexOptions(value string) IndexOptions {
	switch value {
	case "NONE":
		return INDEX_OPTIONS_NONE
	case "DOCS":
		return INDEX_OPTIONS_DOCS
	case "DOCS_AND_FREQS":
		return INDEX_OPTIONS_DOCS_AND_FREQS
	case "DOCS_AND_FREQS_AND_POSITIONS":
		return INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS
	case "DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS":
		return INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS
	default:
		return INDEX_OPTIONS_NONE
	}
}
