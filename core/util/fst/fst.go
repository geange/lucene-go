package fst

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/geange/lucene-go/codecs/utils"
	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/array"
)

type FST[T any] struct {
	inputType InputType

	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput  T
	onceSetEmpty sync.Once

	//hasEmptyOutput bool

	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB).
	// If the FST is less than 1 GB then bytesArray is set instead.
	bytes *ByteStore

	fstStore  Store
	startNode int64
	outputs   Outputs[T]
	//manager   OutputManager[T]
}

func NewFST[T any](inputType InputType, outputs Outputs[T], bytesPageBits int) *FST[T] {
	return &FST[T]{
		inputType: inputType,
		//emptyOutput:    T{},
		//hasEmptyOutput: false,
		bytes:     NewByteStore(bytesPageBits),
		fstStore:  nil,
		startNode: -1,
		outputs:   outputs,
	}
}

// NewFstV1 Load a previously saved FST.
func NewFstV1[T any](ctx context.Context, outputs Outputs[T], metaIn, in store.DataInput) (*FST[T], error) {
	heapStore, err := NewOnHeapStore(DEFAULT_MAX_BLOCK_BITS)
	if err != nil {
		return nil, err
	}
	return NewFstV2(ctx, outputs, heapStore, metaIn, in)
}

// NewFstV2
// Load a previously saved FST; maxBlockBits allows you to control the size of
// the byte[] pages used to hold the FST bytes.
func NewFstV2[T any](ctx context.Context, outputs Outputs[T], fstStore Store, metaIn, in store.DataInput) (*FST[T], error) {
	fst := &FST[T]{
		bytes:    nil,
		fstStore: fstStore,
		outputs:  outputs,
	}

	// NOTE: only reads formats VERSION_START up to VERSION_CURRENT; we don't have
	// back-compat promise for FSTs (they are experimental), but we are sometimes able to offer it
	if _, err := codecs.CheckHeader(ctx, metaIn, FILE_FORMAT_NAME, VERSION_START, VERSION_CURRENT); err != nil {
		return nil, err
	}

	isAcceptsEmpty, err := metaIn.ReadByte()
	if err != nil {
		return nil, err
	}

	if isAcceptsEmpty == 1 {
		// accepts empty string
		// 1 KB blocks:
		emptyBytes := NewByteStore(10)
		numBytes, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}

		if err := emptyBytes.CopyBytes(context.Background(), metaIn, int(numBytes)); err != nil {
			return nil, err
		}

		// De-serialize empty-string output:
		reader, err := emptyBytes.GetReverseReader()
		if err != nil {
			return nil, err
		}

		// NoOutputs uses 0 bytes when writing its output,
		// so we have to check here else BytesStore gets
		// angry:
		if numBytes > 0 {
			if err := reader.SetPosition(int64(numBytes - 1)); err != nil {
				return nil, err
			}
		}

		output, err := outputs.ReadFinalOutput(ctx, reader)
		if err != nil {
			return nil, err
		}
		fst.emptyOutput = output
	} else {
		fst.emptyOutput = *new(T)
	}
	t, err := metaIn.ReadByte()
	if err != nil {
		return nil, err
	}
	switch t {
	case 0:
		fst.inputType = BYTE1
		break
	case 1:
		fst.inputType = BYTE2
		break
	case 2:
		fst.inputType = BYTE4
		break
	default:
		return nil, fmt.Errorf("invalid input type %d", in)
	}
	startNode, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	fst.startNode = int64(startNode)

	numBytes, err := metaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}

	if err = fst.fstStore.Init(in, int64(numBytes)); err != nil {
		return nil, err
	}

	return fst, nil
}

func (f *FST[T]) SetEmptyOutput(output T) error {
	var err error
	var emptyOutput T
	f.onceSetEmpty.Do(func() {
		emptyOutput, err = f.outputs.Merge(f.emptyOutput, output)
		if err == nil {
			f.emptyOutput = emptyOutput
		}
	})
	return err
}

