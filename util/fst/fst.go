package fst

import (
	"github.com/geange/lucene-go/codecs"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	. "github.com/geange/lucene-go/math"
	"github.com/pkg/errors"
	"os"
)

const (
	// DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR Maximum oversizing factor allowed for direct addressing
	// compared to binary search when expansion credits allow the oversizing. This factor prevents expansions
	// that are obviously too costly even if there are sufficient credits.
	// See Also: houldExpandNodeWithDirectAddressing
	DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR = 1.66

	FILE_FORMAT_NAME = "FST"
	VERSION_START    = 6
	VERSION_CURRENT  = 7

	// FINAL_END_NODE Never serialized; just used to represent the virtual
	// final node w/ no arcs:
	FINAL_END_NODE = -1

	// NON_FINAL_END_NODE Never serialized; just used to represent the virtual
	// non-final node w/ no arcs:
	NON_FINAL_END_NODE = 0

	// END_LABEL If arc has this label then that arc is final/accepted
	END_LABEL = -1
)

type INPUT_TYPE int

const (
	BYTE1 = INPUT_TYPE(iota)
	BYTE2
	BYTE4
)

// TODO: break this into WritableFST and ReadOnlyFST.. then
// we can have subclasses of ReadOnlyFST to handle the
// different byte[] level encodings (packed or
// not)... and things like nodeCount, arcCount are read only

// TODO: if FST is pure prefix trie we can do a more compact
// job, ie, once we are at a 'suffix only', just store the
// completion labels as a string not as a series of arcs.

// NOTE: while the FST is able to represent a non-final
// dead-end state (NON_FINAL_END_NODE=0), the layers above
// (FSTEnum, Util) have problems with this!!

// FST Represents an finite state machine (FST), using a compact byte[] format.
// The format is similar to what's used by Morfologik (https://github.com/morfologik/morfologik-stemming).
// See the package documentation for some simple examples.
// lucene.experimental
type FST[T any] struct {
	inputType INPUT_TYPE

	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput *Box[T]

	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB).
	// If the FST is less than 1 GB then bytesArray is set instead.
	bytes *BytesStore

	//
	fstStore FSTStore

	startNode int

	outputs Outputs[T]
}

func (f *FST[T]) finish(newStartNode int) error {
	if f.startNode != -1 {
		return errors.Wrap(ErrIllegalState, "already finished")
	}

	if newStartNode == FINAL_END_NODE && f.emptyOutput != nil {
		newStartNode = 0
	}
	f.startNode = newStartNode
	f.bytes.Finish()
	return nil
}

func (f *FST[T]) GetEmptyOutput() any {
	return f.emptyOutput
}

func (f *FST[T]) setEmptyOutput(v *Box[T]) {
	if f.emptyOutput != nil {
		f.emptyOutput = f.outputs.Merge(f.emptyOutput, v)
	} else {
		f.emptyOutput = v
	}
}

func (f *FST[T]) Save(metaOut store.DataOutput, out store.DataOutput) error {
	if f.startNode == -1 {
		return errors.New("call finish first")
	}

	err := codecs.WriteHeader(metaOut, FILE_FORMAT_NAME, VERSION_CURRENT)
	if err != nil {
		return err
	}

	if f.emptyOutput != nil {
		// Accepts empty string
		metaOut.WriteByte(1)

		// Serialize empty-string output:
		ros := store.NewRAMOutputStream()

		f.outputs.WriteFinalOutput(f.emptyOutput, ros)

		emptyOutputBytes := make([]byte, ros.GetFilePointer())
		ros.WriteToV1(emptyOutputBytes[0:])
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
		metaOut.WriteUvarint(uint64(emptyLen))
		metaOut.WriteBytes(emptyOutputBytes[0:emptyLen])
	} else {
		metaOut.WriteByte(0)
	}

	t := byte(0)
	if f.inputType == BYTE1 {
		t = 0
	} else if f.inputType == BYTE2 {
		t = 1
	} else {
		t = 2
	}
	metaOut.WriteByte(t)
	metaOut.WriteUvarint(uint64(f.startNode))
	if f.bytes != nil {
		numBytes := f.bytes.getPosition()
		metaOut.WriteUvarint(uint64(numBytes))
		f.bytes.WriteTo(out)
	} else {
		f.fstStore.WriteTo(out)
	}
	return nil
}

// SaveToFile Writes an automaton to a file.
func (f *FST[T]) SaveToFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	out := store.NewOutputStreamIndexOutput(file)
	return f.Save(out, out)
}

func Read[T any](path string, outputs Outputs[T]) *FST[T] {
	panic("")
}

func (f *FST[T]) writeLabel(out store.DataOutput, v int) error {
	switch f.inputType {
	case BYTE1:
		return out.WriteByte(byte(v))
	case BYTE2:
		return out.WriteUint16(uint16(v))
	default:
		return out.WriteUvarint(uint64(v))
	}
}

