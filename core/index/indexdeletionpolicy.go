package index

// IndexDeletionPolicy Expert: policy for deletion of stale index commits.
// Implement this interface, and set it on IndexWriterConfig.setIndexDeletionPolicy(IndexDeletionPolicy)
// to customize when older point-in-time commits are deleted from the index directory.
// The default deletion policy is KeepOnlyLastCommitDeletionPolicy, always removes old commits as soon as
// a new commit is done (this matches the behavior before 2.2).
// One expected use case for this (and the reason why it was first created) is to work around problems
// with an index directory accessed via filesystems like NFS because NFS does not provide the "delete on
// last close" semantics that Lucene's "point in time" search normally relies on. By implementing a custom
// deletion policy, such as "a commit is only removed once it has been stale for more than X minutes",
// you can give your readers time to refresh to the new commit before IndexWriter removes the old commits.
// Note that doing so will increase the storage requirements of the index. See LUCENE-710 for details.
type IndexDeletionPolicy interface {

	// OnInit This is called once when a writer is first instantiated to give the policy a chance to remove old commit points.
	// The writer locates all index commits present in the index directory and calls this method. The policy may choose to delete some of the commit points, doing so by calling method delete() of IndexCommit.
	// Note: the last CommitPoint is the most recent one, i.e. the "front index state". Be careful not to delete it, unless you know for sure what you are doing, and unless you can afford to lose the index content while doing that.
	// Params: commits – List of current point-in-time commits, sorted by age (the 0th one is the oldest commit). Note that for a new index this method is invoked with an empty list.
	OnInit(commits []IndexCommit) error

	OnCommit(commits []IndexCommit) error
}
