package fst

import (
	"errors"
	"github.com/geange/lucene-go/codecs"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"os"
	"reflect"
)

var (
	BIT_FINAL_ARC   = 1 << 0
	BIT_LAST_ARC    = 1 << 1
	BIT_TARGET_NEXT = 1 << 2

	// BIT_STOP_NODE TODO: we can free up a bit if we can nuke this:
	BIT_STOP_NODE = 1 << 3

	// BIT_ARC_HAS_OUTPUT This flag is set if the arc has an output.
	BIT_ARC_HAS_OUTPUT = 1 << 4

	BIT_ARC_HAS_FINAL_OUTPUT = 1 << 5

	// ARCS_FOR_BINARY_SEARCH Value of the arc flags to declare a node with fixed length arcs designed for binary search.
	ARCS_FOR_BINARY_SEARCH = byte(BIT_ARC_HAS_FINAL_OUTPUT)

	// ARCS_FOR_DIRECT_ADDRESSING Value of the arc flags to declare a node with fixed length arcs and
	// bit table designed for direct addressing.
	ARCS_FOR_DIRECT_ADDRESSING = 1 << 6

	// FIXED_LENGTH_ARC_SHALLOW_DEPTH See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_SHALLOW_DEPTH = 3

	// FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS = 5

	// FIXED_LENGTH_ARC_DEEP_NUM_ARCS See Also: shouldExpandNodeWithFixedLengthArcs
	FIXED_LENGTH_ARC_DEEP_NUM_ARCS = 10

	// DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR Maximum oversizing factor allowed for direct
	// addressing compared to binary search when expansion credits allow the oversizing. This factor
	// prevents expansions that are obviously too costly even if there are sufficient credits.
	// See Also: shouldExpandNodeWithDirectAddressing
	DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR = 1.66

	// FILE_FORMAT_NAME Increment version to change it
	FILE_FORMAT_NAME = "FST"

	VERSION_START = 6

	VERSION_CURRENT = 7

	// FINAL_END_NODE Never serialized; just used to represent the virtual
	// final node w/ no arcs:
	FINAL_END_NODE = -1

	// NON_FINAL_END_NODE Never serialized; just used to represent the virtual
	// non-final node w/ no arcs:
	NON_FINAL_END_NODE = 0

	// END_LABEL If arc has this label then that arc is final/accepted
	END_LABEL = -1

	DEFAULT_MAX_BLOCK_BITS = 30
)

// FST Represents an finite state machine (FST), using a compact byte[] format.
// The format is similar to what's used by Morfologik (https://github.com/morfologik/morfologik-stemming).
// See the package documentation for some simple examples.
type FST[T any] struct {
	inputType InputType

	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput *T

	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB). If the FST is less than 1 GB then bytesArray is set instead.
	bytes *BytesStore

	fstStore  FSTStore
	startNode int64
	outputs   Outputs[T]
}

func NewFST[T any](metaIn, in store.DataInput, outputs Outputs[T]) (*FST[T], error) {
	fstStore := NewOnHeapFSTStore(DEFAULT_MAX_BLOCK_BITS)
	return NewFSTV1(metaIn, in, fstStore, outputs)
}

