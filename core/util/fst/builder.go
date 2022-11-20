package fst

import (
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"math"
)

// Builder Builds a minimal FST (maps an IntsRef term to an arbitrary output) from pre-sorted terms with outputs.
// The FST becomes an FSA if you use NoOutputs. The FST is written on-the-fly into a compact serialized format
// byte array, which can be saved to / loaded from a Directory or used directly for traversal.
// The FST is always finite (no cycles).
//
// NOTE: The algorithm is described at http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698
//
// The parameterized type T is the output type. See the subclasses of Outputs.
//
// FSTs larger than 2.1GB are now possible (as of Lucene 4.2). FSTs containing more than 2.1B nodes are also
// now possible, however they cannot be packed.
//
// lucene.experimental
type Builder struct {
	dedupHash *NodeHash
	fst       *FST
	//NO_OUTPUT any

	// private static final boolean DEBUG = true;

	// simplistic pruning: we prune node (and all following
	// nodes) if less than this number of terms go through it:
	minSuffixCount1 int64

	// better pruning: we prune node (and all following
	// nodes) if the prior node has less than this number of
	// terms go through it:
	minSuffixCount2 int64

	doShareNonSingletonNodes bool
	shareMaxTailLength       int

	lastInput []rune

	// NOTE: cutting this over to ArrayList instead loses ~6%
	// in build performance on 9.8M Wikipedia terms; so we
	// left this as an array:
	// current "frontier"
	frontier []*UnCompiledNode

	// Used for the BIT_TARGET_NEXT optimization (whereby
	// instead of storing the address of the target node for
	// a given arc, we mark a single bit noting that the next
	// node in the byte[] is the target node):
	lastFrozenNode int64

	// Reused temporarily while building the FST:
	numBytesPerArc      []int
	numLabelBytesPerArc []int64

	fixedLengthArcsBuffer *FixedLengthArcsBuffer

	arcCount                            int64
	nodeCount                           int64
	binarySearchNodeCount               int64
	directAddressingNodeCount           int64
	allowFixedLengthArcs                bool
	directAddressingMaxOversizingFactor float64
	directAddressingExpansionCredit     int64
	bytes                               *ByteStore
}

// NewBuilder Instantiates an FST/FSA builder without any pruning.
// A shortcut to Builder(FST.INPUT_TYPE, int, int, boolean, boolean, int, Outputs, boolean, int)
// with pruning options turned off.
func NewBuilder(inputType INPUT_TYPE, outputs Outputs) *Builder {
	return NewBuilderV1(inputType, 0, 0, true, true,
		math.MaxInt32, outputs, true, 15)
}

// NewBuilderV1 Instantiates an FST/FSA builder with all the possible tuning and construction tweaks. Read parameter documentation carefully.
//
// inputType – The input type (transition labels). Can be anything from FST.INPUT_TYPE enumeration. Shorter types will consume less memory. Strings (character sequences) are represented as FST.INPUT_TYPE.BYTE4 (full unicode codepoints).
// minSuffixCount1 – If pruning the input graph during construction, this threshold is used for telling if a node is kept or pruned. If transition_count(node) >= minSuffixCount1, the node is kept.
// minSuffixCount2 – (Note: only Mike McCandless knows what this one is really doing...)
// doShareSuffix – If true, the shared suffixes will be compacted into unique paths. This requires an additional RAM-intensive hash map for lookups in memory. Setting this parameter to false creates a single suffix path for all input sequences. This will result in a larger FST, but requires substantially less memory and CPU during building.
// doShareNonSingletonNodes – Only used if doShareSuffix is true. Set this to true to ensure FST is fully minimal, at cost of more CPU and more RAM during building.
// shareMaxTailLength – Only used if doShareSuffix is true. Set this to Integer.MAX_VALUE to ensure FST is fully minimal, at cost of more CPU and more RAM during building.
// outputs – The output type for each input sequence. Applies only if building an FST. For FSA, use NoOutputs.getSingleton() and NoOutputs.getNoOutput() as the singleton output object.
// allowFixedLengthArcs – Pass false to disable the fixed length arc optimization (binary search or direct addressing) while building the FST; this will make the resulting FST smaller but slower to traverse.
// bytesPageBits – How many bits wide to make each byte[] block in the BytesStore; if you know the FST will be large then make this larger. For example 15 bits = 32768 byte pages.
func NewBuilderV1(inputType INPUT_TYPE, minSuffixCount1, minSuffixCount2 int,
	doShareSuffix, doShareNonSingletonNodes bool, shareMaxTailLength int, outputs Outputs,
	allowFixedLengthArcs bool, bytesPageBits int) *Builder {

	builder := &Builder{
		minSuffixCount1:          int64(minSuffixCount1),
		minSuffixCount2:          int64(minSuffixCount2),
		doShareNonSingletonNodes: doShareNonSingletonNodes,
		shareMaxTailLength:       shareMaxTailLength,
		allowFixedLengthArcs:     allowFixedLengthArcs,
		fst:                      NewFST(inputType, outputs, bytesPageBits),
		frontier:                 make([]*UnCompiledNode, 0, 10),
		fixedLengthArcsBuffer:    NewFixedLengthArcsBuffer(),
	}

	builder.bytes = builder.fst.bytes
	// TODO: assert bytes != null;
	// pad: ensure no node gets address 0 which is reserved to mean
	// the stop state w/ no arcs
	_ = builder.bytes.WriteByte(0)

	if doShareSuffix {
		reader, err := builder.bytes.getReverseReader(false)
		if err != nil {
			return nil
		}
		builder.dedupHash = NewNodeHash(builder.fst, reader)
	}

	//builder.NO_OUTPUT = outputs.GetNoOutput()

	for i := 0; i < 10; i++ {
		node := NewUnCompiledNode(builder, i)
		builder.frontier = append(builder.frontier, node)
	}
	return builder
}