// ReadLabel Reads one BYTE1/2/4 label from the provided DataInput.
func (f *FST[T]) ReadLabel(in store.DataInput) (int, error) {
	switch f.inputType {
	case BYTE1:
		v, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		return int(v & 0xFF), nil
	case BYTE2:
		v, err := in.ReadUint16()
		if err != nil {
			return 0, err
		}
		return int(v & 0xFFFF), nil
	default:
		v, err := in.ReadUvarint()
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
}

// TargetHasArcs returns true if the node at this address has any outgoing arcs
func TargetHasArcs[T any](arc *Arc[T]) bool {
	return arc.Target() > 0
}

// serializes new node by appending its bytes to the end
// of the current byte[]
func (f *FST[T]) addNode(builder *Builder[T], nodeIn *UnCompiledNode[T]) (int, error) {
	noOutput := f.outputs.GetNoOutput()

	if nodeIn.numArcs == 0 {
		if nodeIn.isFinal {
			return FINAL_END_NODE, nil
		}
		return NON_FINAL_END_NODE, nil
	}

	startAddress := builder.bytes.getPosition()
	doFixedLengthArcs := f.shouldExpandNodeWithFixedLengthArcs(builder, nodeIn)
	if doFixedLengthArcs {
		if len(builder.numBytesPerArc) < nodeIn.numArcs {
			builder.numBytesPerArc = make([]int, util.Oversize(nodeIn.numArcs, 4))
			builder.numLabelBytesPerArc = make([]int, len(builder.numBytesPerArc))
		}
	}

	builder.arcCount += nodeIn.numArcs
	lastArc := nodeIn.numArcs - 1

	lastArcStart := builder.bytes.getPosition()
	maxBytesPerArc, maxBytesPerArcWithoutLabel := 0, 0

	for arcIdx := 0; arcIdx < nodeIn.numArcs; arcIdx++ {
		arc := nodeIn.arcs[arcIdx]
		target := arc.target.(*CompiledNode)
		flags := 0

		if arcIdx == lastArc {
			flags += BIT_LAST_ARC
		}

		if builder.lastFrozenNode == int(target.node) && !doFixedLengthArcs {
			// TODO: for better perf (but more RAM used) we
			// could avoid this except when arc is "near" the
			// last arc:
			flags += BIT_TARGET_NEXT
		}

		if arc.isFinal {
			flags += BIT_FINAL_ARC
			if arc.nextFinalOutput != noOutput {
				flags += BIT_ARC_HAS_FINAL_OUTPUT
			}
		} else {

		}

		targetHasArcs := target.node > 0
		if !targetHasArcs {
			flags += BIT_STOP_NODE
		}

		if arc.output != noOutput {
			flags += BIT_ARC_HAS_OUTPUT
		}

		err := builder.bytes.WriteByte(byte(flags))
		if err != nil {
			return 0, err
		}
		labelStart := builder.bytes.getPosition()
		err = f.writeLabel(builder.bytes, arc.label)
		if err != nil {
			return 0, err
		}
		numLabelBytes := builder.bytes.getPosition() - labelStart

		if arc.output != noOutput {
			err := f.outputs.Write(arc.output, builder.bytes)
			if err != nil {
				return 0, err
			}
		}

		if arc.nextFinalOutput != noOutput {
			err := f.outputs.WriteFinalOutput(arc.nextFinalOutput, builder.bytes)
			if err != nil {
				return 0, err
			}
		}

		if targetHasArcs && (flags&BIT_TARGET_NEXT == 0) {
			err := builder.bytes.WriteUvarint(uint64(target.node))
			if err != nil {
				return 0, err
			}
		}

		// just write the arcs "like normal" on first pass, but record how many bytes each one took
		// and max byte size:
		if doFixedLengthArcs {
			numArcBytes := builder.bytes.getPosition() - lastArcStart
			builder.numBytesPerArc[arcIdx] = numArcBytes
			builder.numLabelBytesPerArc[arcIdx] = numLabelBytes
			lastArcStart = builder.bytes.getPosition()
			maxBytesPerArc = Max(maxBytesPerArc, numArcBytes)
			maxBytesPerArcWithoutLabel = Max(maxBytesPerArcWithoutLabel, numArcBytes-numLabelBytes)
		}
	}

	if doFixedLengthArcs {
		labelRange := nodeIn.arcs[nodeIn.numArcs-1].label - nodeIn.arcs[0].label + 1
		if f.shouldExpandNodeWithDirectAddressing(builder, nodeIn, maxBytesPerArc, maxBytesPerArcWithoutLabel, labelRange) {
			f.writeNodeForDirectAddressing(builder, nodeIn, startAddress, maxBytesPerArcWithoutLabel, labelRange)
			builder.directAddressingNodeCount++
		} else {
			f.writeNodeForBinarySearch(builder, nodeIn, startAddress, maxBytesPerArc)
			builder.binarySearchNodeCount++
		}
	}

	thisNodeAddress := builder.bytes.getPosition() - 1
	builder.bytes.Reverse(startAddress, thisNodeAddress)
	builder.nodeCount++
	return thisNodeAddress, nil
}

// Returns whether the given node should be expanded with fixed length arcs. Nodes will be expanded depending on their depth (distance from the root node) and their number of arcs.
// Nodes with fixed length arcs use more space, because they encode all arcs with a fixed number of bytes, but they allow either binary search or direct addressing on the arcs (instead of linear scan) on lookup by arc label.
func (f *FST[T]) shouldExpandNodeWithFixedLengthArcs(builder *Builder[T], node *UnCompiledNode[T]) bool {
	return builder.allowFixedLengthArcs &&
		((node.depth <= FIXED_LENGTH_ARC_SHALLOW_DEPTH && node.numArcs >= FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS) ||
			node.numArcs >= FIXED_LENGTH_ARC_DEEP_NUM_ARCS)
}

// Returns whether the given node should be expanded with direct addressing instead of binary search.
// Prefer direct addressing for performance if it does not oversize binary search byte size too much, so that the arcs can be directly addressed by label.
// See Also:
// Builder.getDirectAddressingMaxOversizingFactor()
func (f *FST[T]) shouldExpandNodeWithDirectAddressing(builder *Builder[T], nodeIn *UnCompiledNode[T],
	numBytesPerArc, maxBytesPerArcWithoutLabel, labelRange int) bool {

	// Anticipate precisely the size of the encodings.
	sizeForBinarySearch := numBytesPerArc * nodeIn.numArcs
	sizeForDirectAddressing := getNumPresenceBytes(labelRange) + builder.numLabelBytesPerArc[0] +
		maxBytesPerArcWithoutLabel*nodeIn.numArcs

	// Determine the allowed oversize compared to binary search.
	// This is defined by a parameter of FST Builder (default 1: no oversize).
	allowedOversize := sizeForBinarySearch + int(builder.GetDirectAddressingMaxOversizingFactor())
	expansionCost := sizeForDirectAddressing - allowedOversize

	// Select direct addressing if either:
	// - Direct addressing size is smaller than binary search.
	//   In this case, increment the credit by the reduced size (to use it later).
	// - Direct addressing size is larger than binary search, but the positive credit allows the oversizing.
	//   In this case, decrement the credit by the oversize.
	// In addition, do not try to oversize to a clearly too large node size
	// (this is the DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR parameter).
	if expansionCost <= 0 || (builder.directAddressingExpansionCredit >= expansionCost &&
		sizeForDirectAddressing <= int(float64(allowedOversize)*DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR)) {
		builder.directAddressingExpansionCredit -= expansionCost
		return true
	}
	return false
}

func (f *FST[T]) writeNodeForBinarySearch(builder *Builder[T], nodeIn *UnCompiledNode[T],
	startAddress, maxBytesPerArc int) {
	// Build the header in a buffer.
	// It is a false/special arc which is in fact a node header with node flags followed by node metadata.
	builder.fixedLengthArcsBuffer.
		resetPosition().
		writeByte(ARCS_FOR_BINARY_SEARCH).
		writeVInt(nodeIn.numArcs).
		writeVInt(maxBytesPerArc)

	headerLen := builder.fixedLengthArcsBuffer.getPosition()

	// Expand the arcs in place, backwards.
	srcPos := builder.bytes.getPosition()
	destPos := startAddress + headerLen + nodeIn.numArcs*maxBytesPerArc

	if destPos > srcPos {
		builder.bytes.SkipBytes((int)(destPos - srcPos))
		for arcIdx := nodeIn.numArcs - 1; arcIdx >= 0; arcIdx-- {
			destPos -= maxBytesPerArc
			arcLen := builder.numBytesPerArc[arcIdx]
			srcPos -= arcLen
			if srcPos != destPos {
				builder.bytes.CopyBytesSelf(srcPos, destPos, arcLen)
			}
		}
	}

	// Write the header.
	builder.bytes.WriteBytesAt(startAddress, builder.fixedLengthArcsBuffer.getBytes()[0:headerLen])
}

func getNumPresenceBytes(labelRange int) int {
	return (labelRange + 7) >> 3
}

func (f *FST[T]) writeNodeForDirectAddressing(builder *Builder[T], nodeIn *UnCompiledNode[T],
	startAddress, maxBytesPerArcWithoutLabel, labelRange int) {

	// Expand the arcs backwards in a buffer because we remove the labels.
	// So the obtained arcs might occupy less space. This is the reason why this
	// whole method is more complex.
	// Drop the label bytes since we can infer the label based on the arc index,
	// the presence bits, and the first label. Keep the first label.
	headerMaxLen := 11
	numPresenceBytes := getNumPresenceBytes(labelRange)
	srcPos := builder.bytes.getPosition()
	totalArcBytes := builder.numLabelBytesPerArc[0] + nodeIn.numArcs*maxBytesPerArcWithoutLabel
	bufferOffset := headerMaxLen + numPresenceBytes + totalArcBytes
	buffer := builder.fixedLengthArcsBuffer.ensureCapacity(bufferOffset).getBytes()
	// Copy the arcs to the buffer, dropping all labels except first one.
	for arcIdx := nodeIn.numArcs - 1; arcIdx >= 0; arcIdx-- {
		bufferOffset -= maxBytesPerArcWithoutLabel
		srcArcLen := builder.numBytesPerArc[arcIdx]
		srcPos -= srcArcLen
		labelLen := builder.numLabelBytesPerArc[arcIdx]
		// Copy the flags.
		builder.bytes.CopyBytesToArray(srcPos, buffer[bufferOffset:bufferOffset+1])
		// Skip the label, copy the remaining.
		remainingArcLen := srcArcLen - 1 - labelLen
		if remainingArcLen != 0 {
			builder.bytes.CopyBytesToArray(srcPos+1+labelLen, buffer[bufferOffset+1:bufferOffset+1+remainingArcLen])
		}
		if arcIdx == 0 {
			// Copy the label of the first arc only.
			bufferOffset -= labelLen
			builder.bytes.CopyBytesToArray(srcPos+1, buffer[bufferOffset:bufferOffset+labelLen])
		}
	}

	// Build the header in the buffer.
	// It is a false/special arc which is in fact a node header with node flags followed by node metadata.
	builder.fixedLengthArcsBuffer.
		resetPosition().
		writeByte(ARCS_FOR_DIRECT_ADDRESSING).
		writeVInt(labelRange) // labelRange instead of numArcs. .writeVInt(maxBytesPerArcWithoutLabel); // maxBytesPerArcWithoutLabel instead of maxBytesPerArc.
	headerLen := builder.fixedLengthArcsBuffer.getPosition()

	// Prepare the builder byte store. Enlarge or truncate if needed.
	nodeEnd := startAddress + headerLen + numPresenceBytes + totalArcBytes
	currentPosition := builder.bytes.getPosition()
	if nodeEnd >= currentPosition {
		builder.bytes.SkipBytes(nodeEnd - currentPosition)
	} else {
		builder.bytes.Truncate(nodeEnd)
	}

	// Write the header.
	writeOffset := startAddress
	builder.bytes.WriteBytesAt(writeOffset, builder.fixedLengthArcsBuffer.getBytes()[0:headerLen])
	writeOffset += headerLen

	// Write the presence bits
	f.writePresenceBits(builder, nodeIn, writeOffset, numPresenceBytes)
	writeOffset += numPresenceBytes

	// Write the first label and the arcs.
	builder.bytes.WriteBytesAt(writeOffset, builder.fixedLengthArcsBuffer.getBytes()[bufferOffset:bufferOffset+totalArcBytes])
}

func (f *FST[T]) writePresenceBits(builder *Builder[T], nodeIn *UnCompiledNode[T],
	dest, numPresenceBytes int) {

	bytePos := dest
	presenceBits := byte(1) // The first arc is always present.
	presenceIndex := 0
	previousLabel := nodeIn.arcs[0].label
	for arcIdx := 1; arcIdx < nodeIn.numArcs; arcIdx++ {
		label := nodeIn.arcs[arcIdx].label
		presenceIndex += label - previousLabel
		for presenceIndex >= ByteSize {
			builder.bytes.WriteByteAt(bytePos, presenceBits)
			bytePos++
			presenceBits = 0
			presenceIndex -= ByteSize
		}
		// Set the bit at presenceIndex to flag that the corresponding arc is present.
		presenceBits |= 1 << presenceIndex
		previousLabel = label
	}
	builder.bytes.WriteByteAt(bytePos, presenceBits)
	bytePos++

}

// Reads the presence bits of a direct-addressing node. Actually we don't read them here,
// we just keep the pointer to the bit-table start and we skip them.
func (f *FST[T]) readPresenceBytes(arc *Arc[T], in BytesReader) error {
	arc.bitTableStart = in.GetPosition()
	return in.SkipBytes(getNumPresenceBytes(arc.NumArcs()))
}

// Fills virtual 'start' arc, ie, an empty incoming arc to the FST's start node
func (f *FST[T]) getFirstArc(arc *Arc[T]) *Arc[T] {
	noOutput := f.outputs.GetNoOutput()

	if f.emptyOutput != nil {
		arc.flags = BIT_FINAL_ARC | BIT_LAST_ARC
		arc.nextFinalOutput = f.emptyOutput
		if f.emptyOutput != noOutput {
			arc.flags = (byte)(arc.Flags() | BIT_ARC_HAS_FINAL_OUTPUT)
		}
	} else {
		arc.flags = BIT_LAST_ARC
		arc.nextFinalOutput = noOutput
	}
	arc.output = noOutput

	// If there are no nodes, ie, the FST only accepts the
	// empty string, then startNode is 0
	arc.target = f.startNode
	return arc
}

// Follows the follow arc and reads the last arc of its target; this changes the provided arc (2nd arg) in-place and returns it.
// Returns: Returns the second argument (arc).
func (f *FST[T]) readLastTargetArc(follow *Arc[T], arc *Arc[T], in BytesReader) (*Arc[T], error) {
	if !targetHasArcs(follow) {
		arc.label = END_LABEL
		arc.target = FINAL_END_NODE
		arc.output = follow.NextFinalOutput()
		arc.flags = BIT_LAST_ARC
		arc.nodeFlags = arc.flags
		return arc, nil
	} else {
		in.SetPosition(follow.Target())

		v, err := in.ReadByte()
		if err != nil {
			return nil, err
		}
		flags := v
		arc.nodeFlags = v
		if flags == ARCS_FOR_BINARY_SEARCH || flags == ARCS_FOR_DIRECT_ADDRESSING {
			// Special arc which is actually a node header for fixed length arcs.
			// Jump straight to end to find the last arc.
			numArcs, err := in.ReadUvarint()
			if err != nil {
				return nil, err
			}
			arc.numArcs = int(numArcs)
			bytesPerArc, err := in.ReadUvarint()
			if err != nil {
				return nil, err
			}
			arc.bytesPerArc = int(bytesPerArc)
			//System.out.println("  array numArcs=" + arc.numArcs + " bpa=" + arc.bytesPerArc);
			if flags == ARCS_FOR_DIRECT_ADDRESSING {
				f.readPresenceBytes(arc, in)
				arc.firstLabel, err = f.ReadLabel(in)
				arc.posArcsStart = in.GetPosition()
				f.ReadLastArcByDirectAddressing(arc, in)
			} else {
				arc.arcIdx = arc.NumArcs() - 2
				arc.posArcsStart = in.GetPosition()
				f.ReadNextRealArc(arc, in)
			}
		} else {
			arc.flags = flags
			// non-array: linear scan
			arc.bytesPerArc = 0
			//System.out.println("  scan");
			for !arc.IsLast() {
				// skip this arc:
				f.ReadLabel(in)
				if arc.Flag(BIT_ARC_HAS_OUTPUT) {
					f.outputs.SkipOutput(in)
				}
				if arc.Flag(BIT_ARC_HAS_FINAL_OUTPUT) {
					f.outputs.SkipFinalOutput(in)
				}
				if arc.Flag(BIT_STOP_NODE) {
				} else if arc.Flag(BIT_TARGET_NEXT) {
				} else {
					f.readUnpackedNodeTarget(in)
				}
				flags, err := in.ReadByte()
				if err != nil {
					return nil, err
				}
				arc.flags = flags
			}
			// Undo the byte flags we read:
			in.SkipBytes(-1)
			arc.nextArc = in.GetPosition()
			f.ReadNextRealArc(arc, in)
		}
		return arc, nil
	}
}

func (f *FST[T]) readUnpackedNodeTarget(in BytesReader) (int, error) {
	num, err := in.ReadUvarint()
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

// ReadFirstTargetArc Follow the follow arc and read the first arc of its target; this changes
// the provided arc (2nd arg) in-place and returns it.
// Returns: Returns the second argument (arc).
func (f *FST[T]) ReadFirstTargetArc(follow, arc *Arc[T], in BytesReader) (*Arc[T], error) {
	//int pos = address;
	//System.out.println("    readFirstTarget follow.target=" + follow.target + " isFinal=" + follow.isFinal());
	if follow.IsFinal() {
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
		//System.out.println("    insert isFinal; nextArc=" + follow.target + " isLast=" + arc.isLast() + " output=" + outputs.outputToString(arc.output));
		return arc, nil
	} else {
		return f.ReadFirstRealTargetArc(follow.Target(), arc, in)
	}
}

func (f *FST[T]) ReadFirstRealTargetArc(nodeAddress int, arc *Arc[T], in BytesReader) (*Arc[T], error) {
	in.SetPosition(nodeAddress)
	//System.out.println("   flags=" + arc.flags);

	b, err := in.ReadByte()
	if err != nil {
		return nil, err
	}

	flags := b
	arc.nodeFlags = b
	if flags == ARCS_FOR_BINARY_SEARCH || flags == ARCS_FOR_DIRECT_ADDRESSING {
		//System.out.println("  fixed length arc");
		// Special arc which is actually a node header for fixed length arcs.
		numArcs, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.numArcs = int(numArcs)

		bytesPerArc, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc = int(bytesPerArc)
		arc.arcIdx = -1
		if flags == ARCS_FOR_DIRECT_ADDRESSING {
			err := f.readPresenceBytes(arc, in)
			if err != nil {
				return nil, err
			}
			arc.firstLabel, err = f.ReadLabel(in)
			if err != nil {
				return nil, err
			}
			arc.presenceIndex = -1
		}
		arc.posArcsStart = in.GetPosition()
		//System.out.println("  bytesPer=" + arc.bytesPerArc + " numArcs=" + arc.numArcs + " arcsStart=" + pos);
	} else {
		arc.nextArc = nodeAddress
		arc.bytesPerArc = 0
	}

	return f.ReadNextRealArc(arc, in)
}

// Returns whether arc's target points to a node in expanded format (fixed length arcs).
func (f *FST[T]) isExpandedTarget(follow *Arc[T], in BytesReader) (bool, error) {
	if !targetHasArcs(follow) {
		return false, nil
	} else {
		in.SetPosition(follow.Target())
		flags, err := in.ReadByte()
		if err != nil {
			return false, err
		}
		return flags == ARCS_FOR_BINARY_SEARCH || flags == ARCS_FOR_DIRECT_ADDRESSING, nil
	}
}

// ReadNextArc In-place read; returns the arc.
func (f *FST[T]) ReadNextArc(arc *Arc[T], in BytesReader) (*Arc[T], error) {
	if arc.Label() == END_LABEL {
		// This was a fake inserted "final" arc
		if arc.NextArc() <= 0 {
			return nil, errors.New("cannot readNextArc when arc.isLast()=true")
		}
		return f.ReadFirstRealTargetArc(arc.NextArc(), arc, in)
	} else {
		return f.ReadNextRealArc(arc, in)
	}
}

// Peeks at next arc's label; does not alter arc. Do not call this if arc.isLast()!
func (f *FST[T]) readNextArcLabel(arc *Arc[T], in BytesReader) (int, error) {
	if arc.Label() == END_LABEL {
		//System.out.println("    nextArc fake " + arc.nextArc);
		// Next arc is the first arc of a node.
		// Position to read the first arc label.

		in.SetPosition(arc.NextArc())
		flags, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		if flags == ARCS_FOR_BINARY_SEARCH || flags == ARCS_FOR_DIRECT_ADDRESSING {
			//System.out.println("    nextArc fixed length arc");
			// Special arc which is actually a node header for fixed length arcs.
			numArcs, err := in.ReadUvarint()
			if err != nil {
				return 0, err
			}
			in.ReadUvarint() // Skip bytesPerArc.
			if flags == ARCS_FOR_BINARY_SEARCH {
				in.ReadByte() // Skip arc flags.
			} else {
				in.SkipBytes(getNumPresenceBytes(int(numArcs)))
			}
		}
	} else {
		if arc.BytesPerArc() != 0 {
			//System.out.println("    nextArc real array");
			// Arcs have fixed length.
			if arc.NodeFlags() == ARCS_FOR_BINARY_SEARCH {
				// Point to next arc, -1 to skip arc flags.
				in.SetPosition(arc.PosArcsStart() - (1+arc.ArcIdx())*arc.BytesPerArc() - 1)
			} else {
				// Direct addressing node. The label is not stored but rather inferred
				// based on first label and arc index in the range.
				nextIndex := NextBitSet(arc.ArcIdx(), arc, in)
				return arc.FirstLabel() + nextIndex, nil
			}
		} else {
			// Arcs have variable length.
			//System.out.println("    nextArc real list");
			// Position to next arc, -1 to skip flags.
			in.SetPosition(arc.NextArc() - 1)
		}
	}
	return f.ReadLabel(in)
}

func (f *FST[T]) ReadArcByIndex(arc *Arc[T], in BytesReader, idx int) (*Arc[T], error) {
	in.SetPosition(arc.PosArcsStart() - idx*arc.BytesPerArc())
	arc.arcIdx = idx

	b, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	arc.flags = b
	return f.readArc(arc, in)
}

// ReadArcByDirectAddressing Reads a present direct addressing node arc, with the provided index in the label range.
// Params: rangeIndex – The index of the arc in the label range. It must be present. The real arc offset is computed based on the presence bits of the direct addressing node.
func (f *FST[T]) ReadArcByDirectAddressing(arc *Arc[T], in BytesReader, rangeIndex int) (*Arc[T], error) {
	presenceIndex := CountBitsUpTo(rangeIndex, arc, in)
	return f.ReadArcByDirectAddressingV1(arc, in, rangeIndex, presenceIndex)
}

// ReadArcByDirectAddressingV1 Reads a present direct addressing node arc, with the provided
// index in the label range and its corresponding presence index (which is the count of presence bits before it).
func (f *FST[T]) ReadArcByDirectAddressingV1(arc *Arc[T], in BytesReader, rangeIndex, presenceIndex int) (*Arc[T], error) {
	in.SetPosition(arc.PosArcsStart() - presenceIndex*arc.BytesPerArc())
	arc.arcIdx = rangeIndex
	arc.presenceIndex = presenceIndex
	var err error
	arc.flags, err = in.ReadByte()
	if err != nil {
		return nil, err
	}
	return f.readArc(arc, in)
}

// ReadLastArcByDirectAddressing Reads the last arc of a direct addressing node.
// This method is equivalent to call readArcByDirectAddressing(FST.Arc, FST.BytesReader, int)
// with rangeIndex equal to arc.numArcs() - 1, but it is faster.
func (f *FST[T]) ReadLastArcByDirectAddressing(arc *Arc[T], in BytesReader) (*Arc[T], error) {
	presenceIndex := CountBits(arc, in) - 1
	return f.ReadArcByDirectAddressingV1(arc, in, arc.NumArcs()-1, presenceIndex)
}

// ReadNextRealArc Never returns null, but you should never call this if arc.isLast() is true.
func (f *FST[T]) ReadNextRealArc(arc *Arc[T], in BytesReader) (*Arc[T], error) {
	var err error
	switch arc.NodeFlags() {

	case ARCS_FOR_BINARY_SEARCH:
		arc.arcIdx++
		in.SetPosition(arc.PosArcsStart() - arc.ArcIdx()*arc.BytesPerArc())
		arc.flags, err = in.ReadByte()
		if err != nil {
			return nil, err
		}
		break

	case ARCS_FOR_DIRECT_ADDRESSING:

		nextIndex := NextBitSet(arc.ArcIdx(), arc, in)
		return f.ReadArcByDirectAddressingV1(arc, in, nextIndex, arc.presenceIndex+1)

	default:
		// Variable length arcs - linear search.

		in.SetPosition(arc.NextArc())
		arc.flags, err = in.ReadByte()
		if err != nil {
			return nil, err
		}
	}
	return f.readArc(arc, in)
}

// Reads an arc. Precondition: The arc flags byte has already been read and set; the given BytesReader is positioned just after the arc flags byte.
func (f *FST[T]) readArc(arc *Arc[T], in BytesReader) (*Arc[T], error) {
	var err error

	if arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING {
		arc.label = arc.FirstLabel() + arc.ArcIdx()
	} else {
		arc.label, err = f.ReadLabel(in)
		if err != nil {
			return nil, err
		}
	}

	if arc.Flag(BIT_ARC_HAS_OUTPUT) {
		arc.output, err = f.outputs.Read(in)
		if err != nil {
			return nil, err
		}
	} else {
		arc.output = f.outputs.GetNoOutput()
	}

	if arc.Flag(BIT_ARC_HAS_FINAL_OUTPUT) {
		arc.nextFinalOutput, err = f.outputs.ReadFinalOutput(in)
		if err != nil {
			return nil, err
		}
	} else {
		arc.nextFinalOutput = f.outputs.GetNoOutput()
	}

	if arc.Flag(BIT_STOP_NODE) {
		if arc.Flag(BIT_FINAL_ARC) {
			arc.target = FINAL_END_NODE
		} else {
			arc.target = NON_FINAL_END_NODE
		}
		arc.nextArc = in.GetPosition() // Only useful for list.
	} else if arc.Flag(BIT_TARGET_NEXT) {
		arc.nextArc = in.GetPosition() // Only useful for list.
		// TODO: would be nice to make this lazy -- maybe
		// caller doesn't need the target and is scanning arcs...
		if !arc.Flag(BIT_LAST_ARC) {
			if arc.BytesPerArc() == 0 {
				// must scan
				f.seekToNextNode(in)
			} else {
				numArcs := arc.NumArcs()

				if arc.nodeFlags == ARCS_FOR_DIRECT_ADDRESSING {
					numArcs = CountBits(arc, in)
				}

				in.SetPosition(arc.PosArcsStart() - arc.BytesPerArc()*numArcs)
			}
		}
		arc.target = in.GetPosition()
	} else {
		arc.target, err = f.readUnpackedNodeTarget(in)
		if err != nil {
			return nil, err
		}
		arc.nextArc = in.GetPosition() // Only useful for list.
	}
	return arc, nil
}

func readEndArc[T any](follow, arc *Arc[T]) *Arc[T] {
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
		return arc
	} else {
		return nil
	}
}

// FindTargetArc Finds an arc leaving the incoming arc, replacing the arc in place.
// This returns null if the arc was not found, else the incoming arc.
func (f *FST[T]) FindTargetArc(labelToMatch int, follow, arc *Arc[T], in BytesReader) (*Arc[T], error) {
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
			return arc, nil
		} else {
			return nil, nil
		}
	}

	if !targetHasArcs(follow) {
		return nil, nil
	}

	in.SetPosition(follow.Target())

	// System.out.println("fta label=" + (char) labelToMatch);

	b, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	flags := b
	arc.nodeFlags = b

	if flags == ARCS_FOR_DIRECT_ADDRESSING {
		numArcs, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.numArcs = int(numArcs) // This is in fact the label range.

		bytesPerArc, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc = int(bytesPerArc)

		f.readPresenceBytes(arc, in)
		arc.firstLabel, err = f.ReadLabel(in)
		if err != nil {
			return nil, err
		}
		arc.posArcsStart = in.GetPosition()

		arcIndex := labelToMatch - arc.FirstLabel()
		if arcIndex < 0 || arcIndex >= arc.NumArcs() {
			return nil, nil // Before or after label range.
		} else if !IsBitSet(arcIndex, arc, in) {
			return nil, nil // Arc missing in the range.
		}
		return f.ReadArcByDirectAddressing(arc, in, arcIndex)
	} else if flags == ARCS_FOR_BINARY_SEARCH {
		numArcs, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.numArcs = int(numArcs)

		bytesPerArc, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc = int(bytesPerArc)
		arc.posArcsStart = in.GetPosition()

		// Array is sparse; do binary search:
		low := 0
		high := arc.NumArcs() - 1
		for low <= high {
			//System.out.println("    cycle");
			mid := (low + high) >> 1
			// +1 to skip over flags
			in.SetPosition(arc.PosArcsStart() - (arc.BytesPerArc()*mid + 1))
			midLabel, err := f.ReadLabel(in)
			if err != nil {
				return nil, err
			}
			cmp := midLabel - labelToMatch
			if cmp < 0 {
				low = mid + 1
			} else if cmp > 0 {
				high = mid - 1
			} else {
				arc.arcIdx = mid - 1
				//System.out.println("    found!");
				return f.ReadNextRealArc(arc, in)
			}
		}
		return nil, nil
	}

	// Linear scan
	f.ReadFirstRealTargetArc(follow.Target(), arc, in)

	for {
		//System.out.println("  non-bs cycle");
		// TODO: we should fix this code to not have to create
		// object for the output of every arc we scan... only
		// for the matching arc, if found
		if arc.Label() == labelToMatch {
			//System.out.println("    found!");
			return arc, nil
		} else if arc.Label() > labelToMatch {
			return nil, nil
		} else if arc.IsLast() {
			return nil, nil
		} else {
			f.ReadNextRealArc(arc, in)
		}
	}
}