func (f *FST[T]) Save(ctx context.Context, metaOut store.DataOutput, out store.DataOutput) error {
	if f.startNode == -1 {
		return errors.New("call finish first")
	}

	if err := utils.WriteHeader(ctx, metaOut, FILE_FORMAT_NAME, VERSION_CURRENT); err != nil {
		return err
	}

	// TODO: really we should encode this as an arc, arriving
	// to the root node, instead of special casing here:
	if !f.outputs.IsNoOutput(f.emptyOutput) {
		// Accepts empty string
		if err := metaOut.WriteByte(1); err != nil {
			return err
		}

		// Serialize empty-string output:
		ros := store.NewBufferDataOutput()
		if err := f.outputs.WriteFinalOutput(ctx, f.emptyOutput, ros); err != nil {
			return err
		}

		pointer := ros.GetFilePointer()
		emptyOutputBytes := make([]byte, pointer)
		copy(emptyOutputBytes, ros.Bytes())
		emptyLen := len(emptyOutputBytes)

		// reverse
		stopAt := emptyLen / 2
		upto := 0
		for upto < stopAt {
			b := emptyOutputBytes[upto]
			emptyOutputBytes[upto] = emptyOutputBytes[emptyLen-upto-1]
			emptyOutputBytes[emptyLen-upto-1] = b
			upto++
		}
		if err := metaOut.WriteUvarint(ctx, uint64(emptyLen)); err != nil {
			return err
		}

		if _, err := metaOut.Write(emptyOutputBytes); err != nil {
			return err
		}

	} else {
		if err := metaOut.WriteByte(0); err != nil {
			return err
		}
	}

	t := byte(0)
	switch f.inputType {
	case BYTE1:
		t = 0
	case BYTE2:
		t = 1
	default:
		t = 2
	}

	if err := metaOut.WriteByte(t); err != nil {
		return err
	}

	if err := metaOut.WriteUvarint(ctx, uint64(f.startNode)); err != nil {
		return err
	}

	if f.bytes != nil {
		numBytes := f.bytes.GetPosition()
		if err := metaOut.WriteUvarint(ctx, uint64(numBytes)); err != nil {
			return err
		}
		return f.bytes.WriteToDataOutput(out)
	}

	return f.fstStore.WriteTo(ctx, out)
}

func (f *FST[T]) SaveToFile(ctx context.Context, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := store.NewOutputStream("", file)
	return f.Save(ctx, out, out)
}

// NewFSTFromFile Reads an automaton from a file.
func NewFSTFromFile[T any](ctx context.Context, path string, outputs Outputs[T]) (*FST[T], error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	in := store.NewInputStream(file)
	fstStore, err := NewOnHeapStore(DEFAULT_MAX_BLOCK_BITS)
	if err != nil {
		return nil, err
	}
	return NewFstV2(ctx, outputs, fstStore, in, in)
}

func (f *FST[T]) writeLabel(ctx context.Context, out store.DataOutput, v int) error {
	switch f.inputType {
	case BYTE1:
		return out.WriteByte(byte(v))
	case BYTE2:
		return out.WriteUint16(ctx, uint16(v))
	default:
		return out.WriteUvarint(ctx, uint64(v))
	}
}

// ReadLabel Reads one BYTE1/2/4 label from the provided DataInput.
func (f *FST[T]) ReadLabel(ctx context.Context, in store.DataInput) (int, error) {
	var v int
	switch f.inputType {
	case BYTE1:
		n, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		v = int(n)
		return v, nil
	case BYTE2:
		n, err := in.ReadUint16(ctx)
		if err != nil {
			return 0, err
		}
		v = int(n & 0xFFFF)
		return v, nil
	default:
		n, err := in.ReadUvarint(ctx)
		if err != nil {
			return 0, err
		}
		return int(n), nil
	}
}

