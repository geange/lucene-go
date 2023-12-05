package fst

// Returns whether arc's target points to a node in expanded format (fixed length arcs).
func isExpandedTarget[T any](follow *Arc[T], in BytesReader) (bool, error) {
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

// Returns whether the given node should be expanded with fixed length arcs. Nodes will be
// expanded depending on their depth (distance from the root node) and their number of arcs.
// Nodes with fixed length arcs use more space, because they encode all arcs with a fixed
// number of bytes, but they allow either binary search or direct addressing on the arcs
// (instead of linear scan) on lookup by arc label.
func (f *FST[T]) shouldExpandNodeWithFixedLengthArcs(builder *Builder[T], node *UnCompiledNode[T]) bool {
	return builder.allowFixedLengthArcs &&
		((node.Depth <= FIXED_LENGTH_ARC_SHALLOW_DEPTH &&
			node.NumArcs() >= FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS) ||
			node.NumArcs() >= FIXED_LENGTH_ARC_DEEP_NUM_ARCS)
}

// Returns whether the given node should be expanded with direct addressing instead of binary search.
// Prefer direct addressing for performance if it does not oversize binary search byte size too much,
// so that the arcs can be directly addressed by label.
// See Also: Builder.getDirectAddressingMaxOversizingFactor()
func (f *FST[T]) shouldExpandNodeWithDirectAddressing(builder *Builder[T], nodeIn *UnCompiledNode[T],
	numBytesPerArc, maxBytesPerArcWithoutLabel, labelRange int64) (bool, error) {

	// Anticipate precisely the size of the encodings.
	sizeForBinarySearch := numBytesPerArc * nodeIn.NumArcs()

	bytes, err := getNumPresenceBytes(labelRange)
	if err != nil {
		return false, err
	}
	sizeForDirectAddressing := bytes + builder.numLabelBytesPerArc[0] + maxBytesPerArcWithoutLabel*nodeIn.NumArcs()

	// Determine the allowed oversize compared to binary search.
	// This is defined by a parameter of Fst Builder (default 1: no oversize).
	allowedOversize := int64(float64(sizeForBinarySearch) * builder.GetDirectAddressingMaxOversizingFactor())
	expansionCost := (sizeForDirectAddressing) - allowedOversize

	// SelectK direct addressing if either:
	// - Direct addressing size is smaller than binary search.
	//   In this case, increment the credit by the reduced size (to use it later).
	// - Direct addressing size is larger than binary search, but the positive credit allows the oversizing.
	//   In this case, decrement the credit by the oversize.
	// In addition, do not try to oversize to a clearly too large node size
	// (this is the DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR parameter).
	if expansionCost <= 0 || (builder.directAddressingExpansionCredit >= expansionCost &&
		sizeForDirectAddressing <= int64(float64(allowedOversize)*DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR)) {
		builder.directAddressingExpansionCredit -= expansionCost
		return true, nil
	}
	return false, nil
}

func (f *FST[T]) writeNodeForBinarySearch(builder *Builder[T], nodeIn *UnCompiledNode[T],
	startAddress int64, maxBytesPerArc int64) error {
	// Build the header in a buffer.
	// It is a false/special arc which is in fact a node header with node flags followed by node metadata.
	fixedBuffer := builder.fixedLengthArcsBuffer
	if err := fixedBuffer.resetPosition(); err != nil {
		return err
	}
	if err := fixedBuffer.writeByte(ARCS_FOR_BINARY_SEARCH); err != nil {
		return err
	}
	if err := fixedBuffer.writeVInt(nodeIn.NumArcs()); err != nil {
		return err
	}
	if err := fixedBuffer.writeVInt(maxBytesPerArc); err != nil {
		return err
	}

	headerLen := builder.fixedLengthArcsBuffer.getPosition()

	// Expand the arcs in place, backwards.
	srcPos := builder.bytes.GetPosition()
	destPos := startAddress + headerLen + nodeIn.NumArcs()*maxBytesPerArc

	//if err := assert(destPos >= srcPos); err != nil {
	//	return err
	//}

	if destPos > srcPos {
		if err := builder.bytes.SkipBytes(destPos - srcPos); err != nil {
			return err
		}

		for arcIdx := nodeIn.NumArcs() - 1; arcIdx >= 0; arcIdx-- {
			destPos -= maxBytesPerArc
			arcLen := builder.numBytesPerArc[arcIdx]
			srcPos -= int64(arcLen)
			if srcPos != destPos {
				//if err := assert(destPos > srcPos); err != nil {
				//	return err
				//}
				// "destPos=" + destPos + " srcPos=" + srcPos + " arcIdx=" + arcIdx + " maxBytesPerArc=" + maxBytesPerArc + " arcLen=" + arcLen + " nodeIn.numArcs=" + nodeIn.numArcs;
				if err := builder.bytes.MoveBytes(srcPos, destPos, int64(arcLen)); err != nil {
					return err
				}
			}
		}
	}

	// Write the header.
	bytes := builder.fixedLengthArcsBuffer.GetBytes()
	return builder.bytes.WriteBytesAt(startAddress, bytes[0:headerLen])
}

func (f *FST[T]) writeNodeForDirectAddressing(builder *Builder[T], nodeIn *UnCompiledNode[T],
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
	if err := fixedBuffer.ensureCapacity(int(bufferOffset)); err != nil {
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
		if err := builder.bytes.CopyBytesToArray(srcPos, buffer[bufferOffset:bufferOffset+1]); err != nil {
			return err
		}
		// Skip the label, copy the remaining.
		remainingArcLen := srcArcLen - 1 - labelLen
		if remainingArcLen != 0 {
			if err := builder.bytes.CopyBytesToArray(srcPos+1+labelLen,
				buffer[bufferOffset+1:bufferOffset+1+remainingArcLen]); err != nil {
				return err
			}
		}
		if arcIdx == 0 {
			// Copy the label of the first arc only.
			bufferOffset -= labelLen
			if err := builder.bytes.CopyBytesToArray(srcPos+1,
				buffer[bufferOffset:bufferOffset+labelLen]); err != nil {
				return err
			}
		}
	}

	//if err := assert(bufferOffset == headerMaxLen+numPresenceBytes); err != nil {
	//	return err
	//}

	// Build the header in the buffer.
	// It is a false/special arc which is in fact a node header with node flags followed by node metadata.
	//fixedBuffer := builder.fixedLengthArcsBuffer
	if err := fixedBuffer.resetPosition(); err != nil {
		return err
	}
	if err := fixedBuffer.writeByte(ARCS_FOR_DIRECT_ADDRESSING); err != nil {
		return err
	}

	// labelRange instead of numArcs.
	if err := fixedBuffer.writeVInt(labelRange); err != nil {
		return err
	}

	// maxBytesPerArcWithoutLabel instead of maxBytesPerArc.
	if err := fixedBuffer.writeVInt(maxBytesPerArcWithoutLabel); err != nil {
		return err
	}

	headerLen := builder.fixedLengthArcsBuffer.getPosition()
	// Prepare the builder byte store. Enlarge or truncate if needed.
	nodeEnd := startAddress + headerLen + numPresenceBytes + totalArcBytes
	currentPosition := builder.bytes.GetPosition()
	if nodeEnd >= currentPosition {
		if err := builder.bytes.SkipBytes(nodeEnd - currentPosition); err != nil {
			return err
		}
	} else {
		if err := builder.bytes.Truncate(nodeEnd); err != nil {
			return err
		}
	}

	//if err := assert(builder.bytes.GetPosition() == nodeEnd); err != nil {
	//	return err
	//}

	// Write the header.
	writeOffset := startAddress
	buff := builder.fixedLengthArcsBuffer.GetBytes()
	if err := builder.bytes.WriteBytesAt(writeOffset, buff[0:headerLen]); err != nil {
		return err
	}
	writeOffset += headerLen

	// Write the presence bits
	if err := f.writePresenceBits(builder, nodeIn, writeOffset, numPresenceBytes); err != nil {
		return err
	}
	writeOffset += numPresenceBytes

	// Write the first label and the arcs.
	return builder.bytes.WriteBytesAt(writeOffset, buff[bufferOffset:bufferOffset+totalArcBytes])
}

func (f *FST[T]) writePresenceBits(builder *Builder[T], nodeIn *UnCompiledNode[T], dest, numPresenceBytes int64) error {
	bytePos := dest
	presenceBits := 1 // The first arc is always present.
	presenceIndex := 0
	previousLabel := nodeIn.Arcs[0].Label
	for arcIdx := 1; arcIdx < int(nodeIn.NumArcs()); arcIdx++ {
		label := nodeIn.Arcs[arcIdx].Label
		//if err := assert(label > previousLabel); err != nil {
		//	return err
		//}
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
	//err := assert(labelRange >= 0)
	//if err != nil {
	//	return 0, err
	//}
	return (labelRange + 7) >> 3, nil
}

// Reads the presence bits of a direct-addressing node. Actually we don't read them here,
// we just keep the pointer to the bit-table start and we skip them.
func (f *FST[T]) readPresenceBytes(arc *Arc[T], in BytesReader) error {

	// TODO: assert arc.bytesPerArc() > 0;
	// TODO: assert arc.nodeFlags() == ARCS_FOR_DIRECT_ADDRESSING;
	arc.bitTableStart = in.GetPosition()

	bytes, err := getNumPresenceBytes(arc.NumArcs())
	if err != nil {
		return err
	}
	return in.SkipBytes(int(bytes))
}

// Follows the follow arc and reads the last arc of its target; this changes the provided
// arc (2nd arg) in-place and returns it.
// Returns: Returns the second argument (arc).
func (f *FST[T]) readLastTargetArc(follow, arc *Arc[T], in BytesReader) (*Arc[T], error) {
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

func (f *FST[T]) readUnpackedNodeTarget(in BytesReader) (int64, error) {
	num, err := in.ReadUvarint()
	if err != nil {
		return 0, err
	}
	return int64(num), nil
}

// Reads a present direct addressing node arc, with the provided index in the label range and
// its corresponding presence index (which is the count of presence bits before it).
func (f *FST[T]) readArcByDirectAddressing(arc *Arc[T], in BytesReader, rangeIndex, presenceIndex int) (*Arc[T], error) {
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

func (f *FST[T]) readArc(arc *Arc[T], in BytesReader) (*Arc[T], error) {
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
					bits, err := CountBits(arc, in)
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

func flag(flags, bit int) bool {
	return flags&bit != 0
}
