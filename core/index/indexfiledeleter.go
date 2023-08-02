package index

import (
	"context"
	"fmt"
	"github.com/geange/lucene-go/core/store"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	// VERBOSE_REF_COUNTS Change to true to see details of reference counts when infoStream is enabled
	VERBOSE_REF_COUNTS = false
)

// IndexFileDeleter
// This class keeps track of each SegmentInfos instance that
// is still "live", either because it corresponds to a
// segments_N file in the Directory (a "commit", i.e. a
// committed SegmentInfos) or because it's an in-memory
// SegmentInfos that a writer is actively updating but has
// not yet committed.  This class uses simple reference
// counting to map the live SegmentInfos instances to
// individual files in the Directory.
//
// The same directory file may be referenced by more than
// one IndexCommit, i.e. more than one SegmentInfos.
// Therefore we count how many commits reference each file.
// When all the commits referencing a certain file have been
// deleted, the refcount for that file becomes zero, and the
// file is deleted.
//
// A separate deletion policy interface
// (IndexDeletionPolicy) is consulted on creation (onInit)
// and once per commit (onCommit), to decide when a commit
// should be removed.
//
// It is the business of the IndexDeletionPolicy to choose
// when to delete commit points.  The actual mechanics of
// file deletion, retrying, etc, derived from the deletion
// of commit points is the business of the IndexFileDeleter.
//
// The current default deletion policy is {@link
// KeepOnlyLastCommitDeletionPolicy}, which removes all
// prior commits when a new commit has completed.  This
// matches the behavior before 2.2.
//
// Note that you must hold the write.lock before
// instantiating this class.  It opens segments_N file(s)
// directly with no retry logic.
type IndexFileDeleter struct {
	sync.RWMutex

	// Reference count for all files in the index.
	// Counts how many existing commits reference a file.
	refCounts map[string]*RefCount

	// Holds all commits (segments_N) currently in the index.
	// This will have just 1 commit if you are using the
	// default delete policy (KeepOnlyLastCommitDeletionPolicy).
	// Other policies may leave commit points live for longer
	// in which case this list would be longer than 1:
	commits []IndexCommit

	// Holds files we had incref'd from the previous
	// non-commit checkpoint:
	lastFiles map[string]struct{}

	// Commits that the IndexDeletionPolicy have decided to delete:
	commitsToDelete []*CommitPoint

	// for commit point metadata
	directoryOrig store.Directory

	directory store.Directory

	policy IndexDeletionPolicy

	startingCommitDeleted bool

	lastSegmentInfos *SegmentInfos

	writer *IndexWriter
}