// TargetHasArcs returns true if the node at this address has any outgoing arcs
func TargetHasArcs[T any](arc *Arc[T]) bool {
	return arc.Target() > 0
}

// AddNode
// serializes new node by appending its bytes to the end
// of the current byte[]
func (f *FST[T]) AddNode(ctx context.Context, builder *Builder[T], nodeIn *UnCompiledNode[T]) (int64, error) {

	if nodeIn.NumArcs() == 0 {
		if nodeIn.IsFinal {
			return FINAL_END_NODE, nil
		}
		return NON_FINAL_END_NODE, nil
	}
	startAddress := builder.bytes.GetPosition()

	doFixedLengthArcs := shouldExpandNodeWithFixedLengthArcs(builder, nodeIn)
	if doFixedLengthArcs {
		if len(builder.numBytesPerArc) < nodeIn.NumArcs() {
			size := array.Oversize(nodeIn.NumArcs(), INTEGER_BYTES)
			builder.numBytesPerArc = make([]int, size)
			builder.numLabelBytesPerArc = make([]int, size)
		}
	}

	builder.arcCount += nodeIn.NumArcs()

	lastArc := nodeIn.NumArcs() - 1

	lastArcStart := builder.bytes.GetPosition()
	maxBytesPerArc := 0
	maxBytesPerArcWithoutLabel := 0

	numArcs := nodeIn.NumArcs()
	for arcIdx := 0; arcIdx < numArcs; arcIdx++ {
		arc := nodeIn.Arcs[arcIdx]
		target, ok := arc.Target.(*CompiledNode)
		if !ok {
			return 0, errors.New("arc.Target is not *CompiledNode")
		}
		flags := 0

		if arcIdx == lastArc {
			flags += BIT_LAST_ARC
		}

		if builder.lastFrozenNode == target.node && !doFixedLengthArcs {
			// TODO: for better perf (but more RAM used) we
			// could avoid this except when arc is "near" the
			// last arc:
			flags += BIT_TARGET_NEXT
		}

		if arc.IsFinal {
			flags += BIT_FINAL_ARC
			if !f.outputs.IsNoOutput(arc.NextFinalOutput) {
				flags += BIT_ARC_HAS_FINAL_OUTPUT
			}
		}

		targetHasArcs := target.node > 0

		if !targetHasArcs {
			flags += BIT_STOP_NODE
		}

		if !f.outputs.IsNoOutput(arc.Output) {
			flags += BIT_ARC_HAS_OUTPUT
		}

		if err := builder.bytes.WriteByte(byte(flags)); err != nil {
			return 0, err
		}

		labelStart := builder.bytes.GetPosition()
		if err := f.writeLabel(ctx, builder.bytes, arc.Label); err != nil {
			return 0, err
		}

		numLabelBytes := builder.bytes.GetPosition() - labelStart

		if !f.outputs.IsNoOutput(arc.Output) {
			if err := f.outputs.Write(ctx, arc.Output, builder.bytes); err != nil {
				return 0, err
			}
		}

		if !f.outputs.IsNoOutput(arc.NextFinalOutput) {
			if err := f.outputs.WriteFinalOutput(ctx, arc.NextFinalOutput, builder.bytes); err != nil {
				return 0, err
			}
		}

		if targetHasArcs && (flags&BIT_TARGET_NEXT) == 0 {
			if err := builder.bytes.WriteUvarint(ctx, uint64(target.node)); err != nil {
				return 0, err
			}
		}

		// just write the arcs "like normal" on first pass, but record how many bytes each one took
		// and max byte size:
		// 只需在第一次传递时“像平常一样”编写弧，但记录每个弧占用了多少字节以及最大字节大小
		if doFixedLengthArcs {
			numArcBytes := int(builder.bytes.GetPosition() - lastArcStart)
			builder.numBytesPerArc[arcIdx] = numArcBytes
			builder.numLabelBytesPerArc[arcIdx] = int(numLabelBytes)
			lastArcStart = builder.bytes.GetPosition()
			maxBytesPerArc = max(maxBytesPerArc, numArcBytes)
			maxBytesPerArcWithoutLabel = max(maxBytesPerArcWithoutLabel, numArcBytes-int(numLabelBytes))
		}
	}

	if doFixedLengthArcs {
		// 2nd pass just "expands" all arcs to take up a fixed byte size
		labelRange := nodeIn.lastArc().Label - nodeIn.Arcs[0].Label + 1

		ok := shouldExpandNodeWithDirectAddressing(builder, nodeIn, int64(maxBytesPerArc),
			int64(maxBytesPerArcWithoutLabel), labelRange)
		if ok {
			if err := writeNodeForDirectAddressing(ctx, builder, nodeIn,
				startAddress, maxBytesPerArcWithoutLabel, labelRange); err != nil {
				return 0, err
			}

			builder.directAddressingNodeCount++
		} else {
			if err := writeNodeForBinarySearch(ctx, builder, nodeIn,
				startAddress, int64(maxBytesPerArc)); err != nil {
				return 0, err
			}

			builder.binarySearchNodeCount++
		}
	}

	nodeAddress := builder.bytes.GetPosition() - 1
	if err := builder.bytes.Reverse(startAddress, nodeAddress); err != nil {
		return 0, err
	}

	builder.nodeCount++
	return nodeAddress, nil
}

