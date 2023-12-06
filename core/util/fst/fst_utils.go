package fst

import (
	"context"

	"github.com/samber/lo"
)

// Returns whether the given node should be expanded with fixed length arcs. Nodes will be
// expanded depending on their depth (distance from the root node) and their number of arcs.
// Nodes with fixed length arcs use more space, because they encode all arcs with a fixed
// number of bytes, but they allow either binary search or direct addressing on the arcs
// (instead of linear scan) on lookup by arc label.
//
// 返回是否应使用固定长度的弧扩展给定节点。节点将根据其深度（距根节点的距离）和弧数进行扩展。
// 具有固定长度弧的节点使用更多空间，因为它们使用固定数量的字节对所有弧进行编码，
// 但它们允许在通过弧标签查找时对弧进行二分搜索或直接寻址（而不是线性扫描）。
func shouldExpandNodeWithFixedLengthArcs(builder *Builder, node *UnCompiledNode) bool {
	// 如果不支持FixedLength返回false
	if !builder.allowFixedLengthArcs {
		return false
	}
	// 1. arc的数量 >= 10
	// 2. arc的深度 <= 3 且 arc的数量 >= 5
	return node.NumArcs() >= FIXED_LENGTH_ARC_DEEP_NUM_ARCS ||
		(node.Depth <= FIXED_LENGTH_ARC_SHALLOW_DEPTH && node.NumArcs() >= FIXED_LENGTH_ARC_SHALLOW_NUM_ARCS)

}

// Returns whether the given node should be expanded with direct addressing instead of binary search.
// Prefer direct addressing for performance if it does not oversize binary search byte size too much,
// so that the arcs can be directly addressed by label.
// See Also: Builder.getDirectAddressingMaxOversizingFactor()
//
// 返回是否应使用直接寻址而不是二进制搜索来扩展给定节点。如果不使二分搜索字节大小过大，
// 则优先选择直接寻址以提高性能，以便可以通过标签直接寻址弧。
// 另请参阅：Builder.getDirectAddressingMaxOversizingFactor()
func shouldExpandNodeWithDirectAddressing(builder *Builder, nodeIn *UnCompiledNode,
	numBytesPerArc, maxBytesPerArcWithoutLabel int64, labelRange int) bool {

	// Anticipate precisely the size of the encodings.
	sizeForBinarySearch := numBytesPerArc * int64(nodeIn.NumArcs())

	sizeForDirectAddressing := int64(getNumPresenceBytes(labelRange)) +
		int64(builder.numLabelBytesPerArc[0]) +
		maxBytesPerArcWithoutLabel*int64(nodeIn.NumArcs())

	// Determine the allowed oversize compared to binary search.
	// This is defined by a parameter of Fst Builder (default 1: no oversize).
	allowedOversize := int64(float64(sizeForBinarySearch) * builder.GetDirectAddressingMaxOversizingFactor())
	expansionCost := sizeForDirectAddressing - allowedOversize

	// SelectK direct addressing if either:
	// * 直接寻址大小比二分查找小。在这种情况下，请按减少的大小增加信用（以便稍后使用）。
	// * 直接寻址大小比二分查找大，但正信用允许超大大小。在这种情况下，按超额减少信用额。
	// In addition, do not try to oversize to a clearly too large node size
	// (this is the DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR parameter).
	if expansionCost <= 0 {
		builder.directAddressingExpansionCredit -= expansionCost
		return true
	}

	sizeLimit := int64(float64(allowedOversize) * DIRECT_ADDRESSING_MAX_OVERSIZE_WITH_CREDIT_FACTOR)
	if builder.directAddressingExpansionCredit >= expansionCost && sizeForDirectAddressing <= sizeLimit {
		builder.directAddressingExpansionCredit -= expansionCost
		return true
	}
	return false
}

