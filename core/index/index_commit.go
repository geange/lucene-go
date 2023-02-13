package index

// IndexCommit Expert: represents a single commit into an index as seen by the IndexDeletionPolicy or IndexReader.
// Changes to the content of an index are made visible only after the writer who made that change commits by writing a new segments file (segments_N). This point in time, when the action of writing of a new segments file to the directory is completed, is an index commit.
// Each index commit point has a unique segments file associated with it. The segments file associated with a later index commit point would have a larger N.
// lucene.experimental
//
// TODO: this is now a poor name, because this class also represents a
// point-in-time view from an NRT reader
type IndexCommit interface {
}
