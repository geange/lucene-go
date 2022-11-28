package fst

import (
	"errors"
	"fmt"
	"github.com/geange/lucene-go/codecs"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/math"
	"os"
)

var (
	DEFAULT_MAX_BLOCK_BITS = 30
)

type FST struct {
	inputType INPUT_TYPE

	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput any

	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB). If the FST is less than 1 GB then bytesArray is set instead.
	bytes *ByteStore

	fstStore Store

	startNode int64

	outputs Outputs
}

func NewFST(inputType INPUT_TYPE, outputs Outputs, bytesPageBits int) *FST {
	return &FST{
		inputType:   inputType,
		emptyOutput: nil,
		bytes:       NewByteStore(bytesPageBits),
		fstStore:    nil,
		startNode:   -1,
		outputs:     outputs,
	}
}

// NewFSTV1 Load a previously saved FST.
func NewFSTV1(metaIn, in store.DataInput, outputs Outputs) (*FST, error) {
	return NewFSTV2(metaIn, in, outputs, nil)
}

// NewFSTV2 Load a previously saved FST; maxBlockBits allows you to control the size of
// the byte[] pages used to hold the FST bytes.
func NewFSTV2(metaIn, in store.DataInput, outputs Outputs, fstStore Store) (*FST, error) {
	this := &FST{}

	this.bytes = nil
	this.fstStore = fstStore
	this.outputs = outputs

	// NOTE: only reads formats VERSION_START up to VERSION_CURRENT; we don't have
	// back-compat promise for FSTs (they are experimental), but we are sometimes able to offer it
	if _, err := codecs.CheckHeader(metaIn, FILE_FORMAT_NAME, VERSION_START, VERSION_CURRENT); err != nil {
		return nil, err
	}
	if b, err := metaIn.ReadByte(); err == nil && b == 1 {
		// accepts empty string
		// 1 KB blocks:
		emptyBytes := NewByteStore(10)
		numBytes, err := metaIn.ReadUvarint()
		if err != nil {
			return nil, err
		}
		if err := emptyBytes.CopyBytes(metaIn, int(numBytes)); err != nil {
			return nil, err
		}

		// De-serialize empty-string output:
		reader, err := emptyBytes.GetReverseReader()
		// NoOutputs uses 0 bytes when writing its output,
		// so we have to check here else BytesStore gets
		// angry:
		if numBytes > 0 {
			if err := reader.SetPosition(int64(numBytes - 1)); err != nil {
				return nil, err
			}
		}
		this.emptyOutput, err = outputs.ReadFinalOutput(reader)
		if err != nil {
			return nil, err
		}
	} else {
		this.emptyOutput = nil
	}
	t, err := metaIn.ReadByte()
	if err != nil {
		return nil, err
	}
	switch t {
	case 0:
		this.inputType = BYTE1
		break
	case 1:
		this.inputType = BYTE2
		break
	case 2:
		this.inputType = BYTE4
		break
	default:
		return nil, fmt.Errorf("invalid input type %d", in)
	}
	startNode, err := metaIn.ReadUvarint()
	if err != nil {
		return nil, err
	}
	this.startNode = int64(startNode)

	numBytes, err := metaIn.ReadUvarint()
	if err != nil {
		return nil, err
	}
	if err := this.fstStore.Init(in, int64(numBytes)); err != nil {
		return nil, err
	}

	return this, nil
}

func (f *FST) SetEmptyOutput(v any) error {
	if f.emptyOutput != nil {
		var err error
		f.emptyOutput, err = f.outputs.Merge(f.emptyOutput, v)
		if err != nil {
			return err
		}
	} else {
		f.emptyOutput = v
	}
	return nil
}

