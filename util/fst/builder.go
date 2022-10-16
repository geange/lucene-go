package fst

import (
	. "github.com/geange/lucene-go/math"
	. "github.com/geange/lucene-go/util/structure"
	"math"
	"reflect"
)

// TODO: could we somehow stream an FST to disk while we
// build it?

// Builder Builds a minimal FST (maps an IntsRef term to an arbitrary output) from pre-sorted terms with
// outputs. The FST becomes an FSA if you use NoOutputs. The FST is written on-the-fly into a compact
// serialized format byte array, which can be saved to / loaded from a Directory or used directly for traversal.
// The FST is always finite (no cycles).
//
// NOTE: The algorithm is described at http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698
// The parameterized type T is the output type. See the subclasses of Outputs.
//
// FSTs larger than 2.1GB are now possible (as of Lucene 4.2). FSTs containing more than 2.1B nodes are
// also now possible, however they cannot be packed.
//
// lucene.experimental
type Builder[T any] struct {
	dedupHash *NodeHash[T]

	fst *FST[T]

	noOutput *Box[T]

	// private static final boolean DEBUG = true;

	// simplistic pruning: we prune node (and all following
	// nodes) if less than this number of terms go through it:
	minSuffixCount1 int

	// better pruning: we prune node (and all following
	// nodes) if the prior node has less than this number of
	// terms go through it:
	minSuffixCount2 int

	doShareNonSingletonNodes bool

	shareMaxTailLength int

	lastInput *IntsRefBuilder

	// NOTE: cutting this over to ArrayList instead loses ~6%
	// in build performance on 9.8M Wikipedia terms; so we
	// left this as an array:
	// current "frontier"
	frontier []*UnCompiledNode[T]

	// Used for the BIT_TARGET_NEXT optimization (whereby
	// instead of storing the address of the target node for
	// a given arc, we mark a single bit noting that the next
	// node in the byte[] is the target node):
	lastFrozenNode int

	// Reused temporarily while building the FST:
	numBytesPerArc []int

	numLabelBytesPerArc []int

	fixedLengthArcsBuffer               *FixedLengthArcsBuffer
	arcCount                            int
	nodeCount                           int
	binarySearchNodeCount               int
	directAddressingNodeCount           int
	allowFixedLengthArcs                bool
	directAddressingMaxOversizingFactor float64
	directAddressingExpansionCredit     int
	bytes                               *BytesStore
}

// NewBuilder Instantiates an FST/FSA builder without any pruning.
// A shortcut to Builder(FST.INPUT_TYPE, int, int, boolean, boolean, int, Outputs, boolean, int)
// with pruning options turned off.
func NewBuilder[T any](inputType INPUT_TYPE, outputs Outputs[T]) (*Builder[T], error) {
	return NewBuilderV1[T](inputType, 0, 0, true, true, math.MaxInt32, outputs, true, 15)
}

//public Builder(FST.INPUT_TYPE inputType, int minSuffixCount1, int minSuffixCount2, boolean doShareSuffix,
//boolean doShareNonSingletonNodes, int shareMaxTailLength, Outputs<T> outputs,
//boolean allowFixedLengthArcs, int bytesPageBits) {

// NewBuilderV1 nstantiates an FST/FSA builder with all the possible tuning and construction tweaks.
// Read parameter documentation carefully.
// Params:
//
//	inputType – The input type (transition labels). Can be anything from FST.INPUT_TYPE enumeration. Shorter types will consume less memory. Strings (character sequences) are represented as FST.INPUT_TYPE.BYTE4 (full unicode codepoints). minSuffixCount1 – If pruning the input graph during construction, this threshold is used for telling if a node is kept or pruned. If transition_count(node) >= minSuffixCount1, the node is kept.
//	minSuffixCount2 – (Note: only Mike McCandless knows what this one is really doing...)
//	doShareSuffix – If true, the shared suffixes will be compacted into unique paths. This requires an additional RAM-intensive hash map for lookups in memory. Setting this parameter to false creates a single suffix path for all input sequences. This will result in a larger FST, but requires substantially less memory and CPU during building.
//	doShareNonSingletonNodes – Only used if doShareSuffix is true. Set this to true to ensure FST is fully minimal, at cost of more CPU and more RAM during building. shareMaxTailLength – Only used if doShareSuffix is true. Set this to Integer.MAX_VALUE to ensure FST is fully minimal, at cost of more CPU and more RAM during building. outputs – The output type for each input sequence. Applies only if building an FST. For FSA, use NoOutputs.getSingleton() and NoOutputs.getNoOutput() as the singleton output object. allowFixedLengthArcs – Pass false to disable the fixed length arc optimization (binary search or direct addressing) while building the FST; this will make the resulting FST smaller but slower to traverse. bytesPageBits – How many bits wide to make each byte[] block in the BytesStore; if you know the FST will be large then make this larger. For example 15 bits = 32768 byte pages.
func NewBuilderV1[T any](inputType INPUT_TYPE, minSuffixCount1, minSuffixCount2 int,
	doShareSuffix, doShareNonSingletonNodes bool, shareMaxTailLength int, outputs Outputs[T],
	allowFixedLengthArcs bool, bytesPageBits int) (*Builder[T], error) {

	panic("")
}

