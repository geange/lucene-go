package fst

// Builder Builds a minimal FST (maps an IntsRef term to an arbitrary output) from pre-sorted terms with outputs.
// The FST becomes an FSA if you use NoOutputs. The FST is written on-the-fly into a compact serialized format
// byte array, which can be saved to / loaded from a Directory or used directly for traversal.
// The FST is always finite (no cycles).
//
// NOTE: The algorithm is described at http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698
//
// The parameterized type T is the output type. See the subclasses of Outputs.
//
// FSTs larger than 2.1GB are now possible (as of Lucene 4.2). FSTs containing more than 2.1B nodes are also
// now possible, however they cannot be packed.
//
// lucene.experimental
type Builder struct {
	fst       *FST
	NO_OUTPUT any
}