// NewIndexFileDeleter
// Initialize the deleter: find all previous commits in the Directory,
// incref the files they reference, call the policy to let it delete commits. This will remove
// any files not referenced by any of the commits.
// Throws: IOException – if there is a low-level IO error
func NewIndexFileDeleter(ctx context.Context, files []string, directoryOrig, directory store.Directory,
	policy IndexDeletionPolicy, segmentInfos *SegmentInfos, writer *IndexWriter, initialIndexExists,
	isReaderInit bool) (*IndexFileDeleter, error) {

	fd := &IndexFileDeleter{
		refCounts: map[string]*RefCount{},
	}
	fd.writer = writer

	currentSegmentsFile := segmentInfos.GetSegmentsFileName()

	fd.policy = policy
	fd.directoryOrig = directoryOrig
	fd.directory = directory

	// First pass: walk the files and initialize our ref
	// counts:
	var currentCommitPoint *CommitPoint

	if currentSegmentsFile != "" {
		for _, fileName := range files {
			if !CODEC_FILE_PATTERN.MatchString(fileName) {
				continue
			}

			if strings.HasSuffix(fileName, "write.lock") &&
				(CODEC_FILE_PATTERN.MatchString(fileName) ||
					strings.HasPrefix(fileName, SEGMENTS) ||
					strings.HasPrefix(fileName, PENDING_SEGMENTS)) {

				// Add this file to refCounts with initial count 0:
				fd.getRefCount(fileName)

				if strings.HasPrefix(fileName, SEGMENTS) && fileName != OLD_SEGMENTS_GEN {
					// This is a commit (segments or segments_N), and
					// it's valid (<= the max gen).  Load it, then
					// incref all files it refers to:

					sis, err := ReadCommit(ctx, directoryOrig, fileName)
					if err != nil {
						return nil, err
					}

					commitPoint, err := NewCommitPoint(&fd.commitsToDelete, directoryOrig, sis)
					if err != nil {
						return nil, err
					}

					if sis.GetGeneration() == segmentInfos.GetGeneration() {
						currentCommitPoint = commitPoint
					}

					fd.commits = append(fd.commits, commitPoint)
					if err := fd.IncRef(sis, true); err != nil {
						return nil, err
					}

					if fd.lastSegmentInfos == nil ||
						sis.GetGeneration() > fd.lastSegmentInfos.GetGeneration() {
						fd.lastSegmentInfos = sis
					}
				}
			}
		}
	}

	if currentCommitPoint == nil && currentSegmentsFile != "" && initialIndexExists {
		// We did not in fact see the segments_N file
		// corresponding to the segmentInfos that was passed
		// in.  Yet, it must exist, because our caller holds
		// the write lock.  This can happen when the directory
		// listing was stale (eg when index accessed via NFS
		// client with stale directory listing cache).  So we
		// try now to explicitly open this commit point:
		sis, err := ReadCommit(ctx, directoryOrig, currentSegmentsFile)
		if err != nil {
			return nil, err
		}
		currentCommitPoint, err = NewCommitPoint(&fd.commitsToDelete, directoryOrig, sis)
		if err != nil {
			return nil, err
		}
		fd.commits = append(fd.commits, currentCommitPoint)
		if err := fd.IncRef(sis, true); err != nil {
			return nil, err
		}
	}

	if isReaderInit {
		// Incoming SegmentInfos may have NRT changes not yet visible in the latest commit, so we have to protect its files from deletion too:
		if err := fd.Checkpoint(segmentInfos, false); err != nil {
			return nil, err
		}
	}

	// We keep commits list in sorted order (oldest to newest):
	sort.Sort(IndexCommits(fd.commits))
	relevantFiles := make(map[string]struct{})
	for k := range fd.refCounts {
		relevantFiles[k] = struct{}{}
	}
	//pendingDeletions, err := directoryOrig.GetPendingDeletions()
	//if err != nil {
	//	return nil, err
	//}
	//for k := range pendingDeletions {
	//	relevantFiles[k] = struct{}{}
	//}

	// refCounts only includes "normal" filenames (does not include write.lock)
	inflateGens(segmentInfos, relevantFiles)

	// Now delete anything with ref count at 0.  These are
	// presumably abandoned files eg due to crash of
	// IndexWriter.

	toDelete := make(map[string]struct{})
	for fileName, rc := range fd.refCounts {
		if rc.count == 0 {
			if strings.HasPrefix(fileName, SEGMENTS) {
				return nil, fmt.Errorf("file '%s' has refCount=0, which should never happen on init", fileName)
			}
			toDelete[fileName] = struct{}{}
		}
	}

	if err := fd.deleteFiles(toDelete); err != nil {
		return nil, err
	}

	// Finally, give policy a chance to remove things on
	// startup:
	if err := policy.OnInit(fd.commits); err != nil {
		return nil, err
	}

	// Always protect the incoming segmentInfos since
	// sometime it may not be the most recent commit
	if err := fd.Checkpoint(segmentInfos, false); err != nil {
		return nil, err
	}

	if currentCommitPoint == nil {
		fd.startingCommitDeleted = false
	} else {
		fd.startingCommitDeleted = currentCommitPoint.IsDeleted()
	}

	if err := fd.deleteCommits(); err != nil {
		return nil, err
	}

	return fd, nil
}

// Set all gens beyond what we currently see in the directory, to avoid double-write in cases where
// the previous IndexWriter did not gracefully close/rollback (e.g. os/machine crashed or lost power).
// 将所有gens设置为超出我们当前在目录中看到的值，
// 以避免在上一个IndexWriter未正常关闭/回滚（例如，os/machine崩溃或断电）的情况下进行双重写入。
func inflateGens(infos *SegmentInfos, files map[string]struct{}) {
	maxSegmentGen := math.MinInt64
	maxSegmentName := math.MinInt64

	// Confusingly, this is the union of liveDocs, field infos, doc values
	// (and maybe others, in the future) gens.  This is somewhat messy,
	// since it means DV updates will suddenly write to the next gen after
	// live docs' gen, for example, but we don't have the APIs to ask the
	// codec which file is which:
	// 令人困惑的是，这是liveDocs、字段信息、doc值（以及未来可能的其他值）氏族的集合。
	// 这有些混乱，因为这意味着DV更新会在live docs的gen之后突然写入下一代，
	// 例如，但我们没有API来询问某个文件的编解码器是什么
	maxPerSegmentGen := make(map[string]int64)

	for fileName := range files {
		if fileName == OLD_SEGMENTS_GEN || fileName == WRITE_LOCK_NAME {
			// do nothing
		} else if strings.HasPrefix(fileName, SEGMENTS) {
			num, err := GenerationFromSegmentsFileName(fileName)
			if err != nil {
				return
			}
			maxSegmentGen = int(max(num, int64(maxSegmentGen)))
		} else if strings.HasPrefix(fileName, PENDING_SEGMENTS) {
			num, err := GenerationFromSegmentsFileName(fileName[8:])
			if err != nil {
				return
			}
			maxSegmentGen = int(max(num, int64(maxSegmentGen)))
		} else {
			segmentName := ParseSegmentName(fileName)

			if strings.HasSuffix(strings.ToLower(fileName), ".tmp") {
				continue
			}

			parseInt, err := strconv.ParseInt(segmentName[1:], 36, 64)
			if err != nil {
				return
			}

			maxSegmentName = max(maxSegmentGen, int(parseInt))

			curGen, ok := maxPerSegmentGen[segmentName]
			if !ok {
				curGen = 0
			}
			generation := ParseGeneration(fileName)
			curGen = max(curGen, generation)
			maxPerSegmentGen[segmentName] = curGen
		}
	}

	// Generation is advanced before write:
	infos.SetNextWriteGeneration(max(infos.GetGeneration(), int64(maxSegmentGen)))
	value := int64(1 + maxSegmentName)
	if infos.counter < value {
		infos.counter = value
	}

	for _, info := range infos.segments {
		gen := maxPerSegmentGen[info.Info().Name()]
		genLong := gen
		if info.GetNextWriteDelGen() < genLong+1 {
			info.SetNextWriteDelGen(genLong + 1)
		}

		if info.GetNextWriteFieldInfosGen() < genLong+1 {
			info.SetNextWriteFieldInfosGen(genLong + 1)
		}

		if info.GetNextWriteDocValuesGen() < genLong+1 {
			info.SetNextWriteDocValuesGen(genLong + 1)
		}
	}

}