// GetFirstArc Fills virtual 'start' arc, ie, an empty incoming arc to the FST's start node
func (f *FST[T]) GetFirstArc(arc *Arc[T]) (*Arc[T], error) {
	NoOutput := f.outputs.GetNoOutput()

	if !f.outputs.IsNoOutput(f.emptyOutput) {
		arc.flags = BIT_FINAL_ARC | BIT_LAST_ARC
		arc.nextFinalOutput = f.emptyOutput
		if !f.outputs.IsNoOutput(f.emptyOutput) {
			arc.flags = arc.Flags() | BIT_ARC_HAS_FINAL_OUTPUT
		}
	} else {
		arc.flags = BIT_LAST_ARC
		arc.nextFinalOutput = NoOutput
	}
	arc.output = NoOutput

	// If there are no nodes, ie, the FST only accepts the
	// empty string, then startNode is 0
	arc.target = f.startNode
	return arc, nil
}

// ReadFirstTargetArc
// Follow the follow arc and read the first arc of its target; this changes
// the provided arc (3rd arg) in-place and returns it.
// Returns: Returns the second argument (arc).
func (f *FST[T]) ReadFirstTargetArc(ctx context.Context, in BytesReader, follow *Arc[T], arc *Arc[T]) (*Arc[T], error) {
	if !follow.IsFinal() {
		return f.ReadFirstRealTargetArc(ctx, follow.Target(), in, arc)
	}

	// Insert "fake" final first arc:
	arc.label = END_LABEL
	arc.output = follow.NextFinalOutput()
	arc.flags = BIT_FINAL_ARC
	if follow.Target() <= 0 {
		arc.flags |= BIT_LAST_ARC
	} else {
		// NOTE: nextArc is a node (not an address!) in this case:
		arc.nextArc = follow.Target()
	}
	arc.target = FINAL_END_NODE
	arc.nodeFlags = arc.flags
	return arc, nil
}