func (f *FST) Save(metaOut store.DataOutput, out store.DataOutput) error {
	if f.startNode == -1 {
		return errors.New("call finish first")
	}

	err := codecs.WriteHeader(metaOut, FILE_FORMAT_NAME, VERSION_CURRENT)
	if err != nil {
		return err
	}

	// TODO: really we should encode this as an arc, arriving
	// to the root node, instead of special casing here:
	if f.emptyOutput != nil {
		// Accepts empty string
		err := metaOut.WriteByte(1)
		if err != nil {
			return err
		}

		// Serialize empty-string output:
		ros := store.NewRAMOutputStream()
		err = f.outputs.WriteFinalOutput(f.emptyOutput, ros)
		if err != nil {
			return err
		}

		pointer := ros.GetFilePointer()
		emptyOutputBytes := make([]byte, pointer)

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
		err := metaOut.WriteByte(0)
		if err != nil {
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

	if err := metaOut.WriteUvarint(uint64(f.startNode)); err != nil {
		return err
	}

	if f.bytes != nil {
		numBytes := f.bytes.GetPosition()
		if err := metaOut.WriteUvarint(uint64(numBytes)); err != nil {
			return err
		}
		return f.bytes.WriteTo(out)
	}

	return f.fstStore.WriteTo(out)
}

func (f *FST) SaveToFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	out := store.NewOutputStreamDataOutput(file)
	return f.Save(out, out)
}

// Reads an automaton from a file.
func (f *FST) Read(path string, outputs Outputs) (*FST, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	in := store.NewInputStreamDataInput(file)
	fstStore, err := NewOnHeapFSTStore(DEFAULT_MAX_BLOCK_BITS)
	if err != nil {
		return nil, err
	}
	return NewFSTV2(in, in, outputs, fstStore)
}

func (f *FST) writeLabel(out store.DataOutput, v int) error {
	// TODO: assert v >= 0: "v=" + v;
	if f.inputType == BYTE1 {
		// TODO: assert v <= 255: "v=" + v;
		err := out.WriteByte(byte(v))
		if err != nil {
			return err
		}
	} else if f.inputType == BYTE2 {
		// TODO: assert v <= 65535: "v=" + v;
		err := out.WriteUint16(uint16(v))
		if err != nil {
			return err
		}
	} else {
		return out.WriteUvarint(uint64(v))
	}
	return nil
}

// ReadLabel Reads one BYTE1/2/4 label from the provided DataInput.
func (f *FST) ReadLabel(in store.DataInput) (int, error) {
	var v int
	switch f.inputType {
	case BYTE1:
		n, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		v = int(n & 0xFF)
		return v, nil
	case BYTE2:
		n, err := in.ReadUint16()
		if err != nil {
			return 0, err
		}
		v = int(n & 0xFFFF)
		return v, nil
	default:
		n, err := in.ReadUvarint()
		if err != nil {
			return 0, err
		}
		return int(n), nil
	}
}

// TargetHasArcs returns true if the node at this address has any outgoing arcs
func TargetHasArcs(arc *Arc) bool {
	return arc.Target() > 0
}

// AddNode serializes new node by appending its bytes to the end
// of the current byte[]
func (f *FST) AddNode(builder *Builder, nodeIn *UnCompiledNode) (int64, error) {
	//noOutput := f.outputs.GetNoOutput()

	if nodeIn.NumArcs() == 0 {
		if nodeIn.IsFinal {
			return FINAL_END_NODE, nil
		} else {
			return NON_FINAL_END_NODE, nil
		}
	}
	startAddress := builder.bytes.GetPosition()

	doFixedLengthArcs := f.shouldExpandNodeWithFixedLengthArcs(builder, nodeIn)
	if doFixedLengthArcs {
		if int64(len(builder.numBytesPerArc)) < nodeIn.NumArcs() {
			builder.numBytesPerArc = make([]int, util.Oversize(nodeIn.NumArcs(), int64(INTEGER_BYTES)))
			builder.numLabelBytesPerArc = make([]int64, len(builder.numBytesPerArc))
		}
	}

	builder.arcCount += nodeIn.NumArcs()

	lastArc := nodeIn.NumArcs() - 1

	lastArcStart := builder.bytes.GetPosition()
	maxBytesPerArc := 0
	maxBytesPerArcWithoutLabel := 0
	for arcIdx := 0; arcIdx < int(nodeIn.NumArcs()); arcIdx++ {
		arc := nodeIn.Arcs[arcIdx]
		target := arc.Target.(*CompiledNode)
		flags := 0

		if arcIdx == int(lastArc) {
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
			if arc.NextFinalOutput != nil {
				flags += BIT_ARC_HAS_FINAL_OUTPUT
			}
		} else {
			// TODO: assert arc.nextFinalOutput == noOutput;
		}

		targetHasArcs := target.node > 0

		if !targetHasArcs {
			flags += BIT_STOP_NODE
		}

		if arc.Output != nil {
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

		if arc.Output != nil {
			err := f.outputs.Write(arc.Output, builder.bytes)
			if err != nil {
				return 0, err
			}
		}

		if arc.NextFinalOutput != nil {
			//System.out.println("    write final output");
			err := f.outputs.WriteFinalOutput(arc.NextFinalOutput, builder.bytes)
			if err != nil {
				return 0, err
			}
		}

		if targetHasArcs && (flags&BIT_TARGET_NEXT) == 0 {
			// TODO: assert target.node > 0;
			//System.out.println("    write target");
			err := builder.bytes.WriteUvarint(uint64(target.node))
			if err != nil {
				return 0, err
			}
		}

		// just write the arcs "like normal" on first pass, but record how many bytes each one took
		// and max byte size:
		if doFixedLengthArcs {
			numArcBytes := int(builder.bytes.GetPosition() - lastArcStart)
			builder.numBytesPerArc[arcIdx] = numArcBytes
			builder.numLabelBytesPerArc[arcIdx] = numLabelBytes
			lastArcStart = builder.bytes.GetPosition()
			maxBytesPerArc = math.Max(maxBytesPerArc, numArcBytes)
			maxBytesPerArcWithoutLabel = math.Max(maxBytesPerArcWithoutLabel, numArcBytes-int(numLabelBytes))
			//System.out.println("    arcBytes=" + numArcBytes + " labelBytes=" + numLabelBytes);
		}
	}

	if doFixedLengthArcs {
		// TODO: assert maxBytesPerArc > 0;
		// 2nd pass just "expands" all arcs to take up a fixed byte size

		labelRange := nodeIn.Arcs[nodeIn.NumArcs()-1].Label - nodeIn.Arcs[0].Label + 1
		// TODO: assert labelRange > 0;
		if ok, err := f.shouldExpandNodeWithDirectAddressing(builder, nodeIn, int64(maxBytesPerArc), int64(maxBytesPerArcWithoutLabel), int64(labelRange)); ok && err == nil {
			err := f.writeNodeForDirectAddressing(builder, nodeIn, startAddress, int64(maxBytesPerArcWithoutLabel), int64(labelRange))
			if err != nil {
				return 0, err
			}
			builder.directAddressingNodeCount++
		} else {
			err := f.writeNodeForBinarySearch(builder, nodeIn, startAddress, int64(maxBytesPerArc))
			if err != nil {
				return 0, err
			}
			builder.binarySearchNodeCount++
		}
	}

	thisNodeAddress := builder.bytes.GetPosition() - 1
	err := builder.bytes.Reverse(startAddress, thisNodeAddress)
	if err != nil {
		return 0, err
	}
	builder.nodeCount++
	return thisNodeAddress, nil
}

// Returns whether the given node should be expanded with fixed length arcs. Nodes will be
// expanded depending on their depth (distance from the root node) and their number of arcs.
// Nodes with fixed length arcs use more space, because they encode all arcs with a fixed
// number of bytes, but they allow either binary search or direct addressing on the arcs
// (instead of linear scan) on lookup by arc label.
func (f *FST) shouldExpandNodeWithFixedLengthArcs(builder *Builder, node *UnCompiledNode) bool {
	return builder.allowFixedLengthArcs &&
		((node.Depth <= FIXED_LENGTH_ARC_SHALLOW_DEPTH && node.NumArcs() >= FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS) ||
			node.NumArcs() >= FIXED_LENGTH_ARC_DEEP_NUM_ARCS)
}

// Returns whether the given node should be expanded with direct addressing instead of binary search.
// Prefer direct addressing for performance if it does not oversize binary search byte size too much,
// so that the arcs can be directly addressed by label.
// See Also: Builder.getDirectAddressingMaxOversizingFactor()
func (f *FST) shouldExpandNodeWithDirectAddressing(builder *Builder, nodeIn *UnCompiledNode,
	numBytesPerArc, maxBytesPerArcWithoutLabel, labelRange int64) (bool, error) {

	// Anticipate precisely the size of the encodings.
	sizeForBinarySearch := numBytesPerArc * nodeIn.NumArcs()

	bytes, err := getNumPresenceBytes(labelRange)
	if err != nil {
		return false, err
	}
	sizeForDirectAddressing := bytes + builder.numLabelBytesPerArc[0] + maxBytesPerArcWithoutLabel*nodeIn.NumArcs()

	// Determine the allowed oversize compared to binary search.
	// This is defined by a parameter of FST Builder (default 1: no oversize).
	allowedOversize := int64(float64(sizeForBinarySearch) * builder.GetDirectAddressingMaxOversizingFactor())
	expansionCost := (sizeForDirectAddressing) - allowedOversize

	// Select direct addressing if either:
	// - Direct addressing size is smaller than binary search.
	//   In this case, increment the credit by the reduced size (to use it later).
	// - Direct addressing size is larger than binary search, but the positive credit allows the oversizing.
	//   In this case, decrement the credit by the oversize.
	// In addition, do not try to oversize to a clearly too large node size
	// (this is the DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR parameter).
	if expansionCost <= 0 || (builder.directAddressingExpansionCredit >= expansionCost &&
		sizeForDirectAddressing <= int64(float64(allowedOversize)*DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR)) {
		builder.directAddressingExpansionCredit -= int64(expansionCost)
		return true, nil
	}
	return false, nil
}

func (f *FST) writeNodeForBinarySearch(builder *Builder, nodeIn *UnCompiledNode,
	startAddress int64, maxBytesPerArc int64) error {
	// Build the header in a buffer.
	// It is a false/special arc which is in fact a node header with node flags followed by node metadata.
	fixedBuffer := builder.fixedLengthArcsBuffer
	err := fixedBuffer.resetPosition()
	if err != nil {
		return err
	}
	err = fixedBuffer.writeByte(ARCS_FOR_BINARY_SEARCH)
	if err != nil {
		return err
	}
	err = fixedBuffer.writeVInt(nodeIn.NumArcs())
	if err != nil {
		return err
	}
	err = fixedBuffer.writeVInt(maxBytesPerArc)
	if err != nil {
		return err
	}

	headerLen := builder.fixedLengthArcsBuffer.getPosition()

	// Expand the arcs in place, backwards.
	srcPos := builder.bytes.GetPosition()
	destPos := startAddress + headerLen + nodeIn.NumArcs()*maxBytesPerArc
	// TODO: assert destPos >= srcPos;
	if destPos > srcPos {
		err := builder.bytes.SkipBytes(destPos - srcPos)
		if err != nil {
			return err
		}
		for arcIdx := nodeIn.NumArcs() - 1; arcIdx >= 0; arcIdx-- {
			destPos -= int64(maxBytesPerArc)
			arcLen := builder.numBytesPerArc[arcIdx]
			srcPos -= int64(arcLen)
			if srcPos != destPos {
				// TODO: assert destPos > srcPos: "destPos=" + destPos + " srcPos=" + srcPos + " arcIdx=" + arcIdx + " maxBytesPerArc=" + maxBytesPerArc + " arcLen=" + arcLen + " nodeIn.numArcs=" + nodeIn.numArcs;
				err := builder.bytes.MoveBytes(srcPos, destPos, int64(arcLen))
				if err != nil {
					return err
				}
			}
		}
	}

	// Write the header.
	bytes := builder.fixedLengthArcsBuffer.GetBytes()
	return builder.bytes.WriteBytesAt(startAddress, bytes[0:headerLen])
}

func (f *FST) writeNodeForDirectAddressing(builder *Builder, nodeIn *UnCompiledNode,
	startAddress, maxBytesPerArcWithoutLabel, labelRange int64) error {

	// Expand the arcs backwards in a buffer because we remove the labels.
	// So the obtained arcs might occupy less space. This is the reason why this
	// whole method is more complex.
	// Drop the label bytes since we can infer the label based on the arc index,
	// the presence bits, and the first label. Keep the first label.
	headerMaxLen := int64(11)
	numPresenceBytes, err := getNumPresenceBytes(labelRange)
	if err != nil {
		return err
	}
	srcPos := builder.bytes.GetPosition()
	totalArcBytes := builder.numLabelBytesPerArc[0] + nodeIn.NumArcs()*maxBytesPerArcWithoutLabel
	bufferOffset := headerMaxLen + numPresenceBytes + totalArcBytes
	fixedBuffer := builder.fixedLengthArcsBuffer
	err = fixedBuffer.ensureCapacity(int(bufferOffset))
	if err != nil {
		return err
	}
	buffer := fixedBuffer.GetBytes()

	// Copy the arcs to the buffer, dropping all labels except first one.
	for arcIdx := nodeIn.NumArcs() - 1; arcIdx >= 0; arcIdx-- {
		bufferOffset -= maxBytesPerArcWithoutLabel
		srcArcLen := int64(builder.numBytesPerArc[arcIdx])
		srcPos -= srcArcLen
		labelLen := int64(builder.numLabelBytesPerArc[arcIdx])
		// Copy the flags.
		err := builder.bytes.CopyBytesToArray(srcPos, buffer[bufferOffset:bufferOffset+1])
		if err != nil {
			return err
		}
		// Skip the label, copy the remaining.
		remainingArcLen := srcArcLen - 1 - labelLen
		if remainingArcLen != 0 {
			err := builder.bytes.CopyBytesToArray(srcPos+1+labelLen, buffer[bufferOffset+1:bufferOffset+1+remainingArcLen])
			if err != nil {
				return err
			}
		}
		if arcIdx == 0 {
			// Copy the label of the first arc only.
			bufferOffset -= labelLen
			err := builder.bytes.CopyBytesToArray(srcPos+1, buffer[bufferOffset:bufferOffset+labelLen])
			if err != nil {
				return err
			}
		}
	}

	// TODO: assert bufferOffset == headerMaxLen + numPresenceBytes;

	// Build the header in the buffer.
	// It is a false/special arc which is in fact a node header with node flags followed by node metadata.
	//fixedBuffer := builder.fixedLengthArcsBuffer
	err = fixedBuffer.resetPosition()
	if err != nil {
		return err
	}
	err = fixedBuffer.writeByte(ARCS_FOR_DIRECT_ADDRESSING)
	if err != nil {
		return err
	}
	err = fixedBuffer.writeVInt(labelRange) // labelRange instead of numArcs.
	if err != nil {
		return err
	}
	err = fixedBuffer.writeVInt(maxBytesPerArcWithoutLabel) // maxBytesPerArcWithoutLabel instead of maxBytesPerArc.
	if err != nil {
		return err
	}

	headerLen := builder.fixedLengthArcsBuffer.getPosition()
	// Prepare the builder byte store. Enlarge or truncate if needed.
	nodeEnd := startAddress + headerLen + numPresenceBytes + totalArcBytes
	currentPosition := builder.bytes.GetPosition()
	if nodeEnd >= currentPosition {
		err := builder.bytes.SkipBytes(nodeEnd - currentPosition)
		if err != nil {
			return err
		}
	} else {
		err := builder.bytes.Truncate(nodeEnd)
		if err != nil {
			return err
		}
	}
	// TODO: assert builder.bytes.getPosition() == nodeEnd

	// Write the header.
	writeOffset := startAddress
	buff := builder.fixedLengthArcsBuffer.GetBytes()
	err = builder.bytes.WriteBytesAt(writeOffset, buff[0:headerLen])
	if err != nil {
		return err
	}
	writeOffset += headerLen

	// Write the presence bits
	err = f.writePresenceBits(builder, nodeIn, writeOffset, numPresenceBytes)
	if err != nil {
		return err
	}
	writeOffset += numPresenceBytes

	// Write the first label and the arcs.
	return builder.bytes.WriteBytesAt(writeOffset, buff[bufferOffset:bufferOffset+totalArcBytes])
}

func (f *FST) writePresenceBits(builder *Builder, nodeIn *UnCompiledNode, dest, numPresenceBytes int64) error {
	bytePos := dest
	presenceBits := 1 // The first arc is always present.
	presenceIndex := 0
	previousLabel := nodeIn.Arcs[0].Label
	for arcIdx := 1; arcIdx < int(nodeIn.NumArcs()); arcIdx++ {
		label := nodeIn.Arcs[arcIdx].Label
		// TODO: assert label > previousLabel;
		presenceIndex += label - previousLabel
		for presenceIndex >= BYTE_SIZE {
			err := builder.bytes.WriteByteAt(bytePos, byte(presenceBits))
			if err != nil {
				return err
			}
			bytePos++
			presenceBits = 0
			presenceIndex -= BYTE_SIZE
		}
		// Set the bit at presenceIndex to flag that the corresponding arc is present.
		presenceBits |= 1 << presenceIndex
		previousLabel = label
	}
	// TODO:assert presenceIndex == (nodeIn.arcs[nodeIn.numArcs - 1].label - nodeIn.arcs[0].label) % 8;
	// TODO:assert presenceBits != 0; // The last byte is not 0.
	// TODO:assert (presenceBits & (1 << presenceIndex)) != 0; // The last arc is always present.
	err := builder.bytes.WriteByteAt(bytePos, byte(presenceBits))
	if err != nil {
		return err
	}
	bytePos++
	// TODO:assert bytePos - dest == numPresenceBytes;
	return nil
}

// Gets the number of bytes required to flag the presence of each arc in the given label range, one bit per arc.
func getNumPresenceBytes(labelRange int64) (int64, error) {
	err := assert(labelRange >= 0)
	if err != nil {
		return 0, err
	}
	return (labelRange + 7) >> 3, nil
}

// Reads the presence bits of a direct-addressing node. Actually we don't read them here,
// we just keep the pointer to the bit-table start and we skip them.
func (f *FST) readPresenceBytes(arc *Arc, in BytesReader) error {

	// TODO: assert arc.bytesPerArc() > 0;
	// TODO: assert arc.nodeFlags() == ARCS_FOR_DIRECT_ADDRESSING;
	arc.bitTableStart = in.GetPosition()

	bytes, err := getNumPresenceBytes(arc.NumArcs())
	if err != nil {
		return err
	}
	return in.SkipBytes(int(bytes))
}

// GetFirstArc Fills virtual 'start' arc, ie, an empty incoming arc to the FST's start node
func (f *FST) GetFirstArc(arc *Arc) (*Arc, error) {
	noOutput := f.outputs.GetNoOutput()

	if f.emptyOutput != nil {
		arc.flags = BIT_FINAL_ARC | BIT_LAST_ARC
		arc.nextFinalOutput = f.emptyOutput
		if f.emptyOutput != noOutput {
			arc.flags = arc.Flags() | BIT_ARC_HAS_FINAL_OUTPUT
		}
	} else {
		arc.flags = BIT_LAST_ARC
		arc.nextFinalOutput = nil
	}
	arc.output = noOutput

	// If there are no nodes, ie, the FST only accepts the
	// empty string, then startNode is 0
	arc.target = f.startNode
	return arc, nil
}

// Follows the follow arc and reads the last arc of its target; this changes the provided
// arc (2nd arg) in-place and returns it.
// Returns: Returns the second argument (arc).
func (f *FST) readLastTargetArc(follow, arc *Arc, in BytesReader) (*Arc, error) {
	if !TargetHasArcs(follow) {
		//System.out.println("  end node");
		// TODO: assert follow.isFinal();
		arc.label = END_LABEL
		arc.target = FINAL_END_NODE
		arc.output = follow.NextFinalOutput()
		arc.flags = BIT_LAST_ARC
		arc.nodeFlags = arc.flags
		return arc, nil
	}
	in.SetPosition(follow.Target())
	flags, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	arc.nodeFlags = flags

	if flags == ARCS_FOR_BINARY_SEARCH || flags == ARCS_FOR_DIRECT_ADDRESSING {
		// Special arc which is actually a node header for fixed length arcs.
		// Jump straight to end to find the last arc.
		numArcs, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.numArcs = int64(numArcs)

		bytesPerArc, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc = int(bytesPerArc)
		//System.out.println("  array numArcs=" + arc.numArcs + " bpa=" + arc.bytesPerArc);
		if flags == ARCS_FOR_DIRECT_ADDRESSING {
			if err := f.readPresenceBytes(arc, in); err != nil {
				return nil, err
			}
			arc.firstLabel, err = f.ReadLabel(in)
			if err != nil {
				return nil, err
			}
			arc.posArcsStart = in.GetPosition()
			if _, err := f.ReadLastArcByDirectAddressing(arc, in); err != nil {
				return nil, err
			}
		} else {
			arc.arcIdx = int(arc.NumArcs() - 2)
			arc.posArcsStart = in.GetPosition()
			if _, err := f.ReadNextRealArc(arc, in); err != nil {
				return nil, err
			}
		}
	} else {
		arc.flags = flags
		// non-array: linear scan
		arc.bytesPerArc = 0
		//System.out.println("  scan");
		for !arc.IsLast() {
			// skip this arc:
			if _, err := f.ReadLabel(in); err != nil {
				return nil, err
			}
			if arc.flag(BIT_ARC_HAS_OUTPUT) {
				if err := f.outputs.SkipOutput(in); err != nil {
					return nil, err
				}
			}
			if arc.flag(BIT_ARC_HAS_FINAL_OUTPUT) {
				if err := f.outputs.SkipFinalOutput(in); err != nil {
					return nil, err
				}
			}
			if arc.flag(BIT_STOP_NODE) {
			} else if arc.flag(BIT_TARGET_NEXT) {
			} else {
				if _, err := f.readUnpackedNodeTarget(in); err != nil {
					return nil, err
				}
			}
			arc.flags, err = in.ReadByte()
			if err != nil {
				return nil, err
			}
		}
		// Undo the byte flags we read:
		if err := in.SkipBytes(-1); err != nil {
			return nil, err
		}

		arc.nextArc = in.GetPosition()
		if _, err := f.ReadNextRealArc(arc, in); err != nil {
			return nil, err
		}

	}
	// TODO: assert arc.isLast();
	return arc, nil

}

func (f *FST) readUnpackedNodeTarget(in BytesReader) (int64, error) {
	num, err := in.ReadUvarint()
	if err != nil {
		return 0, err
	}
	return int64(num), nil
}

// ReadFirstTargetArc Follow the follow arc and read the first arc of its target; this changes
// fthe provided arc (2nd arg) in-place and returns it.
// Returns: Returns the second argument (arc).
func (f *FST) ReadFirstTargetArc(follow, arc *Arc, in BytesReader) (*Arc, error) {
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
		return arc, nil
	} else {
		return f.ReadFirstRealTargetArc(follow.Target(), arc, in)
	}
}

func (f *FST) ReadFirstRealTargetArc(nodeAddress int64, arc *Arc, in BytesReader) (*Arc, error) {
	err := in.SetPosition(nodeAddress)
	if err != nil {
		return nil, err
	}

	b, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	var flags byte
	flags, arc.nodeFlags = b, b

	if flags == ARCS_FOR_BINARY_SEARCH || flags == ARCS_FOR_DIRECT_ADDRESSING {
		num1, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.numArcs = int64(num1)

		num2, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc = int(num2)

		arc.arcIdx = -1

		if flags == ARCS_FOR_DIRECT_ADDRESSING {
			err := f.readPresenceBytes(arc, in)
			if err != nil {
				return nil, err
			}

			label, err := f.ReadLabel(in)
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

	return f.ReadNextRealArc(arc, in)
}

// Returns whether arc's target points to a node in expanded format (fixed length arcs).
func isExpandedTarget(follow *Arc, in BytesReader) (bool, error) {
	if !TargetHasArcs(follow) {
		return false, nil
	} else {
		if err := in.SetPosition(follow.Target()); err != nil {
			return false, err
		}
		flags, err := in.ReadByte()
		if err != nil {
			return false, err
		}
		return flags == ARCS_FOR_BINARY_SEARCH || flags == ARCS_FOR_DIRECT_ADDRESSING, nil
	}
}

// ReadNextArc In-place read; returns the arc.
func (f *FST) ReadNextArc(arc *Arc, in BytesReader) (*Arc, error) {
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
func (f *FST) readNextArcLabel(arc *Arc, in BytesReader) (int, error) {
	// TODO: assert !arc.isLast();

	if arc.Label() == END_LABEL {
		//System.out.println("    nextArc fake " + arc.nextArc);
		// Next arc is the first arc of a node.
		// Position to read the first arc label.

		if err := in.SetPosition(arc.NextArc()); err != nil {
			return 0, err
		}
		flags, err := in.ReadByte()
		if err != nil {
			return 0, err
		}
		if flags == ARCS_FOR_BINARY_SEARCH || flags == ARCS_FOR_DIRECT_ADDRESSING {

			// Special arc which is actually a node header for fixed length arcs.
			numArcs, err := in.ReadUvarint()
			if err != nil {
				return 0, err
			}
			if _, err := in.ReadUvarint(); err != nil {
				return 0, err
			} // Skip bytesPerArc.
			if flags == ARCS_FOR_BINARY_SEARCH {
				if _, err := in.ReadByte(); err != nil {
					return 0, err
				} // Skip arc flags.
			} else {
				bytes, err := getNumPresenceBytes(int64(numArcs))
				if err != nil {
					return 0, err
				}
				if err := in.SkipBytes(int(bytes)); err != nil {
					return 0, err
				}
			}
		}
	} else {
		if arc.BytesPerArc() != 0 {
			// Arcs have fixed length.
			if arc.NodeFlags() == ARCS_FOR_BINARY_SEARCH {
				// Point to next arc, -1 to skip arc flags.
				if err := in.SetPosition(arc.PosArcsStart() - (1+int64(arc.ArcIdx()))*int64(arc.BytesPerArc()) - 1); err != nil {
					return 0, err
				}
			} else {
				// TODO: assert arc.nodeFlags() == ARCS_FOR_DIRECT_ADDRESSING;
				// Direct addressing node. The label is not stored but rather inferred
				// based on first label and arc index in the range.
				// TODO: assert BitTable.assertIsValid(arc, in);
				// TODO: assert BitTable.isBitSet(arc.arcIdx(), arc, in);
				nextIndex, err := BitTable.nextBitSet(arc.ArcIdx(), arc, in)
				if err != nil {
					return 0, err
				}
				// TODO: assert nextIndex != -1;
				return arc.FirstLabel() + nextIndex, nil
			}
		} else {
			// Arcs have variable length.
			//System.out.println("    nextArc real list");
			// Position to next arc, -1 to skip flags.
			if err := in.SetPosition(arc.NextArc() - 1); err != nil {
				return 0, err
			}
		}
	}
	return f.ReadLabel(in)
}

func (f *FST) ReadArcByIndex(arc *Arc, in BytesReader, idx int) (*Arc, error) {
	// TODO: assert arc.bytesPerArc() > 0;
	// TODO: assert arc.nodeFlags() == ARCS_FOR_BINARY_SEARCH;
	// TODO: assert idx >= 0 && idx < arc.numArcs();
	if err := in.SetPosition(arc.PosArcsStart() - int64(idx*arc.BytesPerArc())); err != nil {
		return nil, err
	}
	arc.arcIdx = idx
	var err error
	arc.flags, err = in.ReadByte()
	if err != nil {
		return nil, err
	}
	return f.readArc(arc, in)
}

// ReadArcByDirectAddressing Reads a present direct addressing node arc, with the provided index in the label range.
// Params: 	rangeIndex – The index of the arc in the label range. It must be present.
//
//	The real arc offset is computed based on the presence bits of the direct addressing node.
func (f *FST) ReadArcByDirectAddressing(arc *Arc, in BytesReader, rangeIndex int) (*Arc, error) {
	// TODO: assert BitTable.assertIsValid(arc, in);
	// TODO: assert rangeIndex >= 0 && rangeIndex < arc.numArcs();
	// TODO: assert BitTable.isBitSet(rangeIndex, arc, in);
	presenceIndex, err := BitTable.countBitsUpTo(rangeIndex, arc, in)
	if err != nil {
		return nil, err
	}
	return f.readArcByDirectAddressingV1(arc, in, rangeIndex, presenceIndex)
}

// Reads a present direct addressing node arc, with the provided index in the label range and its corresponding presence index (which is the count of presence bits before it).
func (f *FST) readArcByDirectAddressingV1(arc *Arc, in BytesReader, rangeIndex, presenceIndex int) (*Arc, error) {
	if err := in.SetPosition(arc.PosArcsStart() - int64(presenceIndex*arc.BytesPerArc())); err != nil {
		return nil, err
	}
	arc.arcIdx = rangeIndex
	arc.presenceIndex = presenceIndex

	flags, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	arc.flags = flags

	return f.readArc(arc, in)
}

// ReadLastArcByDirectAddressing Reads the last arc of a direct addressing node.
// This method is equivalent to call readArcByDirectAddressing(FST.Arc, FST.BytesReader, int)
// with rangeIndex equal to arc.numArcs() - 1, but it is faster.
func (f *FST) ReadLastArcByDirectAddressing(arc *Arc, in BytesReader) (*Arc, error) {
	// TODO: assert BitTable.assertIsValid(arc, in);
	presenceIndex, err := BitTable.countBits(arc, in)
	if err != nil {
		return nil, err
	}
	presenceIndex -= 1
	return f.readArcByDirectAddressingV1(arc, in, int(arc.NumArcs()-1), int(presenceIndex))
}

// ReadNextRealArc Never returns null, but you should never call this if arc.isLast() is true.
func (f *FST) ReadNextRealArc(arc *Arc, in BytesReader) (*Arc, error) {
	// TODO: can't assert this because we call from readFirstArc
	// assert !flag(arc.flags, BIT_LAST_ARC);

	switch arc.NodeFlags() {
	case ARCS_FOR_BINARY_SEARCH:
		// TODO: assert arc.bytesPerArc() > 0;
		arc.arcIdx++
		// TODO: assert arc.arcIdx() >= 0 && arc.arcIdx() < arc.numArcs()
		err := in.SetPosition(arc.PosArcsStart() - int64(arc.ArcIdx()*arc.BytesPerArc()))
		if err != nil {
			return nil, err
		}

		flags, err := in.ReadByte()
		if err != nil {
			return nil, err
		}
		arc.flags = flags

	case ARCS_FOR_DIRECT_ADDRESSING:
		// TODO: assert BitTable.assertIsValid(arc, in);
		// TODO: assert arc.arcIdx() == -1 || BitTable.isBitSet(arc.arcIdx(), arc, in);
		nextIndex, err := BitTable.nextBitSet(arc.ArcIdx(), arc, in)
		if err != nil {
			return nil, err
		}
		return f.readArcByDirectAddressingV1(arc, in, nextIndex, arc.presenceIndex+1)

	default:
		if arc.BytesPerArc() != 0 {
			return nil, fmt.Errorf("arc.BytesPerArc() != 0; arc.BytesPerArc() is %d", arc.BytesPerArc())
		}

		err := in.SetPosition(arc.NextArc())
		if err != nil {
			return nil, err
		}

		flags, err := in.ReadByte()
		if err != nil {
			return nil, err
		}
		arc.flags = flags
	}
	return f.readArc(arc, in)
}

func (f *FST) readArc(arc *Arc, in BytesReader) (*Arc, error) {
	if arc.NodeFlags() == ARCS_FOR_DIRECT_ADDRESSING {
		arc.label = arc.FirstLabel() + arc.ArcIdx()
	} else {
		label, err := f.ReadLabel(in)
		if err != nil {
			return nil, err
		}
		arc.label = label
	}

	if arc.flag(BIT_ARC_HAS_OUTPUT) {
		output, err := f.outputs.Read(in)
		if err != nil {
			return nil, err
		}
		arc.output = output
	} else {
		arc.output = nil
	}

	if arc.flag(BIT_ARC_HAS_FINAL_OUTPUT) {
		output, err := f.outputs.ReadFinalOutput(in)
		if err != nil {
			return nil, err
		}
		arc.nextFinalOutput = output
	} else {
		arc.nextFinalOutput = nil
	}

	if arc.flag(BIT_STOP_NODE) {
		if arc.flag(BIT_FINAL_ARC) {
			arc.target = FINAL_END_NODE
		} else {
			arc.target = NON_FINAL_END_NODE
		}
		arc.nextArc = in.GetPosition() // Only useful for list.
	} else if arc.flag(BIT_TARGET_NEXT) {
		arc.nextArc = in.GetPosition() // Only useful for list.
		// TODO: would be nice to make this lazy -- maybe
		// caller doesn't need the target and is scanning arcs...
		if !arc.flag(BIT_LAST_ARC) {
			if arc.BytesPerArc() == 0 {
				// must scan
				err := f.seekToNextNode(in)
				if err != nil {
					return nil, err
				}
			} else {
				var numArcs int
				if arc.nodeFlags == ARCS_FOR_DIRECT_ADDRESSING {
					bits, err := BitTable.countBits(arc, in)
					if err != nil {
						return nil, err
					}
					numArcs = int(bits)
				} else {
					numArcs = int(arc.NumArcs())
				}
				err := in.SetPosition(arc.PosArcsStart() - int64(arc.BytesPerArc()*numArcs))
				if err != nil {
					return nil, err
				}
			}
		}
		arc.target = in.GetPosition()
	} else {
		target, err := f.readUnpackedNodeTarget(in)
		if err != nil {
			return nil, err
		}
		arc.target = target
		arc.nextArc = in.GetPosition()
	}
	return arc, nil
}

func readEndArc(follow, arc *Arc) *Arc {
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
func (f *FST) FindTargetArc(labelToMatch int, follow, arc *Arc, in BytesReader) (*Arc, error) {

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

	if !TargetHasArcs(follow) {
		return nil, nil
	}

	if err := in.SetPosition(follow.Target()); err != nil {
		return nil, err
	}

	// System.out.println("fta label=" + (char) labelToMatch);
	flags, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	arc.nodeFlags = flags
	if flags == ARCS_FOR_DIRECT_ADDRESSING {
		numArcs, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.numArcs = int64(numArcs) // This is in fact the label range.

		bytesPerArc, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc = int(bytesPerArc)
		if err := f.readPresenceBytes(arc, in); err != nil {
			return nil, err
		}
		arc.firstLabel, err = f.ReadLabel(in)
		if err != nil {
			return nil, err
		}
		arc.posArcsStart = in.GetPosition()

		arcIndex := labelToMatch - arc.FirstLabel()
		if arcIndex < 0 || arcIndex >= int(arc.NumArcs()) {
			return nil, nil // Before or after label range.
		}

		if ok, err := BitTable.isBitSet(arcIndex, arc, in); err != nil {
			return nil, err
		} else if !ok {
			return nil, nil // Arc missing in the range.
		}

		return f.ReadArcByDirectAddressing(arc, in, arcIndex)
	} else if flags == ARCS_FOR_BINARY_SEARCH {
		numArcs, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.numArcs = int64(numArcs)

		bytesPerArc, err := in.ReadUvarint()
		if err != nil {
			return nil, err
		}
		arc.bytesPerArc = int(bytesPerArc)
		arc.posArcsStart = in.GetPosition()

		// Array is sparse; do binary search:
		low := 0
		high := int(arc.NumArcs() - 1)
		for low <= high {
			//System.out.println("    cycle");
			mid := (low + high) >> 1
			// +1 to skip over flags
			if err := in.SetPosition(arc.PosArcsStart() - int64(arc.BytesPerArc()*mid+1)); err != nil {
				return nil, err
			}
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
	if _, err := f.ReadFirstRealTargetArc(follow.Target(), arc, in); err != nil {
		return nil, err
	}

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
			if _, err := f.ReadNextRealArc(arc, in); err != nil {
				return nil, err
			}
		}
	}
}

func (f *FST) seekToNextNode(in BytesReader) error {
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

const (
	INTEGER_BYTES = 4
)

func (f *FST) Finish(newStartNode int64) error {
	// TODO: assert newStartNode <= bytes.getPosition();
	if f.startNode != -1 {
		return errors.New("already finished")
	}
	if newStartNode == FINAL_END_NODE && f.emptyOutput != nil {
		newStartNode = 0
	}
	f.startNode = newStartNode
	return f.bytes.Finish()
}

func flag(flags, bit int) bool {
	return flags&bit != 0
}