// SetDirectAddressingMaxOversizingFactor Overrides the default the maximum oversizing of fixed array allowed to enable direct addressing of arcs instead of binary search.
// Setting this factor to a negative value (e.g. -1) effectively disables direct addressing, only binary search nodes will be created.
// 请参阅:
// DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR
func (b *Builder) SetDirectAddressingMaxOversizingFactor(factor float64) *Builder {
	b.directAddressingMaxOversizingFactor = factor
	return b
}

func (b *Builder) GetDirectAddressingMaxOversizingFactor() float64 {
	return b.directAddressingMaxOversizingFactor
}

func (b *Builder) GetTermCount() int64 {
	return b.frontier[0].InputCount
}

func (b *Builder) GetNodeCount() int64 {
	// 1+ in order to count the -1 implicit final node
	return b.nodeCount + 1
}

func (b *Builder) GetArcCount() int64 {
	return b.arcCount
}

func (b *Builder) compileNode(nodeIn *UnCompiledNode, tailLength int) (*CompiledNode, error) {
	var node int64
	var err error
	bytesPosStart := b.bytes.GetPosition()
	if b.dedupHash != nil && (b.doShareNonSingletonNodes || nodeIn.NumArcs() <= 1) && tailLength <= b.shareMaxTailLength {
		if nodeIn.NumArcs() == 0 {
			node, err = b.fst.AddNode(b, nodeIn)
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
		node, err = b.fst.AddNode(b, nodeIn)
	}
	// TODO: assert node != -2;

	bytesPosEnd := b.bytes.GetPosition()
	if bytesPosEnd != bytesPosStart {
		// The FST added a new node:
		// TODO: assert bytesPosEnd > bytesPosStart;
		b.lastFrozenNode = node
	}

	nodeIn.Clear()

	fn := NewCompiledNode()
	fn.node = node
	return fn, nil
}

func (b *Builder) freezeTail(prefixLenPlus1 int) error {
	downTo := util.Max(1, prefixLenPlus1)

	for idx := len(b.lastInput); idx >= downTo; idx-- {
		doPrune := false   // 需要修剪
		doCompile := false // 需要编译

		node := b.frontier[idx]
		parent := b.frontier[idx-1]

		if node.InputCount < b.minSuffixCount1 {
			doPrune = true
			doCompile = true
		} else if idx > prefixLenPlus1 {
			// prune if parent's inputCount is less than suffixMinCount2
			if parent.InputCount < b.minSuffixCount2 || (b.minSuffixCount2 == 1 && parent.InputCount == 1 && idx > 1) {
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

		if node.InputCount < b.minSuffixCount2 || (b.minSuffixCount2 == 1 && node.InputCount == 1 && idx > 1) {
			// drop all arcs
			for arcIdx := 0; arcIdx < int(node.NumArcs()); arcIdx++ {
				target := node.Arcs[arcIdx].Target.(*UnCompiledNode)
				target.Clear()
			}
			node.Arcs = node.Arcs[:0]
		}

		if doPrune {
			// this node doesn't make it -- deref it
			node.Clear()
			err := parent.DeleteLast(int(b.lastInput[idx-1]), node)
			if err != nil {
				return err
			}
		} else {

			if b.minSuffixCount2 != 0 {
				err := b.compileAllTargets(node, len(b.lastInput)-idx)
				if err != nil {
					return err
				}
			}
			nextFinalOutput := node.Output

			// We "fake" the node as being final if it has no
			// outgoing arcs; in theory we could leave it
			// as non-final (the FST can represent this), but
			// FSTEnum, Util, etc., have trouble w/ non-final
			// dead-end states:
			isFinal := node.IsFinal || node.NumArcs() == 0

			if doCompile {
				// this node makes it and we now compile it.  first,
				// compile any targets that were previously
				// undecided:
				compileNode, err := b.compileNode(node, 1+len(b.lastInput)-idx)
				if err != nil {
					return err
				}

				parent.ReplaceLast(int(b.lastInput[idx-1]), compileNode, nextFinalOutput, isFinal)
			} else {
				// replaceLast just to install
				// nextFinalOutput/isFinal onto the arc
				parent.ReplaceLast(int(b.lastInput[idx-1]), node, nextFinalOutput, isFinal)
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

// Add the next input/output pair. The provided input must be sorted after the previous one according
// to IntsRef.compareTo. It's also OK to add the same input twice in a row with different outputs,
// as long as Outputs implements the Outputs.merge method. Note that input is fully consumed after
// this method is returned (so caller is free to reuse), but output is not. So if your outputs are
// changeable (eg ByteSequenceOutputs or IntSequenceOutputs) then you cannot reuse across calls.
//
// 添加input/output。提供的input必须要先进行排序。如果输入相同的input+不同的output，output需要实现merge方法。
func (b *Builder) Add(input []rune, output any) error {
	// TODO: if (output.equals(NO_OUTPUT)) {
	//      output = NO_OUTPUT;
	//    }

	// TODO: assert lastInput.length() == 0 || input.compareTo(lastInput.get()) >= 0: "inputs are added out of order lastInput=" + lastInput.get() + " vs input=" + input;
	// TODO: assert validOutput(output);

	if b.fst.outputs.IsNoOutput(output) {
		output = nil
	}

	if len(input) == 0 {
		// empty input: only allowed as first input.  we have
		// to special case this because the packed FST
		// format cannot represent the empty input since
		// 'finalness' is stored on the incoming arc, not on
		// the node
		b.frontier[0].InputCount++
		b.frontier[0].IsFinal = true
		b.fst.SetEmptyOutput(output)
		return nil
	}

	// compare shared prefix length
	pos1 := 0
	pos2 := 0
	pos1Stop := util.Min(len(b.lastInput), len(input))
	for {
		b.frontier[pos1].InputCount++
		if pos1 >= pos1Stop || b.lastInput[pos1] != input[pos2] {
			break
		}
		pos1++
		pos2++
	}

	prefixLenPlus1 := pos1 + 1

	// minimize/compile states from previous input's orphan'd suffix
	err := b.freezeTail(prefixLenPlus1)
	if err != nil {
		return err
	}

	// init tail states for current input
	for idx := prefixLenPlus1; idx <= len(input); idx++ {
		b.frontier[idx-1].AddArc(int(input[idx-1]), b.frontier[idx])
		b.frontier[idx].InputCount++
	}

	lastNode := b.frontier[len(input)]
	if len(b.lastInput) != len(input) || prefixLenPlus1 != len(input)+1 {
		lastNode.IsFinal = true
		//lastNode.Output = b.NO_OUTPUT
	}

	// push conflicting outputs forward, only as far as
	// needed
	for idx := 1; idx < prefixLenPlus1; idx++ {
		node := b.frontier[idx]
		parentNode := b.frontier[idx-1]

		lastOutput := parentNode.GetLastOutput(int(input[idx-1]))
		// TODO: assert validOutput(lastOutput);

		var commonOutputPrefix any
		var wordSuffix any

		if lastOutput != nil {
			commonOutputPrefix, err = b.fst.outputs.Common(output, lastOutput)
			if err != nil {
				return err
			}
			// assert validOutput(commonOutputPrefix);
			wordSuffix, err = b.fst.outputs.Subtract(lastOutput, commonOutputPrefix)
			if err != nil {
				return err
			}
			// TODO: assert validOutput(wordSuffix);
			err := parentNode.SetLastOutput(int(input[idx-1]), commonOutputPrefix)
			if err != nil {
				return err
			}
			err = node.PrependOutput(wordSuffix)
			if err != nil {
				return err
			}
		}

		output, err = b.fst.outputs.Subtract(output, commonOutputPrefix)
		if err != nil {
			return err
		}
		// TODO: assert validOutput(output);
	}

	if len(b.lastInput) == len(input) && prefixLenPlus1 == 1+len(input) {
		// same input more than 1 time in a row, mapping to
		// multiple outputs
		lastNode.Output, err = b.fst.outputs.Merge(lastNode.Output, output)
		if err != nil {
			return err
		}
	} else {
		// this new arc is private to this new input; set its
		// arc output to the leftover output:
		err := b.frontier[prefixLenPlus1-1].SetLastOutput(int(input[prefixLenPlus1-1]), output)
		if err != nil {
			return err
		}
	}

	// save last input
	b.lastInput = input

	//System.out.println("  count[0]=" + frontier[0].inputCount);
	return err
}

// Finish Returns final FST. NOTE: this will return null if nothing is accepted by the FST.
func (b *Builder) Finish() (*FST, error) {

	root := b.frontier[0]

	// minimize nodes in the last word's suffix
	err := b.freezeTail(0)
	if err != nil {
		return nil, err
	}
	if root.InputCount < b.minSuffixCount1 || root.InputCount < b.minSuffixCount2 || root.NumArcs() == 0 {
		if b.fst.emptyOutput == nil {
			return nil, nil
		} else if b.minSuffixCount1 > 0 || b.minSuffixCount2 > 0 {
			// empty string got pruned
			return nil, nil
		}
	} else {
		if b.minSuffixCount2 != 0 {
			err := b.compileAllTargets(root, len(b.lastInput))
			if err != nil {
				return nil, err
			}
		}
	}
	//if (DEBUG) System.out.println("  builder.finish root.isFinal=" + root.isFinal + " root.output=" + root.output);
	compileNode, err := b.compileNode(root, len(b.lastInput))
	if err != nil {
		return nil, err
	}

	err = b.fst.Finish(compileNode.node)
	if err != nil {
		return nil, err
	}

	return b.fst, nil
}

func (b *Builder) compileAllTargets(node *UnCompiledNode, tailLength int) error {
	for arcIdx := 0; arcIdx < int(node.NumArcs()); arcIdx++ {
		arc := node.Arcs[arcIdx]
		if !arc.Target.IsCompiled() {
			// not yet compiled
			n := arc.Target.(*UnCompiledNode)
			if n.NumArcs() == 0 {
				//System.out.println("seg=" + segment + "        FORCE final arc=" + (char) arc.label);
				arc.IsFinal, n.IsFinal = true, true
			}
			var err error
			arc.Target, err = b.compileNode(n, tailLength-1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR Default oversizing factor used to decide whether to encode a node with direct addressing or binary search. Default is 1: ensure no oversizing on average.
// This factor does not determine whether to encode a node with a list of variable length arcs or with fixed length arcs. It only determines the effective encoding of a node that is already known to be encoded with fixed length arcs. See FST.shouldExpandNodeWithFixedLengthArcs() and FST.shouldExpandNodeWithDirectAddressing().
// For English words we measured 217K nodes, only 3.27% nodes are encoded with fixed length arcs, and 99.99% of them with direct addressing. Overall FST memory reduced by 1.67%.
// For worst case we measured 168K nodes, 50% of them are encoded with fixed length arcs, and 14% of them with direct encoding. Overall FST memory reduced by 0.8%.
// Use TestFstDirectAddressing.main() and TestFstDirectAddressing.testWorstCaseForDirectAddressing() to evaluate a change.
// see: setDirectAddressingMaxOversizingFactor
const DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR = 1.0

// FixedLengthArcsBuffer Reusable buffer for building nodes with fixed length arcs (binary search or direct addressing).
type FixedLengthArcsBuffer struct {
	bytes []byte

	bado *store.ByteArrayDataOutput
}

func NewFixedLengthArcsBuffer() *FixedLengthArcsBuffer {
	bytes := make([]byte, 11)
	return &FixedLengthArcsBuffer{
		bytes: bytes,
		bado:  store.NewByteArrayDataOutput(bytes),
	}
}

// Ensures the capacity of the internal byte array. Enlarges it if needed.
func (f *FixedLengthArcsBuffer) ensureCapacity(capacity int) error {
	if len(f.bytes) < capacity {
		f.bytes = make([]byte, capacity)
		return f.bado.Reset(f.bytes)
	}
	return nil
}

func (f *FixedLengthArcsBuffer) resetPosition() error {
	return f.bado.Reset(f.bytes)
}

func (f *FixedLengthArcsBuffer) writeByte(b byte) error {
	return f.bado.WriteByte(b)
}

func (f *FixedLengthArcsBuffer) writeVInt(i int64) error {
	return f.bado.WriteUvarint(uint64(i))
}

func (f *FixedLengthArcsBuffer) getPosition() int64 {
	return int64(f.bado.GetPosition())
}

func (f *FixedLengthArcsBuffer) GetBytes() []byte {
	return f.bytes
}