func (f *FST[T]) ReadFirstRealTargetArc(ctx context.Context, nodeAddress int64, in BytesReader, arc *Arc[T]) (*Arc[T], error) {
	if err := in.SetPosition(nodeAddress); err != nil {
		return nil, err
	}

	flags, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	arc.nodeFlags = flags

	if flags == ArcsForBinarySearch || flags == ArcsForDirectAddressing {
		numArcs, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		arc.numArcs = int(numArcs)

		bytesPerArc, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc = int(bytesPerArc)

		arc.arcIdx = -1

		if flags == ArcsForDirectAddressing {
			if err := readPresenceBytes(ctx, in, arc); err != nil {
				return nil, err
			}

			label, err := f.ReadLabel(ctx, in)
			if err != nil {
				return nil, err
			}

			arc.firstLabel = label
			arc.presenceIndex = -1
		}

		arc.posArcsStart = in.GetPosition()
	} else {
		arc.nextArc = nodeAddress
		arc.bytesPerArc = 0
	}

	return f.ReadNextRealArc(ctx, in, arc)
}

// ReadNextArc In-place read; returns the arc.
func (f *FST[T]) ReadNextArc(ctx context.Context, arc *Arc[T], in BytesReader) (*Arc[T], error) {
	if arc.Label() == END_LABEL {
		// This was a fake inserted "final" arc
		if arc.NextArc() <= 0 {
			return nil, errors.New("cannot readNextArc when arc.isLast()=true")
		}
		return f.ReadFirstRealTargetArc(ctx, arc.NextArc(), in, arc)
	}
	return f.ReadNextRealArc(ctx, in, arc)
}

// Peeks at next arc's label; does not alter arc. Do not call this if arc.isLast()!
func (f *FST[T]) readNextArcLabel(ctx context.Context, arc *Arc[T], in BytesReader) (int, error) {
	if arc.Label() == END_LABEL {
		// Next arc is the first arc of a node.
		// Position to read the first arc label.
		if err := in.SetPosition(arc.NextArc()); err != nil {
			return 0, err
		}
		flags, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		switch flags {
		case ArcsForBinarySearch, ArcsForDirectAddressing:
			// Special arc which is actually a node header for fixed length arcs.
			numArcs, err := in.ReadUvarint(ctx)
			if err != nil {
				return 0, err
			}
			if _, err := in.ReadUvarint(ctx); err != nil {
				return 0, err
			} // Skip bytesPerArc.
			if flags == ArcsForBinarySearch {
				if _, err := in.ReadByte(); err != nil {
					return 0, err
				} // Skip arc flags.
			} else {
				numBytes := getNumPresenceBytes(int(numArcs))
				if err := in.SkipBytes(ctx, numBytes); err != nil {
					return 0, err
				}
			}
		}
		return f.ReadLabel(ctx, in)
	}

	if arc.BytesPerArc() != 0 {
		// Arcs have fixed length.
		if arc.NodeFlags() == ArcsForBinarySearch {
			// Point to next arc, -1 to skip arc flags.
			pos := arc.PosArcsStart() - (1+int64(arc.ArcIdx()))*int64(arc.BytesPerArc()) - 1
			if err := in.SetPosition(pos); err != nil {
				return 0, err
			}
		} else {
			// Direct addressing node. The label is not stored but rather inferred
			// based on first label and arc index in the range.
			nextIndex, err := NextBitSet(ctx, arc.ArcIdx(), arc, in)
			if err != nil {
				return 0, err
			}
			return arc.FirstLabel() + nextIndex, nil
		}
	} else {
		// Arcs have variable length.
		// Position to next arc, -1 to skip flags.
		if err := in.SetPosition(arc.NextArc() - 1); err != nil {
			return 0, err
		}
	}
	return f.ReadLabel(ctx, in)
}

func (f *FST[T]) ReadArcByIndex(ctx context.Context, in BytesReader, idx int, arc *Arc[T]) (*Arc[T], error) {
	if err := in.SetPosition(arc.PosArcsStart() - int64(idx*arc.BytesPerArc())); err != nil {
		return nil, err
	}
	arc.arcIdx = idx

	flags, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	arc.flags = flags

	return f.readArc(ctx, in, arc)
}

