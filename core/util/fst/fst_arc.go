package fst

type FSTArc struct {
	label           int
	output          any
	target          int64
	flags           byte
	nextFinalOutput any
	nextArc         int64
	nodeFlags       byte

	//*** Fields for arcs belonging to a node with fixed length arcs.
	// So only valid when bytesPerArc != 0.
	// nodeFlags == ARCS_FOR_BINARY_SEARCH || nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	bytesPerArc  int
	posArcsStart int64
	arcIdx       int
	numArcs      int

	// *** Fields for a direct addressing node. nodeFlags == ARCS_FOR_DIRECT_ADDRESSING.

	// Start position in the FST.BytesReader of the presence bits for a direct addressing node, aka the bit-table
	bitTableStart int64

	// First label of a direct addressing node.
	firstLabel int

	// Index of the current label of a direct addressing node. While arcIdx is the current index in the label range, presenceIndex is its corresponding index in the list of actually present labels. It is equal to the number of bits set before the bit at arcIdx in the bit-table. This field is a cache to avoid to count bits set repeatedly when iterating the next arcs.
	presenceIndex int
}