func writeNodeForBinarySearch(ctx context.Context, builder *Builder,
	nodeIn *UnCompiledNode, startAddress int64, maxBytesPerArc int64) error {

	// Build the header in a buffer.
	// It is a false/special arc which is in fact a node header with node flags followed by node metadata.
	buffer := builder.fixedLengthArcsBuffer
	buffer.Reset()

	if err := buffer.WriteByte(ArcsForBinarySearch); err != nil {
		return err
	}
	if err := buffer.WriteUvarint(uint64(nodeIn.NumArcs())); err != nil {
		return err
	}
	if err := buffer.WriteUvarint(uint64(maxBytesPerArc)); err != nil {
		return err
	}

	headerLen := buffer.Len()

	// Expand the arcs in place, backwards.
	srcPos := builder.bytes.GetPosition()
	destPos := startAddress + int64(headerLen) + int64(nodeIn.NumArcs())*maxBytesPerArc

	if destPos > srcPos {
		if err := builder.bytes.SkipBytes(destPos - srcPos); err != nil {
			return err
		}

		for arcIdx := nodeIn.NumArcs() - 1; arcIdx >= 0; arcIdx-- {
			destPos -= maxBytesPerArc
			arcLen := int64(builder.numBytesPerArc[arcIdx])
			srcPos -= arcLen
			if srcPos != destPos {
				if err := builder.bytes.MoveBytes(ctx, srcPos, destPos, arcLen); err != nil {
					return err
				}
			}
		}
	}

	// Write the header.
	return builder.bytes.WriteBytesAt(ctx, startAddress, buffer.Bytes())
}

func writeNodeForDirectAddressing(ctx context.Context, builder *Builder, node *UnCompiledNode,
	startAddress int64, maxBytesPerArcWithoutLabel int, labelRange int) error {

	// Expand the arcs backwards in a buffer because we remove the labels.
	// So the obtained arcs might occupy less space. This is the reason why this
	// whole method is more complex.
	// Drop the label bytes since we can infer the label based on the arc index,
	// the presence bits, and the first label. Keep the first label.
	//
	// 因为我们删除了标签，所以在缓冲区中向后展开弧。因此获得的弧可能占用更少的空间。
	// 这就是整个方法更加复杂的原因。删除标签字节，因为我们可以根据弧索引、存在位和第一个标签推断标签。
	// 保留第一个标签。

	fixedBuffer := builder.fixedLengthArcsBuffer

	// Build the header in the buffer.
	// It is a false/special arc which is in fact a node header with node flags followed by node metadata.
	//fixedBuffer := builder.fixedLengthArcsBuffer
	fixedBuffer.Reset()
	if err := fixedBuffer.WriteByte(ArcsForDirectAddressing); err != nil {
		return err
	}
	// labelRange instead of numArcs.
	if err := fixedBuffer.WriteUvarint(uint64(labelRange)); err != nil {
		return err
	}
	// maxBytesPerArcWithoutLabel instead of maxBytesPerArc.
	if err := fixedBuffer.WriteUvarint(uint64(maxBytesPerArcWithoutLabel)); err != nil {
		return err
	}

	// 写入PresenceBytes
	numPresenceBytes := int64(getNumPresenceBytes(labelRange))
	presenceBits := getPresenceBits(node, numPresenceBytes)
	if _, err := fixedBuffer.Write(presenceBits); err != nil {
		return err
	}

	srcPos := builder.bytes.GetPosition() - int64(lo.Sum(builder.numLabelBytesPerArc))

	for arcIdx := 0; arcIdx < node.NumArcs(); arcIdx++ {
		labelLen := int64(builder.numLabelBytesPerArc[arcIdx])
		srcArcLen := int64(builder.numBytesPerArc[arcIdx])

		if arcIdx == 0 {
			if err := builder.bytes.CopyTo(ctx, srcPos, srcArcLen, fixedBuffer); err != nil {
				return err
			}

		} else {
			// Copy the flags.
			if err := builder.bytes.CopyTo(ctx, srcPos, 1, fixedBuffer); err != nil {
				return err
			}

			// Skip the label, copy the remaining.
			remainingArcLen := srcArcLen - 1 - labelLen
			if err := builder.bytes.CopyTo(ctx, srcPos+1+labelLen, remainingArcLen, fixedBuffer); err != nil {
				return err
			}
		}

		srcPos += srcArcLen
	}

	// Prepare the builder byte store. Enlarge or truncate if needed.
	nodeEnd := startAddress + int64(fixedBuffer.Len())
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

	return builder.bytes.WriteBytesAt(ctx, startAddress, fixedBuffer.Bytes())
}