func NewFSTV1[T any](metaIn, in store.DataInput, fstStore FSTStore, outputs Outputs[T]) (*FST[T], error) {
	fst := &FST[T]{
		bytes:    nil,
		fstStore: fstStore,
		outputs:  outputs,
	}
	_, err := codecs.CheckHeader(metaIn, FILE_FORMAT_NAME, VERSION_START, VERSION_CURRENT)
	if err != nil {
		return nil, err
	}
	b, err := metaIn.ReadByte()
	if err != nil {
		return nil, err
	}
	if b == 1 {
		// accepts empty string
		// 1 KB blocks:
		emptyBytes := NewBytesStore(10)
		numBytes, err := metaIn.ReadUvarint()
		if err != nil {
			return nil, err
		}
		err = emptyBytes.CopyBytes(metaIn, int(numBytes))
		if err != nil {
			return nil, err
		}

		// De-serialize empty-string output:
		reader := emptyBytes.GetReverseReader()
		// NoOutputs uses 0 bytes when writing its output,
		// so we have to check here else BytesStore gets
		// angry:
		if numBytes > 0 {
			err := reader.SetPosition(int64(numBytes - 1))
			if err != nil {
				return nil, err
			}
		}
		v, err := outputs.(OutputsExt[T]).ReadFinalOutput(reader)
		if err != nil {
			return nil, err
		}
		fst.emptyOutput = &v
	} else {
		fst.emptyOutput = nil
	}

	t, err := metaIn.ReadByte()
	if err != nil {
		return nil, err
	}
	switch t {
	case 0:
		fst.inputType = BYTE1
	case 1:
		fst.inputType = BYTE2
	case 2:
		fst.inputType = BYTE4
	default:
		return nil, errors.New("invalid input type")
	}

	node, err := metaIn.ReadUvarint()
	if err != nil {
		return nil, err
	}
	fst.startNode = int64(node)

	numBytes, err := metaIn.ReadUvarint()
	if err != nil {
		return nil, err
	}
	err = fst.fstStore.Init(in, int64(numBytes))
	if err != nil {
		return nil, err
	}
	return fst, nil
}

func (f *FST[T]) finish(newStartNode int64) error {
	if f.startNode != -1 {
		return errors.New("already finished")
	}
	if newStartNode == int64(FINAL_END_NODE) && f.emptyOutput != nil {
		newStartNode = 0
	}
	f.startNode = newStartNode
	f.bytes.Finish()
	return nil
}

func (f *FST[T]) GetEmptyOutput() any {
	return f.emptyOutput
}

func (f *FST[T]) setEmptyOutput(v T) {
	if f.emptyOutput != nil {
		*f.emptyOutput = f.outputs.(*OutputsImp[T]).Merge(*f.emptyOutput, v)
	} else {
		f.emptyOutput = &v
	}
}

func (f *FST[T]) Save(metaOut, out store.DataOutput) error {
	if f.startNode == -1 {
		return errors.New("call finish first")
	}
	if err := codecs.WriteHeader(metaOut, FILE_FORMAT_NAME, VERSION_CURRENT); err != nil {
		return err
	}
	// TODO: really we should encode this as an arc, arriving
	// to the root node, instead of special casing here:
	if f.emptyOutput != nil {
		if err := metaOut.WriteByte(1); err != nil {
			return err
		}
		ros := store.NewRAMOutputStream()

		err := f.outputs.(OutputsExt[T]).WriteFinalOutput(*f.emptyOutput, ros)
		if err != nil {
			return err
		}

		emptyOutputBytes := make([]byte, ros.GetFilePointer())
		err = ros.WriteToV1(emptyOutputBytes)
		if err != nil {
			return err
		}
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

		err = metaOut.WriteUvarint(uint64(emptyLen))
		if err != nil {
			return err
		}
		err = metaOut.WriteBytes(emptyOutputBytes)
		if err != nil {
			return err
		}
	} else {
		if err := metaOut.WriteByte(0); err != nil {
			return err
		}
	}

	var t byte
	if f.inputType == BYTE1 {
		t = 0
	} else if f.inputType == BYTE2 {
		t = 1
	} else {
		t = 2
	}
	err := metaOut.WriteByte(t)
	if err != nil {
		return err
	}
	err = metaOut.WriteUvarint(uint64(f.startNode))
	if err != nil {
		return err
	}
	if f.bytes != nil {
		numBytes := f.bytes.GetPosition()
		err := metaOut.WriteUvarint(uint64(numBytes))
		if err != nil {
			return err
		}
		err = f.bytes.WriteTo(out)
		if err != nil {
			return err
		}
	} else {
		err := f.fstStore.WriteTo(out)
		if err != nil {
			return err
		}
	}
	return nil
}

// SaveToFile Writes an automaton to a file.
func (f *FST[T]) SaveToFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	out := store.NewOutputStreamDataOutput(file)
	return f.Save(out, out)
}

