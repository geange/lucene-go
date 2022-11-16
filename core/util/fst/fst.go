package fst

import (
	"fmt"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/math"
)

type FST struct {
	inputType INPUT_TYPE

	// if non-null, this FST accepts the empty string and
	// produces this output
	emptyOutput any

	// A BytesStore, used during building, or during reading when the FST is very large (more than 1 GB). If the FST is less than 1 GB then bytesArray is set instead.
	bytes *ByteStore

	fstStore FSTStore

	startNode int64

	outputs Outputs
}

func NewFST(inputType INPUT_TYPE, outputs Outputs, bytesPageBits int) *FST {
	return &FST{
		inputType:   inputType,
		emptyOutput: nil,
		bytes:       NewByteStore(bytesPageBits),
		fstStore:    nil,
		startNode:   0,
		outputs:     outputs,
	}
}

func (f *FST) ReadFirstRealTargetArc(nodeAddress int64, arc *FSTArc, in BytesReader) (*FSTArc, error) {
	in.SetPosition(nodeAddress)

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

func (f *FST) readPresenceBytes(arc *FSTArc, in BytesReader) error {

	// TODO: assert arc.bytesPerArc() > 0;
	// TODO: assert arc.nodeFlags() == ARCS_FOR_DIRECT_ADDRESSING;
	arc.bitTableStart = in.GetPosition()

	bytes, err := getNumPresenceBytes(arc.NumArcs())
	if err != nil {
		return err
	}
	return in.SkipBytes(int(bytes))
}

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

// ReadNextRealArc Never returns null, but you should never call this if arc.isLast() is true.
func (f *FST) ReadNextRealArc(arc *FSTArc, in BytesReader) (*FSTArc, error) {
	// TODO: can't assert this because we call from readFirstArc
	// assert !flag(arc.flags, BIT_LAST_ARC);

	switch arc.NodeFlags() {
	case ARCS_FOR_BINARY_SEARCH:
		// TODO: assert arc.bytesPerArc() > 0;
		arc.arcIdx++
		// TODO: assert arc.arcIdx() >= 0 && arc.arcIdx() < arc.numArcs()
		in.SetPosition(arc.PosArcsStart() - int64(arc.ArcIdx()*arc.BytesPerArc()))

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
		return f.readArcByDirectAddressing(arc, in, nextIndex, arc.presenceIndex+1)

	default:
		if arc.BytesPerArc() != 0 {
			return nil, fmt.Errorf("arc.BytesPerArc() != 0; arc.BytesPerArc() is %d", arc.BytesPerArc())
		}

		in.SetPosition(arc.NextArc())

		flags, err := in.ReadByte()
		if err != nil {
			return nil, err
		}
		arc.flags = flags
	}
	return f.readArc(arc, in)
}

func (f *FST) readArc(arc *FSTArc, in BytesReader) (*FSTArc, error) {
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
		arc.output = f.outputs.GetNoOutput()
	}

	if arc.flag(BIT_ARC_HAS_FINAL_OUTPUT) {
		output, err := f.outputs.ReadFinalOutput(in)
		if err != nil {
			return nil, err
		}
		arc.nextFinalOutput = output
	} else {
		arc.nextFinalOutput = f.outputs.GetNoOutput()
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
				in.SetPosition(arc.PosArcsStart() - int64(arc.BytesPerArc()*numArcs))
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

func (f *FST) readUnpackedNodeTarget(in BytesReader) (int64, error) {
	num, err := in.ReadUvarint()
	if err != nil {
		return 0, err
	}
	return int64(num), nil
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

func (f *FST) readArcByDirectAddressing(arc *FSTArc, in BytesReader, rangeIndex, presenceIndex int) (*FSTArc, error) {
	in.SetPosition(arc.PosArcsStart() - int64(presenceIndex*arc.BytesPerArc()))
	arc.arcIdx = rangeIndex
	arc.presenceIndex = presenceIndex

	flags, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	arc.flags = flags

	return f.readArc(arc, in)
}

// AddNode serializes new node by appending its bytes to the end
// of the current byte[]
func (f *FST) AddNode(builder *Builder, nodeIn *UnCompiledNode) (int64, error) {
	NO_OUTPUT := f.outputs.GetNoOutput()

	if nodeIn.NumArcs == 0 {
		if nodeIn.IsFinal {
			return FINAL_END_NODE, nil
		} else {
			return NON_FINAL_END_NODE, nil
		}
	}
	startAddress := builder.bytes.GetPosition()

	doFixedLengthArcs := f.shouldExpandNodeWithFixedLengthArcs(builder, nodeIn)
	if doFixedLengthArcs {
		if int64(len(builder.numBytesPerArc)) < nodeIn.NumArcs {
			builder.numBytesPerArc = make([]int, util.Oversize(nodeIn.NumArcs, int64(INTEGER_BYTES)))
			builder.numLabelBytesPerArc = make([]int64, len(builder.numBytesPerArc))
		}
	}

	builder.arcCount += int64(nodeIn.NumArcs)

	lastArc := nodeIn.NumArcs - 1

	lastArcStart := builder.bytes.GetPosition()
	maxBytesPerArc := 0
	maxBytesPerArcWithoutLabel := 0
	for arcIdx := 0; arcIdx < int(nodeIn.NumArcs); arcIdx++ {
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
			if arc.NextFinalOutput != NO_OUTPUT {
				flags += BIT_ARC_HAS_FINAL_OUTPUT
			}
		} else {
			// TODO: assert arc.nextFinalOutput == NO_OUTPUT;
		}

		targetHasArcs := target.node > 0

		if !targetHasArcs {
			flags += BIT_STOP_NODE
		}

		if arc.Output != NO_OUTPUT {
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

		if arc.Output != NO_OUTPUT {
			err := f.outputs.Write(arc.Output, builder.bytes)
			if err != nil {
				return 0, err
			}
		}

		if arc.NextFinalOutput != NO_OUTPUT {
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

		labelRange := nodeIn.Arcs[nodeIn.NumArcs-1].Label - nodeIn.Arcs[0].Label + 1
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

func (f *FST) shouldExpandNodeWithDirectAddressing(builder *Builder, nodeIn *UnCompiledNode,
	numBytesPerArc, maxBytesPerArcWithoutLabel, labelRange int64) (bool, error) {

	// Anticipate precisely the size of the encodings.
	sizeForBinarySearch := numBytesPerArc * nodeIn.NumArcs

	bytes, err := getNumPresenceBytes(labelRange)
	if err != nil {
		return false, err
	}
	sizeForDirectAddressing := bytes + builder.numLabelBytesPerArc[0] + maxBytesPerArcWithoutLabel*nodeIn.NumArcs

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
	totalArcBytes := builder.numLabelBytesPerArc[0] + nodeIn.NumArcs*maxBytesPerArcWithoutLabel
	bufferOffset := headerMaxLen + numPresenceBytes + totalArcBytes
	fixedBuffer := builder.fixedLengthArcsBuffer
	err = fixedBuffer.ensureCapacity(int(bufferOffset))
	if err != nil {
		return err
	}
	buffer := fixedBuffer.GetBytes()

	// Copy the arcs to the buffer, dropping all labels except first one.
	for arcIdx := nodeIn.NumArcs - 1; arcIdx >= 0; arcIdx-- {
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
	for arcIdx := 1; arcIdx < int(nodeIn.NumArcs); arcIdx++ {
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
	err = fixedBuffer.writeVInt(nodeIn.NumArcs)
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
	destPos := startAddress + int64(headerLen) + nodeIn.NumArcs*maxBytesPerArc
	// TODO: assert destPos >= srcPos;
	if destPos > srcPos {
		err := builder.bytes.SkipBytes(destPos - srcPos)
		if err != nil {
			return err
		}
		for arcIdx := nodeIn.NumArcs - 1; arcIdx >= 0; arcIdx-- {
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

const (
	INTEGER_BYTES = 4
)

func (f *FST) shouldExpandNodeWithFixedLengthArcs(builder *Builder, node *UnCompiledNode) bool {
	return builder.allowFixedLengthArcs &&
		((node.Depth <= FIXED_LENGTH_ARC_SHALLOW_DEPTH && node.NumArcs >= FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS) ||
			node.NumArcs >= FIXED_LENGTH_ARC_DEEP_NUM_ARCS)
}

func flag(flags, bit int) bool {
	return flags&bit != 0
}

func getNumPresenceBytes(labelRange int64) (int64, error) {
	err := assert(labelRange >= 0)
	if err != nil {
		return 0, err
	}
	return (labelRange + 7) >> 3, nil
}