func getPresenceBits(nodeIn *UnCompiledNode, numPresenceBytes int64) []byte {
	presenceBits := 1 // The first arc is always present.
	presenceIndex := 0

	res := make([]byte, 0, numPresenceBytes)

	previousLabel := nodeIn.Arcs[0].Label
	for arcIdx := 1; arcIdx < nodeIn.NumArcs(); arcIdx++ {
		label := nodeIn.Arcs[arcIdx].Label
		presenceIndex += label - previousLabel
		for presenceIndex >= BYTE_SIZE {
			res = append(res, byte(presenceBits))
			presenceBits = 0
			presenceIndex -= BYTE_SIZE
		}
		// Set the bit at presenceIndex to flag that the corresponding arc is present.
		presenceBits |= 1 << presenceIndex
		previousLabel = label
	}

	res = append(res, byte(presenceBits))

	return res
}

// Gets the number of bytes required to flag the presence of each arc in the given label range, one bit per arc.
// 获取标记给定标签范围中每个弧的存在所需的字节数，每个弧一位。
func getNumPresenceBytes(labelRange int) int {
	// 可以看作labelRange/8 + (labelRange%8>8)? 1: 0
	return (labelRange + 7) >> 3
}

// Reads the presence bits of a direct-addressing node. Actually we don't read them here,
// we just keep the pointer to the bit-table start and we skip them.
func readPresenceBytes(ctx context.Context, in BytesReader, arc *Arc) error {
	arc.bitTableStart = in.GetPosition()
	numBytes := getNumPresenceBytes(arc.NumArcs())
	return in.SkipBytes(ctx, numBytes)
}

// Follows the follow arc and reads the last arc of its target; this changes the provided
// arc (2nd arg) in-place and returns it.
// Returns: Returns the second argument (arc).
func (f *FST) readLastTargetArc(ctx context.Context, in BytesReader, follow, arc *Arc) (*Arc, error) {
	if !TargetHasArcs(follow) {
		arc.label = END_LABEL
		arc.target = FINAL_END_NODE
		arc.output = follow.NextFinalOutput()
		arc.flags = BitLastArc
		arc.nodeFlags = arc.flags
		return arc, nil
	}
	if err := in.SetPosition(follow.Target()); err != nil {
		return nil, err
	}

	flags, err := in.ReadByte()
	if err != nil {
		return nil, err
	}
	arc.nodeFlags = flags

	if flags == ArcsForBinarySearch || flags == ArcsForDirectAddressing {
		// Special arc which is actually a node header for fixed length arcs.
		// Jump straight to end to find the last arc.
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

		if flags == ArcsForDirectAddressing {
			if err := readPresenceBytes(ctx, in, arc); err != nil {
				return nil, err
			}

			arc.firstLabel, err = f.ReadLabel(ctx, in)
			if err != nil {
				return nil, err
			}

			arc.posArcsStart = in.GetPosition()

			if _, err := f.ReadLastArcByDirectAddressing(ctx, arc, in); err != nil {
				return nil, err
			}
		} else {
			arc.arcIdx = arc.NumArcs() - 2
			arc.posArcsStart = in.GetPosition()
			if _, err := f.ReadNextRealArc(ctx, in, arc); err != nil {
				return nil, err
			}
		}
		return arc, nil
	}

	arc.flags = flags
	// non-array: linear scan
	arc.bytesPerArc = 0

	for !arc.IsLast() {
		// skip this arc:
		if _, err := f.ReadLabel(ctx, in); err != nil {
			return nil, err
		}
		if arc.matchFlag(BitArcHasOutput) {
			if err := f.manager.SkipOutput(ctx, in); err != nil {
				return nil, err
			}
		}
		if arc.matchFlag(BitArcHasFinalOutput) {
			if err := f.manager.SkipFinalOutput(ctx, in); err != nil {
				return nil, err
			}
		}

		if arc.matchFlag(BitStopNode) {
		} else if arc.matchFlag(BitTargetNext) {
		} else {
			if _, err := f.readUnpackedNodeTarget(ctx, in); err != nil {
				return nil, err
			}
		}

		flagByte, err := in.ReadByte()
		if err != nil {
			return nil, err
		}
		arc.flags = flagByte
	}
	// Undo the byte flags we read:
	if err := in.SkipBytes(ctx, -1); err != nil {
		return nil, err
	}

	arc.nextArc = in.GetPosition()
	if _, err := f.ReadNextRealArc(ctx, in, arc); err != nil {
		return nil, err
	}

	return arc, nil
}

func (f *FST) readUnpackedNodeTarget(ctx context.Context, in BytesReader) (int64, error) {
	num, err := in.ReadUvarint(ctx)
	if err != nil {
		return 0, err
	}
	return int64(num), nil
}

