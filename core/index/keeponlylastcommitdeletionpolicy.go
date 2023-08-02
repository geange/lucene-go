package index

var _ IndexDeletionPolicy = &KeepOnlyLastCommitDeletionPolicy{}

// KeepOnlyLastCommitDeletionPolicy
// This IndexDeletionPolicy implementation that keeps only the most recent commit and immediately removes all
// prior commits after a new commit is done. This is the default deletion policy.
type KeepOnlyLastCommitDeletionPolicy struct {
}

func NewKeepOnlyLastCommitDeletionPolicy() *KeepOnlyLastCommitDeletionPolicy {
	return &KeepOnlyLastCommitDeletionPolicy{}
}

func (k *KeepOnlyLastCommitDeletionPolicy) OnInit(commits []IndexCommit) error {
	// Note that commits.size() should normally be 1:
	return k.OnCommit(commits)
}

func (k *KeepOnlyLastCommitDeletionPolicy) OnCommit(commits []IndexCommit) error {
	// Note that commits.size() should normally be 2 (if not
	// called by onInit above):
	size := len(commits)
	for i := 0; i < size-1; i++ {
		commit := commits[i]
		if err := commit.Delete(); err != nil {
			return err
		}
	}
	return nil
}