// SetDirectAddressingMaxOversizingFactor Overrides the default the maximum oversizing of fixed array allowed
// to enable direct addressing of arcs instead of binary search.
// Setting this factor to a negative value (e.g. -1) effectively disables direct addressing, only binary
// search nodes will be created.
// See Also: DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR
func (b *Builder[T]) SetDirectAddressingMaxOversizingFactor(factor float64) *Builder[T] {
	b.directAddressingMaxOversizingFactor = factor
	return b
}

// GetDirectAddressingMaxOversizingFactor See Also: SetDirectAddressingMaxOversizingFactor(float)
func (b *Builder[T]) GetDirectAddressingMaxOversizingFactor() float64 {
	return b.directAddressingMaxOversizingFactor
}

func (b *Builder[T]) GetTermCount() int {
	return b.frontier[0].inputCount
}

func (b *Builder[T]) GetNodeCount() int {
	// 1+ in order to count the -1 implicit final node
	return 1 + b.nodeCount
}

func (b *Builder[T]) GetArcCount() int {
	return b.arcCount
}

func (b *Builder[T]) GetMappedStateCount() int {
	if b.dedupHash == nil {
		return 0
	}
	return b.nodeCount
}

func (b *Builder[T]) compileNode(nodeIn *UnCompiledNode[T], tailLength int) (*CompiledNode, error) {
	node := 0
	bytesPosStart := b.bytes.getPosition()
	var err error
	if b.dedupHash != nil && (b.doShareNonSingletonNodes || nodeIn.numArcs <= 1) && tailLength <= b.shareMaxTailLength {
		if nodeIn.numArcs == 0 {
			node, err = b.fst.addNode(b, nodeIn)
			if err != nil {
				return nil, err
			}
			b.lastFrozenNode = node
		} else {
			node, err = b.dedupHash.Add(b, nodeIn)
			if err != nil {
				return nil, err
			}
		}
	} else {
		node, err = b.fst.addNode(b, nodeIn)
		if err != nil {
			return nil, err
		}
	}
	// assert node != -2;

	bytesPosEnd := b.bytes.getPosition()
	if bytesPosEnd != bytesPosStart {
		// The FST added a new node:
		// assert bytesPosEnd > bytesPosStart;
		b.lastFrozenNode = node
	}

	nodeIn.Clear()

	fn := NewCompiledNode()
	fn.node = int64(node)
	return fn, nil
}