func (f *FST[T]) seekToNextNode(in BytesReader) error {
	for {

		flags, err := in.ReadByte()
		if err != nil {
			return err
		}
		_, err = f.ReadLabel(in)
		if err != nil {
			return err
		}

		if flag(int(flags), BIT_ARC_HAS_OUTPUT) {
			err := f.outputs.SkipOutput(in)
			if err != nil {
				return err
			}
		}

		if flag(int(flags), BIT_ARC_HAS_FINAL_OUTPUT) {
			err := f.outputs.SkipFinalOutput(in)
			if err != nil {
				return err
			}
		}

		if !flag(int(flags), BIT_STOP_NODE) && !flag(int(flags), BIT_TARGET_NEXT) {
			_, err := f.readUnpackedNodeTarget(in)
			if err != nil {
				return err
			}
		}

		if flag(int(flags), BIT_LAST_ARC) {
			return nil
		}
	}
}

// GetBytesReader Returns a FST.BytesReader for this FST, positioned at position 0.
func (f *FST[T]) GetBytesReader() BytesReader {
	if f.fstStore != nil {
		return f.fstStore.GetReverseBytesReader()
	} else {
		return f.bytes.GetForwardReader()
	}
}

func targetHasArcs[T any](arc *Arc[T]) bool {
	return arc.Target() > 0
}

// BytesReader Reads bytes stored in an FST.
type BytesReader interface {
	store.DataInput

	// GetPosition Get current read position.
	GetPosition() int

	// SetPosition Set current read position.
	SetPosition(pos int)

	// Reversed Returns true if this reader uses reversed bytes under-the-hood.
	Reversed() bool
}

const (
	ByteSize  = 8
	ByteBytes = 1

	IntegerSize = 32
	LongSize    = 64
	LongBytes   = 8
)