// ReadArcByDirectAddressing Reads a present direct addressing node arc, with the provided index in the label range.
// rangeIndex: The index of the arc in the label range. It must be present.
// The real arc offset is computed based on the presence bits of the direct addressing node.
func (f *FST[T]) ReadArcByDirectAddressing(ctx context.Context, in BytesReader, rangeIndex int, arc *Arc[T]) (*Arc[T], error) {
	presenceIndex, err := CountBitsUpTo(rangeIndex, arc, in)
	if err != nil {
		return nil, err
	}
	return f.readArcByDirectAddressing(ctx, arc, in, rangeIndex, presenceIndex)
}

// ReadLastArcByDirectAddressing Reads the last arc of a direct addressing node.
// This method is equivalent to call readArcByDirectAddressing(Fst.Arc, Fst.BytesReader, int)
// with rangeIndex equal to arc.numArcs() - 1, but it is faster.
func (f *FST[T]) ReadLastArcByDirectAddressing(ctx context.Context, arc *Arc[T], in BytesReader) (*Arc[T], error) {
	presenceIndex, err := CountBits(arc, in)
	if err != nil {
		return nil, err
	}

	presenceIndex -= 1
	return f.readArcByDirectAddressing(ctx, arc, in, arc.NumArcs()-1, presenceIndex)
}

// ReadNextRealArc Never returns null, but you should never call this if arc.isLast() is true.
func (f *FST[T]) ReadNextRealArc(ctx context.Context, in BytesReader, arc *Arc[T]) (*Arc[T], error) {
	switch arc.NodeFlags() {
	case ArcsForBinarySearch:
		arc.arcIdx++
		if err := in.SetPosition(arc.PosArcsStart() - int64(arc.ArcIdx()*arc.BytesPerArc())); err != nil {
			return nil, err
		}

		flags, err := in.ReadByte()
		if err != nil {
			return nil, err
		}
		arc.flags = flags

	case ArcsForDirectAddressing:
		nextIndex, err := NextBitSet(ctx, arc.ArcIdx(), arc, in)
		if err != nil {
			return nil, err
		}
		return f.readArcByDirectAddressing(ctx, arc, in, nextIndex, arc.presenceIndex+1)

	default:
		if arc.BytesPerArc() != 0 {
			return nil, fmt.Errorf("arc.BytesPerArc() != 0; arc.BytesPerArc() is %d", arc.BytesPerArc())
		}

		if err := in.SetPosition(arc.NextArc()); err != nil {
			return nil, err
		}

		flags, err := in.ReadByte()
		if err != nil {
			return nil, err
		}
		arc.flags = flags
	}
	return f.readArc(ctx, in, arc)
}