func (b *Builder[T]) freezeTail(prefixLenPlus1 int) error {
	//System.out.println("  compileTail " + prefixLenPlus1);
	downTo := Max(1, prefixLenPlus1)

	// 如果 lastInput 长度为0，啥都不干
	for idx := b.lastInput.Length(); idx >= downTo; idx-- {

		doPrune := false
		doCompile := false

		node := b.frontier[idx]
		parent := b.frontier[idx-1]

		if node.inputCount < b.minSuffixCount1 {
			doPrune = true
			doCompile = true
		} else if idx > prefixLenPlus1 {
			// prune if parent's inputCount is less than suffixMinCount2
			// 父级的inputCount 小于 minSuffixCount2，进行修建
			if parent.inputCount < b.minSuffixCount2 || (b.minSuffixCount2 == 1 && parent.inputCount == 1 && idx > 1) {
				// my parent, about to be compiled, doesn't make the cut, so
				// I'm definitely pruned

				// if minSuffixCount2 is 1, we keep only up
				// until the 'distinguished edge', ie we keep only the
				// 'divergent' part of the FST. if my parent, about to be
				// compiled, has inputCount 1 then we are already past the
				// distinguished edge.  NOTE: this only works if
				// the FST outputs are not "compressible" (simple
				// ords ARE compressible).
				doPrune = true
			} else {
				// my parent, about to be compiled, does make the cut, so
				// I'm definitely not pruned
				doPrune = false
			}
			doCompile = true
		} else {
			// if pruning is disabled (count is 0) we can always
			// compile current node
			doCompile = b.minSuffixCount2 == 0
		}

		//System.out.println("    label=" + ((char) lastInput.ints[lastInput.offset+idx-1]) + " idx=" + idx + " inputCount=" + frontier[idx].inputCount + " doCompile=" + doCompile + " doPrune=" + doPrune);

		if node.inputCount < b.minSuffixCount2 || (b.minSuffixCount2 == 1 && node.inputCount == 1 && idx > 1) {
			// drop all arcs
			for arcIdx := 0; arcIdx < node.numArcs; arcIdx++ {
				target := node.arcs[arcIdx].target.(*UnCompiledNode[T])
				target.Clear()
			}
			node.numArcs = 0
		}

		if doPrune {
			// this node doesn't make it -- deref it
			node.Clear()
			parent.DeleteLast(b.lastInput.IntAt(idx-1), node)
		} else {

			if b.minSuffixCount2 != 0 {
				b.compileAllTargets(node, b.lastInput.Length()-idx)
			}
			nextFinalOutput := node.output

			// We "fake" the node as being final if it has no
			// outgoing arcs; in theory we could leave it
			// as non-final (the FST can represent this), but
			// FSTEnum, Util, etc., have trouble w/ non-final
			// dead-end states:
			isFinal := node.isFinal || node.numArcs == 0

			if doCompile {
				// this node makes it and we now compile it.  first,
				// compile any targets that were previously
				// undecided:
				compileNode, err := b.compileNode(node, 1+b.lastInput.Length()-idx)
				if err != nil {
					return err
				}
				parent.ReplaceLast(b.lastInput.IntAt(idx-1), compileNode, nextFinalOutput, isFinal)
			} else {
				// replaceLast just to install
				// nextFinalOutput/isFinal onto the arc
				parent.ReplaceLast(b.lastInput.IntAt(idx-1), node, nextFinalOutput, isFinal)
				// this node will stay in play for now, since we are
				// undecided on whether to prune it.  later, it
				// will be either compiled or pruned, so we must
				// allocate a new node:

				b.frontier[idx] = NewUnCompiledNode(b, idx)
			}
		}
	}
	return nil
}