func (r *IndexFileDeleter) getRefCount(fileName string) *RefCount {
	if rc, ok := r.refCounts[fileName]; ok {
		return rc
	}

	rc := NewRefCount(fileName)
	r.refCounts[fileName] = rc
	return rc
}

func (r *IndexFileDeleter) IncRef(segmentInfos *SegmentInfos, isCommit bool) error {

	files, err := segmentInfos.Files(isCommit)
	if err != nil {
		return err
	}

	for fileName := range files {
		err := r.incRefFileName(fileName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *IndexFileDeleter) IncRefFiles(files map[string]struct{}) error {
	for file := range files {
		err := r.incRefFileName(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *IndexFileDeleter) incRefFileName(fileName string) error {
	_, err := r.getRefCount(fileName).IncRef()
	return err
}

// Checkpoint For definition of "check point" see IndexWriter comments: "Clarification:
// Check Points (and commits)". IndexWriter calls this when it has made a "consistent change" to the index,
// meaning new files are written to the index and the in-memory SegmentInfos have been modified to
// point to those files. This may or may not be a commit (segments_N may or may not have been written).
// We simply incref the files referenced by the new SegmentInfos and decref the files we had previously
// seen (if any). If this is a commit, we also call the policy to give it a chance to remove other commits.
// If any commits are removed, we decref their files as well.
func (r *IndexFileDeleter) Checkpoint(segmentInfos *SegmentInfos, isCommit bool) error {
	err := r.IncRef(segmentInfos, isCommit)
	if err != nil {
		return err
	}

	if isCommit {
		// Append to our commits list:
		commitPoint, err := NewCommitPoint(&r.commitsToDelete, r.directoryOrig, segmentInfos)
		if err != nil {
			return err
		}
		r.commits = append(r.commits, commitPoint)

		// Tell policy so it can remove commits:
		if err := r.policy.OnCommit(r.commits); err != nil {
			return err
		}

		// Decref files for commits that were deleted by the policy:
		return r.deleteCommits()
	}

	if err := r.DecRef(r.lastFiles); err != nil {
		return err
	}
	r.lastFiles = map[string]struct{}{}

	files, err := segmentInfos.Files(false)
	if err != nil {
		return err
	}

	// Save files so we can decr on next checkpoint/commit:
	for fileName := range files {
		r.lastFiles[fileName] = struct{}{}
	}
	return nil
}

// Remove the IndexCommits in the commitsToDelete List by DecRef'ing all files from each SegmentInfos.
func (r *IndexFileDeleter) deleteCommits() error {
	for _, commit := range r.commitsToDelete {
		if err := r.DecRef(commit.files); err != nil {
			return err
		}
	}

	r.commitsToDelete = r.commitsToDelete[:0]

	// Now compact commits to remove deleted ones (preserving the sort):
	size := len(r.commits)

	writeTo := 0
	for readFrom := 0; readFrom < len(r.commits); readFrom++ {
		commit := r.commits[readFrom]
		if !commit.IsDeleted() {
			if writeTo != readFrom {
				r.commits[writeTo] = commit
			}
			writeTo++
		}
	}

	for size > writeTo {
		r.commits = r.commits[:size-1]
		size--
	}

	return nil
}

func (r *IndexFileDeleter) DecRef(files map[string]struct{}) error {
	toDelete := make(map[string]struct{})
	for file := range files {
		if r.decRef(file) {
			toDelete[file] = struct{}{}
		}
	}
	return r.deleteFiles(toDelete)
}

// Returns true if the file should now be deleted.
func (r *IndexFileDeleter) decRef(fileName string) bool {
	rc := r.getRefCount(fileName)

	if rc.DecRef() == 0 {
		// This file is no longer referenced by any past
		// commit points nor by the in-memory SegmentInfos:
		delete(r.refCounts, fileName)
		return true
	}
	return false
}

func (r *IndexFileDeleter) deleteFiles(names map[string]struct{}) error {
	for name := range names {
		if !strings.HasPrefix(name, SEGMENTS) {
			continue
		}

		if err := r.deleteFile(name); err != nil {
			return err
		}
	}

	for name := range names {
		if strings.HasPrefix(name, SEGMENTS) {
			continue
		}

		if err := r.deleteFile(name); err != nil {
			return err
		}
	}

	return nil
}

func (r *IndexFileDeleter) deleteFile(name string) error {
	return r.directory.DeleteFile(nil, name)
}

func (r *IndexFileDeleter) deleteNewFiles(files map[string]struct{}) error {
	toDelete := make(map[string]struct{})

	for fileName := range files {
		// NOTE: it's very unusual yet possible for the
		// refCount to be present and 0: it can happen if you
		// open IW on a crashed index, and it removes a bunch
		// of unref'd files, and then you add new docs / do
		// merging, and it reuses that segment name.
		// TestCrash.testCrashAfterReopen can hit this:
		if _, ok := r.refCounts[fileName]; !ok {
			toDelete[fileName] = struct{}{}
		}
	}
	return r.deleteFiles(toDelete)
}

type RefCount struct {
	fileName string
	initDone bool
	count    int
}

func (r *RefCount) IncRef() (int, error) {
	if !r.initDone {
		r.initDone = true
	}
	r.count++
	return r.count, nil
}

func (r *RefCount) DecRef() int {
	r.count--
	return r.count
}

func NewRefCount(fileName string) *RefCount {
	return &RefCount{fileName: fileName}
}

var _ sort.Interface = IndexCommits{}

type IndexCommits []IndexCommit

func (list IndexCommits) Len() int {
	return len(list)
}

func (list IndexCommits) Less(i, j int) bool {
	return list[i].CompareTo(list[j]) < 0
}

func (list IndexCommits) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

var _ IndexCommit = &CommitPoint{}

// CommitPoint Holds details for each commit point. This class is also passed to the deletion policy.
// Note: this class has a natural ordering that is inconsistent with equals.
type CommitPoint struct {
	files            map[string]struct{}
	segmentsFileName string
	deleted          bool
	directoryOrig    store.Directory
	commitsToDelete  *[]*CommitPoint
	generation       int64
	userData         map[string]string
	segmentCount     int
}

func NewCommitPoint(commitsToDelete *[]*CommitPoint,
	directoryOrig store.Directory, segmentInfos *SegmentInfos) (*CommitPoint, error) {

	files, err := segmentInfos.Files(true)
	if err != nil {
		return nil, err
	}

	return &CommitPoint{
		files:            files,
		segmentsFileName: segmentInfos.GetSegmentsFileName(),
		deleted:          false,
		directoryOrig:    directoryOrig,
		commitsToDelete:  commitsToDelete,
		generation:       segmentInfos.GetGeneration(),
		userData:         segmentInfos.GetUserData(),
		segmentCount:     segmentInfos.Size(),
	}, nil
}

func (c *CommitPoint) GetSegmentsFileName() string {
	return c.segmentsFileName
}

func (c *CommitPoint) GetFileNames() (map[string]struct{}, error) {
	return c.files, nil
}

func (c *CommitPoint) GetDirectory() store.Directory {
	return c.directoryOrig
}

// Delete Called only be the deletion policy, to remove this commit point from the index.
func (c *CommitPoint) Delete() error {
	if !c.deleted {
		c.deleted = true
		*c.commitsToDelete = append(*c.commitsToDelete, c)
	}
	return nil
}

func (c *CommitPoint) IsDeleted() bool {
	return c.deleted
}

func (c *CommitPoint) GetSegmentCount() int {
	return c.segmentCount
}

func (c *CommitPoint) GetGeneration() int64 {
	return c.generation
}

func (c *CommitPoint) GetUserData() (map[string]string, error) {
	return c.userData, nil
}

func (c *CommitPoint) CompareTo(commit IndexCommit) int {
	gen := c.GetGeneration()
	comgen := commit.GetGeneration()
	return Compare(gen, comgen)
}

func (c *CommitPoint) GetReader() *StandardDirectoryReader {
	return nil
}