// FindTargetArc Finds an arc leaving the incoming arc, replacing the arc in place.
// This returns null if the arc was not found, else the incoming arc.
// 查找follow后满足label=${labelToMatch}的Arc
func (f *FST[T]) FindTargetArc(ctx context.Context, labelToMatch int, in BytesReader, follow, arc *Arc[T]) (*Arc[T], bool, error) {

	if labelToMatch == END_LABEL {
		if follow.IsFinal() {
			if follow.Target() <= 0 {
				arc.flags = BIT_LAST_ARC
			} else {
				arc.flags = 0
				// NOTE: nextArc is a node (not an address!) in this case:
				arc.nextArc = follow.Target()
			}
			arc.output = follow.NextFinalOutput()
			arc.label = END_LABEL
			arc.nodeFlags = arc.flags
			return arc, true, nil
		} else {
			return nil, false, nil
		}
	}

	if !TargetHasArcs(follow) {
		return nil, false, nil
	}

	if err := in.SetPosition(follow.Target()); err != nil {
		return nil, false, err
	}

	flags, err := in.ReadByte()
	if err != nil {
		return nil, false, err
	}
	arc.nodeFlags = flags

	switch flags {
	case ArcsForDirectAddressing:
		numArcs, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, false, err
		}
		arc.numArcs = int(numArcs) // This is in fact the label range.

		bytesPerArc, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, false, err
		}
		arc.bytesPerArc = int(bytesPerArc)
		if err := readPresenceBytes(ctx, in, arc); err != nil {
			return nil, false, err
		}
		arc.firstLabel, err = f.ReadLabel(ctx, in)
		if err != nil {
			return nil, false, err
		}
		arc.posArcsStart = in.GetPosition()

		arcIndex := labelToMatch - arc.FirstLabel()
		if arcIndex < 0 || arcIndex >= arc.NumArcs() {
			return nil, false, nil // Before or after label range.
		}

		if ok, err := IsBitSet(ctx, arcIndex, arc, in); err != nil {
			return nil, false, err
		} else if !ok {
			return nil, false, nil // Arc missing in the range.
		}

		addressing, err := f.ReadArcByDirectAddressing(ctx, in, arcIndex, arc)
		if err != nil {
			return nil, false, err
		}
		return addressing, true, nil
	case ArcsForBinarySearch:
		numArcs, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, false, err
		}
		arc.numArcs = int(numArcs)

		bytesPerArc, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, false, err
		}
		arc.bytesPerArc = int(bytesPerArc)
		arc.posArcsStart = in.GetPosition()

		// Array is sparse; do binary search:
		low := 0
		high := arc.NumArcs() - 1
		for low <= high {
			mid := (low + high) >> 1
			// +1 to skip over flags
			if err := in.SetPosition(arc.PosArcsStart() - int64(arc.BytesPerArc()*mid+1)); err != nil {
				return nil, false, err
			}
			midLabel, err := f.ReadLabel(ctx, in)
			if err != nil {
				return nil, false, err
			}
			cmp := midLabel - labelToMatch
			if cmp < 0 {
				low = mid + 1
			} else if cmp > 0 {
				high = mid - 1
			} else {
				arc.arcIdx = mid - 1
				realArc, err := f.ReadNextRealArc(ctx, in, arc)
				if err != nil {
					return nil, false, err
				}
				return realArc, true, nil
			}
		}
		return nil, false, nil
	}

	// Linear scan
	if _, err := f.ReadFirstRealTargetArc(ctx, follow.Target(), in, arc); err != nil {
		return nil, false, err
	}

	for {
		// TODO: we should fix this code to not have to create
		// object for the output of every arc we scan... only
		// for the matching arc, if found
		if arc.Label() == labelToMatch {
			return arc, true, nil
		} else if arc.Label() > labelToMatch {
			return nil, false, nil
		} else if arc.IsLast() {
			return nil, false, nil
		} else {
			if _, err := f.ReadNextRealArc(ctx, in, arc); err != nil {
				return nil, false, err
			}
		}
	}
}