// Add the next input/output pair. The provided input must be sorted after the previous one
// according to IntsRef.compareTo. It's also OK to add the same input twice in a row with
// different outputs, as long as Outputs implements the Outputs.merge method. Note that input
// is fully consumed after this method is returned (so caller is free to reuse), but output is not.
// So if your outputs are changeable (eg ByteSequenceOutputs or IntSequenceOutputs) then you cannot
// reuse across calls.
//
//   - 计算当前字符串和上一个字符串的公共前缀
//   - 调用freezeTail方法, 从尾部一直到公共前缀的节点，将已经确定的状态节点冻结。
//     这里会将UncompiledNode序列化到bytes当中，并转换成CompiledNode。
//   - 将当前字符串形成状态节点加入到frontier中
//   - 调整每条边输出值。
func (b *Builder[T]) Add(input *Box[[]int], output *Box[T]) error {
	// De-dup NO_OUTPUT since it must be a singleton:
	if reflect.DeepEqual(output, b.noOutput) {
		output = b.noOutput
	}

	if len(input.V) == 0 {
		// empty input: only allowed as first input.  we have
		// to special case this because the packed FST
		// format cannot represent the empty input since
		// 'finalness' is stored on the incoming arc, not on
		// the node
		// 空输入：仅允许作为第一个输入。我们必须对这种情况进行特殊处理，因为压缩的FST格式不能表示空输入，
		// 因为“最终性”存储在输入弧（Arc）上，而不是节点上
		b.frontier[0].inputCount++
		b.frontier[0].isFinal = true
		b.fst.setEmptyOutput(output)
		return nil
	}

	// compare shared prefix length
	// 比较共享前缀的长度
	pos1, pos2 := 0, 0
	pos1Stop := Min(b.lastInput.Length(), len(input.V))
	for {
		b.frontier[pos1].inputCount++
		if pos1 >= pos1Stop || b.lastInput.IntAt(pos1) != input.V[pos2] {
			break
		}
		pos1++
		pos2++
	}
	prefixLenPlus1 := pos1 + 1

	if len(b.frontier) < len(input.V)+1 {
		next := Grow(b.frontier, len(input.V)+1)
		for i := len(b.frontier); i < len(next); i++ {
			next[i] = NewUnCompiledNode(b, i)
		}
		b.frontier = next
	}

	// minimize/compile states from previous input's
	// orphan'd suffix
	err := b.freezeTail(prefixLenPlus1)
	if err != nil {
		return err
	}

	// init tail states for current input
	// 在 node 中创建 arc
	for idx := prefixLenPlus1; idx <= len(input.V); idx++ {

		// input.ints[input.offset + idx - 1] => 返回
		b.frontier[idx-1].AddArc(input.V[idx-1], b.frontier[idx])
		// 输入数+1
		b.frontier[idx].inputCount++
	}

	lastNode := b.frontier[len(input.V)]
	if b.lastInput.Length() != len(input.V) || prefixLenPlus1 != len(input.V)+1 {
		lastNode.isFinal = true
		lastNode.output = b.noOutput
	}

	// push conflicting outputs forward, only as far as needed
	// 在需要时将冲突的输出（outputs）向前推进
	// 首次不触发
	for idx := 1; idx < prefixLenPlus1; idx++ {
		node := b.frontier[idx]
		parentNode := b.frontier[idx-1]

		lastOutput := parentNode.GetLastOutput(int(input.V[idx-1]))
		// assert validOutput(lastOutput);

		var commonOutputPrefix, wordSuffix *Box[T]

		if lastOutput != b.noOutput {
			// 获取
			commonOutputPrefix = b.fst.outputs.Common(output, lastOutput)
			// assert validOutput(commonOutputPrefix);
			wordSuffix = b.fst.outputs.Subtract(lastOutput, commonOutputPrefix)
			// assert validOutput(wordSuffix);
			parentNode.SetLastOutput(input.V[idx-1], commonOutputPrefix)
			node.PrependOutput(wordSuffix)
		} else {
			commonOutputPrefix, wordSuffix = b.noOutput, b.noOutput
		}

		output = b.fst.outputs.Subtract(output, commonOutputPrefix)
		// assert validOutput(output);
	}

	if b.lastInput.Length() == len(input.V) && prefixLenPlus1 == 1+len(input.V) {
		// same input more than 1 time in a row, mapping to
		// multiple outputs
		// 输入的字符串相同，需要merge操作
		lastNode.output = b.fst.outputs.Merge(lastNode.output, output)
	} else {
		// this new arc is private to this new input; set its
		// arc output to the leftover output:
		b.frontier[prefixLenPlus1-1].SetLastOutput(input.V[prefixLenPlus1-1], output)
	}

	// save last input
	b.lastInput.CopyInts(input.V)
	return nil
}

func (b *Builder[T]) validOutput(output any) bool {
	panic("")
}

// Finish Returns final FST. NOTE: this will return null if nothing is accepted by the FST.
func (b *Builder[T]) Finish() (*FST[T], error) {
	root := b.frontier[0]

	// minimize nodes in the last word's suffix
	err := b.freezeTail(0)
	if err != nil {
		return nil, err
	}

	if root.inputCount < b.minSuffixCount1 ||
		root.inputCount < b.minSuffixCount2 ||
		root.numArcs == 0 {

		if b.fst.emptyOutput == nil {
			return nil, nil
		} else if b.minSuffixCount1 > 0 || b.minSuffixCount2 > 0 {
			// empty string got pruned
			return nil, nil
		}
	} else {
		if b.minSuffixCount2 != 0 {
			err := b.compileAllTargets(root, b.lastInput.Length())
			if err != nil {
				return nil, err
			}
		}
	}

	//if (DEBUG) System.out.println("  builder.finish root.isFinal=" + root.isFinal + " root.output=" + root.output);
	node, err := b.compileNode(root, b.lastInput.Length())
	if err != nil {
		return nil, err
	}
	err = b.fst.finish(int(node.node))
	if err != nil {
		return nil, err
	}
	return b.fst, nil
}

func (b *Builder[T]) compileAllTargets(node *UnCompiledNode[T], tailLength int) error {
	for arcIdx := 0; arcIdx < node.numArcs; arcIdx++ {
		arc := node.arcs[arcIdx]
		if !arc.target.IsCompiled() {
			// not yet compiled
			n := arc.target.(*UnCompiledNode[T])
			if n.numArcs == 0 {
				//System.out.println("seg=" + segment + "        FORCE final arc=" + (char) arc.label);
				arc.isFinal, n.isFinal = true, true
			}
			var err error
			arc.target, err = b.compileNode(n, tailLength-1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