// Reads a present direct addressing node arc, with the provided index in the label range and
// its corresponding presence index (which is the count of presence bits before it).
func (f *FST) readArcByDirectAddressing(ctx context.Context, arc *Arc, in BytesReader,
	rangeIndex, presenceIndex int) (*Arc, error) {

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

	return f.readArc(ctx, in, arc)
}

// Reads an arc.
// Precondition: The arc flags byte has already been read and set;
// the given BytesReader is positioned just after the arc flags byte.
func (f *FST) readArc(ctx context.Context, in BytesReader, arc *Arc) (*Arc, error) {
	if arc.NodeFlags() == ArcsForDirectAddressing {
		arc.label = arc.FirstLabel() + arc.ArcIdx()
	} else {
		label, err := f.ReadLabel(ctx, in)
		if err != nil {
			return nil, err
		}
		arc.label = label
	}

	if arc.matchFlag(BitArcHasOutput) {
		output := f.manager.New()
		if err := f.manager.Read(ctx, in, output); err != nil {
			return nil, err
		}
		arc.output = output
	} else {
		arc.output = f.manager.EmptyOutput()
	}

	if arc.matchFlag(BitArcHasFinalOutput) {
		output := f.manager.New()
		if err := f.manager.ReadFinalOutput(ctx, in, output); err != nil {
			return nil, err
		}
		arc.nextFinalOutput = output
	} else {
		arc.nextFinalOutput = f.manager.EmptyOutput()
	}

	if arc.matchFlag(BitStopNode) {
		if arc.matchFlag(BitFinalArc) {
			arc.target = FINAL_END_NODE
		} else {
			arc.target = NON_FINAL_END_NODE
		}
		arc.nextArc = in.GetPosition() // Only useful for list.
	} else if arc.matchFlag(BitTargetNext) {
		arc.nextArc = in.GetPosition() // Only useful for list.
		// TODO: would be nice to make this lazy -- maybe
		// caller doesn't need the target and is scanning arcs...
		if !arc.matchFlag(BitLastArc) {
			if arc.BytesPerArc() == 0 {
				// must scan
				if err := f.seekToNextNode(ctx, in); err != nil {
					return nil, err
				}
			} else {
				var numArcs int
				if arc.nodeFlags == ArcsForDirectAddressing {
					bits, err := CountBits(arc, in)
					if err != nil {
						return nil, err
					}
					numArcs = bits
				} else {
					numArcs = arc.NumArcs()
				}

				if err := in.SetPosition(arc.PosArcsStart() - int64(arc.BytesPerArc()*numArcs)); err != nil {
					return nil, err
				}
			}
		}
		arc.target = in.GetPosition()
	} else {
		target, err := f.readUnpackedNodeTarget(ctx, in)
		if err != nil {
			return nil, err
		}
		arc.target = target
		arc.nextArc = in.GetPosition()
	}
	return arc, nil
}

func readEndArc(follow, arc *Arc) *Arc {
	if !follow.IsFinal() {
		return nil
	}

	if follow.Target() <= 0 {
		arc.flags = BitLastArc
	} else {
		arc.flags = 0
		// NOTE: nextArc is a node (not an address!) in this case:
		arc.nextArc = follow.Target()
	}
	arc.output = follow.NextFinalOutput()
	arc.label = END_LABEL
	return arc
}

func (f *FST) seekToNextNode(ctx context.Context, in BytesReader) error {
	for {
		flags, err := in.ReadByte()
		if err != nil {
			return err
		}

		if _, err = f.ReadLabel(ctx, in); err != nil {
			return err
		}

		if flag(int(flags), BitArcHasOutput) {
			if err := f.manager.SkipOutput(ctx, in); err != nil {
				return err
			}
		}

		if flag(int(flags), BitArcHasFinalOutput) {
			if err := f.manager.SkipFinalOutput(ctx, in); err != nil {
				return err
			}
		}

		if !flag(int(flags), BitStopNode) && !flag(int(flags), BitTargetNext) {
			if _, err := f.readUnpackedNodeTarget(ctx, in); err != nil {
				return err
			}
		}

		if flag(int(flags), BitLastArc) {
			return nil
		}
	}
}

func flag(flags, bit int) bool {
	return flags&bit != 0
}