// Read Reads an automaton from a file.
func Read[T any](path string, outputs Outputs[T]) (*FST[T], error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	in := store.NewInputStreamDataInput(file)
	return NewFST(in, in, outputs)
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

// Reads one BYTE1/2/4 label from the provided DataInput.
func (f *FST[T]) readLabel(in store.DataInput) (int, error) {
	switch f.inputType {
	case BYTE1:
		v, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		return int(v), nil
	case BYTE2:
		v, err := in.ReadUint16()
		if err != nil {
			return 0, err
		}
		return int(v), nil
	default:
		v, err := in.ReadUvarint()
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
}

// returns true if the node at this address has any outgoing arcs
func targetHasArcs[T any](arc *Arc[T]) bool {
	return arc.Target() > 0
}

// serializes new node by appending its bytes to the end
// of the current byte[]
func (f *FST[T]) addNode(builder *Builder[T], nodeIn *UnCompiledNode[T]) (int64, error) {
	NO_OUTPUT := f.outputs.GetNoOutput()

	if nodeIn.numArcs == 0 {
		if nodeIn.isFinal {
			return int64(FINAL_END_NODE), nil
		} else {
			return int64(NON_FINAL_END_NODE), nil
		}
	}

	startAddress := builder.bytes.GetPosition()

	doFixedLengthArcs := f.shouldExpandNodeWithFixedLengthArcs(builder, nodeIn)
	if doFixedLengthArcs {
		if len(builder.numBytesPerArc) < nodeIn.numArcs {
			// Integer.BYTES
			builder.numBytesPerArc = make([]int, util.Oversize(nodeIn.numArcs, 4))
			builder.numLabelBytesPerArc = make([]int, len(builder.numBytesPerArc))
		}
	}

	builder.arcCount += int64(nodeIn.numArcs)

	lastArc := nodeIn.numArcs - 1

	lastArcStart := builder.bytes.GetPosition()
	maxBytesPerArc := 0
	maxBytesPerArcWithoutLabel := 0
	for arcIdx := 0; arcIdx < nodeIn.numArcs; arcIdx++ {
		arc := nodeIn.arcs[arcIdx]
		target := arc.Target.(*CompiledNode)
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
			// arc.nextFinalOutput != NO_OUTPUT
			// TODO: fix
			if !reflect.DeepEqual(arc.NextFinalOutput, NO_OUTPUT) {
				flags += BIT_ARC_HAS_FINAL_OUTPUT
			}
		}

		targetHasArcs := target.node > 0

		if !targetHasArcs {
			flags += BIT_STOP_NODE
		}

		// TODO: arc.output != NO_OUTPUT
		if !reflect.DeepEqual(arc.Output, NO_OUTPUT) {
			flags += BIT_ARC_HAS_OUTPUT
		}

		err := builder.bytes.WriteByte(byte(flags))
		if err != nil {
			return 0, err
		}
		labelStart := builder.bytes.GetPosition()
		err = f.writeLabel(builder.bytes, arc.Label)
		if err != nil {
			return 0, err
		}
		numLabelBytes := builder.bytes.GetPosition() - labelStart

		// arc.output != NO_OUTPUT
		if !reflect.DeepEqual(arc.Output, NO_OUTPUT) {
			err := f.outputs.Write(arc.Output, builder.bytes)
			if err != nil {
				return 0, err
			}
		}

		// arc.nextFinalOutput != NO_OUTPUT
		if !reflect.DeepEqual(arc.NextFinalOutput, NO_OUTPUT) {
			err := f.outputs.(OutputsExt[T]).WriteFinalOutput(arc.NextFinalOutput, builder.bytes)
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
			numArcBytes := builder.bytes.GetPosition() - lastArcStart
			builder.numBytesPerArc[arcIdx] = int(numArcBytes)
			builder.numLabelBytesPerArc[arcIdx] = int(numLabelBytes)

			lastArcStart = builder.bytes.GetPosition()
			maxBytesPerArc = util.Max(maxBytesPerArc, int(numArcBytes))
			maxBytesPerArcWithoutLabel = util.Max(maxBytesPerArcWithoutLabel, int(numArcBytes-numLabelBytes))
		}
	}

	// TODO: try to avoid wasteful cases: disable doFixedLengthArcs in that case
	/*
	    *
	    * LUCENE-4682: what is a fair heuristic here?
	    * It could involve some of these:
	    * 1. how "busy" the node is: nodeIn.inputCount relative to frontier[0].inputCount?
	    * 2. how much binSearch saves over scan: nodeIn.numArcs
	    * 3. waste: numBytes vs numBytesExpanded
	    *
	    * the one below just looks at #3
	   if (doFixedLengthArcs) {
	     // rough heuristic: make this 1.25 "waste factor" a parameter to the phd ctor????
	     int numBytes = lastArcStart - startAddress;
	     int numBytesExpanded = maxBytesPerArc * nodeIn.numArcs;
	     if (numBytesExpanded > numBytes*1.25) {
	       doFixedLengthArcs = false;
	     }
	   }
	*/

	if doFixedLengthArcs {
		labelRange := nodeIn.arcs[nodeIn.numArcs-1].Label - nodeIn.arcs[0].Label + 1
		if f.shouldExpandNodeWithDirectAddressing(builder, nodeIn, maxBytesPerArc, maxBytesPerArcWithoutLabel, labelRange) {
			f.writeNodeForDirectAddressing(builder, nodeIn, startAddress, maxBytesPerArcWithoutLabel, labelRange)
			builder.directAddressingNodeCount++
		} else {
			f.writeNodeForBinarySearch(builder, nodeIn, startAddress, maxBytesPerArc)
			builder.binarySearchNodeCount++
		}
	}

	thisNodeAddress := builder.bytes.GetPosition() - 1
	builder.bytes.Reverse(int(startAddress), int(thisNodeAddress))
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

func (f *FST[T]) shouldExpandNodeWithDirectAddressing(builder *Builder[T], nodeIn *UnCompiledNode[T],
	numBytesPerArc, maxBytesPerArcWithoutLabel, labelRange int) bool {

	// Anticipate precisely the size of the encodings.
	sizeForBinarySearch := int(numBytesPerArc * nodeIn.numArcs)
	sizeForDirectAddressing := getNumPresenceBytes(labelRange) + builder.numLabelBytesPerArc[0] +
		maxBytesPerArcWithoutLabel*nodeIn.numArcs

	// Determine the allowed oversize compared to binary search.
	// This is defined by a parameter of FST Builder (default 1: no oversize).
	allowedOversize := int(float64(sizeForBinarySearch) * builder.directAddressingMaxOversizingFactor)
	expansionCost := sizeForDirectAddressing - allowedOversize

	// Select direct addressing if either:
	// - Direct addressing size is smaller than binary search.
	//   In this case, increment the credit by the reduced size (to use it later).
	// - Direct addressing size is larger than binary search, but the positive credit allows the oversizing.
	//   In this case, decrement the credit by the oversize.
	// In addition, do not try to oversize to a clearly too large node size
	// (this is the DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR parameter).
	if expansionCost <= 0 || (builder.directAddressingExpansionCredit >= int64(expansionCost) &&
		sizeForDirectAddressing <= int(float64(allowedOversize)*DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR)) {
		builder.directAddressingExpansionCredit -= int64(expansionCost)
		return true
	}
	return false
}

func getNumPresenceBytes(labelRange int) int {
	return (labelRange + 7) >> 3
}

func (f *FST[T]) writeNodeForBinarySearch(builder *Builder[T], nodeIn *UnCompiledNode[T],
	startAddress int64, maxBytesPerArc int) {

	builder.fixedLengthArcsBuffer.
		resetPosition().
		writeByte(ARCS_FOR_BINARY_SEARCH).
		writeVInt(nodeIn.numArcs).
		writeVInt(maxBytesPerArc)

	headerLen := builder.fixedLengthArcsBuffer.getPosition()

	// Expand the arcs in place, backwards.
	srcPos := int(builder.bytes.GetPosition())
	destPos := int(startAddress) + headerLen + nodeIn.numArcs*maxBytesPerArc
	if destPos > srcPos {
		builder.bytes.SkipBytes(destPos - srcPos)
		for arcIdx := nodeIn.numArcs - 1; arcIdx >= 0; arcIdx-- {
			destPos -= maxBytesPerArc
			arcLen := builder.numBytesPerArc[arcIdx]
			srcPos -= arcLen
			if srcPos != destPos {
				builder.bytes.CopyBytesToSelf(srcPos, destPos, arcLen)
			}
		}
	}

	builder.bytes.writeBytes(int(startAddress), builder.fixedLengthArcsBuffer.getBytes()[:headerLen])
}

func (f *FST[T]) writeNodeForDirectAddressing(builder *Builder[T], nodeIn *UnCompiledNode[T],
	startAddress int64, maxBytesPerArcWithoutLabel, labelRange int) {

	// Expand the arcs backwards in a buffer because we remove the labels.
	// So the obtained arcs might occupy less space. This is the reason why this
	// whole method is more complex.
	// Drop the label bytes since we can infer the label based on the arc index,
	// the presence bits, and the first label. Keep the first label.
	headerMaxLen := 11
	numPresenceBytes := getNumPresenceBytes(labelRange)
	srcPos := int(builder.bytes.GetPosition())
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
		writeByte(byte(ARCS_FOR_DIRECT_ADDRESSING)).
		writeVInt(labelRange).                // labelRange instead of numArcs.
		writeVInt(maxBytesPerArcWithoutLabel) // maxBytesPerArcWithoutLabel instead of maxBytesPerArc.
	headerLen := builder.fixedLengthArcsBuffer.getPosition()

	// Prepare the builder byte store. Enlarge or truncate if needed.
	nodeEnd := int(startAddress) + headerLen + numPresenceBytes + totalArcBytes
	currentPosition := int(builder.bytes.GetPosition())
	if nodeEnd >= currentPosition {
		builder.bytes.SkipBytes(nodeEnd - currentPosition)
	} else {
		builder.bytes.Truncate(nodeEnd)
	}

	// Write the header.
	writeOffset := int(startAddress)
	builder.bytes.writeBytes(int(writeOffset), builder.fixedLengthArcsBuffer.getBytes())
	writeOffset += headerLen

	// Write the presence bits
	f.writePresenceBits(builder, nodeIn, writeOffset, numPresenceBytes)
	writeOffset += numPresenceBytes

	// Write the first label and the arcs.
	builder.bytes.writeBytes(writeOffset, builder.fixedLengthArcsBuffer.getBytes()[bufferOffset:bufferOffset+totalArcBytes])
}

func (f *FST[T]) writePresenceBits(builder *Builder[T], nodeIn *UnCompiledNode[T], dest, numPresenceBytes int) {
	bytePos := dest
	presenceBits := byte(1) // The first arc is always present.
	presenceIndex := 0
	previousLabel := nodeIn.arcs[0].Label

	for arcIdx := 1; arcIdx < nodeIn.numArcs; arcIdx++ {
		label := nodeIn.arcs[arcIdx].Label
		presenceIndex += label - previousLabel

		for presenceIndex >= 8 {
			builder.bytes.writeByte(bytePos, presenceBits)
			bytePos++
			presenceBits = 0
			presenceIndex -= 8
		}

		// Set the bit at presenceIndex to flag that the corresponding arc is present.
		presenceBits |= 1 << presenceIndex
		previousLabel = label
	}

	builder.bytes.writeByte(bytePos, presenceBits)
}

func flag(flags, bits int) bool {
	return flags&bits != 0
}

// InputType Specifies allowed range of each int input label for this FST.
type InputType int

const (
	BYTE1 = InputType(iota)
	BYTE2
	BYTE4
)

// BytesReader Reads bytes stored in an FST.
type BytesReader interface {
	store.DataInput

	// GetPosition Get current read position.
	GetPosition() int64

	// SetPosition Set current read position.
	SetPosition(pos int64) error

	// Reversed Returns true if this reader uses reversed bytes under-the-hood.
	Reversed() bool
}