func (f *FST[T]) FindTarget(ctx context.Context, labelToMatch int, current *Arc[T], in BytesReader) (*Arc[T], bool, error) {
	targetArc := &Arc[T]{}

	if labelToMatch == END_LABEL {
		if current.IsFinal() {
			if current.Target() <= 0 {
				targetArc.flags = BIT_LAST_ARC
			} else {
				targetArc.flags = 0
				// NOTE: nextArc is a node (not an address!) in this case:
				targetArc.nextArc = current.Target()
			}
			targetArc.output = current.NextFinalOutput()
			targetArc.label = END_LABEL
			targetArc.nodeFlags = targetArc.flags
			return targetArc, true, nil
		} else {
			return nil, false, nil
		}
	}

	if !TargetHasArcs(current) {
		return nil, false, nil
	}

	if err := in.SetPosition(current.Target()); err != nil {
		return nil, false, err
	}

	flags, err := in.ReadByte()
	if err != nil {
		return nil, false, err
	}
	targetArc.nodeFlags = flags

	switch flags {
	case ArcsForDirectAddressing:
		numArcs, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, false, err
		}
		targetArc.numArcs = int(numArcs) // This is in fact the label range.

		bytesPerArc, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, false, err
		}
		targetArc.bytesPerArc = int(bytesPerArc)

		if err := readPresenceBytes(ctx, in, targetArc); err != nil {
			return nil, false, err
		}

		firstLabel, err := f.ReadLabel(ctx, in)
		if err != nil {
			return nil, false, err
		}
		targetArc.firstLabel = firstLabel

		targetArc.posArcsStart = in.GetPosition()

		arcIndex := labelToMatch - targetArc.FirstLabel()
		if arcIndex < 0 || arcIndex >= targetArc.NumArcs() {
			return nil, false, nil // Before or after label range.
		}

		if ok, err := IsBitSet(ctx, arcIndex, targetArc, in); err != nil {
			return nil, false, err
		} else if !ok {
			return nil, false, nil // Arc missing in the range.
		}

		arc, err := f.ReadArcByDirectAddressing(ctx, in, arcIndex, targetArc)
		if err != nil {
			return nil, false, err
		}
		return arc, true, nil

	case ArcsForBinarySearch:
		numArcs, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, false, err
		}
		targetArc.numArcs = int(numArcs)

		bytesPerArc, err := in.ReadUvarint(ctx)
		if err != nil {
			return nil, false, err
		}
		targetArc.bytesPerArc = int(bytesPerArc)
		targetArc.posArcsStart = in.GetPosition()

		// Array is sparse; do binary search:
		low := 0
		high := targetArc.NumArcs() - 1
		for low <= high {
			mid := (low + high) >> 1
			// +1 to skip over flags
			if err := in.SetPosition(targetArc.PosArcsStart() - int64(targetArc.BytesPerArc()*mid+1)); err != nil {
				return nil, false, err
			}
			midLabel, err := f.ReadLabel(ctx, in)
			if err != nil {
				return nil, false, err
			}
			cmp := midLabel - labelToMatch
			if cmp < 0 {
				low = mid + 1
			} else if cmp > 0 {
				high = mid - 1
			} else {
				targetArc.arcIdx = mid - 1
				arc, err := f.ReadNextRealArc(ctx, in, targetArc)
				if err != nil {
					return nil, false, err
				}
				return arc, true, nil
			}
		}
		return nil, false, nil
	}

	// Linear scan
	if _, err := f.ReadFirstRealTargetArc(ctx, current.Target(), in, targetArc); err != nil {
		return nil, false, err
	}

	for {
		// TODO: we should fix this code to not have to create
		// object for the output of every targetArc we scan... only
		// for the matching targetArc, if found
		if targetArc.Label() == labelToMatch {
			return targetArc, true, nil
		} else if targetArc.Label() > labelToMatch {
			return nil, false, nil
		} else if targetArc.IsLast() {
			return nil, false, nil
		} else {
			if _, err := f.ReadNextRealArc(ctx, in, targetArc); err != nil {
				return nil, false, err
			}
		}
	}
}

// GetBytesReader Returns a Fst.BytesReader for this FST, positioned at position 0.
func (f *FST[T]) GetBytesReader() (BytesReader, error) {
	if f.fstStore != nil {
		return f.fstStore.GetReverseBytesReader()
	}
	return f.bytes.GetReverseReader()
}

func (f *FST[T]) Finish(newStartNode int64) error {
	// TODO: assert newStartNode <= bytes.getPosition();
	if f.startNode != -1 {
		return errors.New("already finished")
	}
	if newStartNode == FINAL_END_NODE && !f.outputs.IsNoOutput(f.emptyOutput) {
		newStartNode = 0
	}
	f.startNode = newStartNode
	return f.bytes.Finish()
}

func (f *FST[T]) NumBytes() int {
	return int(f.bytes.GetPosition())
}

func (f *FST[T]) GetEmptyOutput() T {
	return f.emptyOutput
}
