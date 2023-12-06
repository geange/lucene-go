package fst

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"slices"
)

// Builder
// Builds a minimal FST (maps a term([]int) to an arbitrary output) from pre-sorted terms with output.
// The FST becomes an FSA if you use NoOutputs. The FST is written on-the-fly into a compact serialized format
// byte array, which can be saved to / loaded from a Directory or used directly for traversal.
// The FST is always finite (no cycles).
//
// NOTE: The algorithm is described at http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.24.3698
//
// The parameterized type T is the output type. See the subclasses of Output.
//
// FSTs larger than 2.1GB are now possible (as of Lucene 4.2). FSTs containing more than 2.1B nodes are also
// now possible, however they cannot be packed.
//
// lucene.experimental
type Builder struct {
	dedupHash *NodeHash
	fst       *FST
	noOutput  Output

	// simplistic pruning: we prune node (and all following
	// nodes) if less than this number of terms go through it:
	//
	// 简单化修剪：如果通过节点的项数少于此数量，我们将忽略节点（以及所有后续节点）
	minSuffixCount1 int

	// better pruning: we prune node (and all following nodes)
	// if the prior node has less than this number of terms
	// go through it:
	//
	// 更好的修剪：如果先前节点经过它的项数少于此数量，我们会修剪节点（以及所有后续节点
	minSuffixCount2 int

	doShareNonSingletonNodes bool
	shareMaxTailLength       int
	lastInput                []int

	// NOTE: cutting this over to ArrayList instead loses ~6%
	// in build performance on 9.8M Wikipedia terms; so we
	// left this as an array:
	// current "frontier"
	// 注意：将其切换到 ArrayList 会导致 980 万个维基百科术语的构建性能损失约 6%；所以我们将其保留为数组
	frontier []*UnCompiledNode

	// Used for the BitTargetNext optimization (whereby
	// instead of storing the address of the target node for
	// a given arc, we mark a single bit noting that the next
	// node in the byte[] is the target node):
	//
	// 用于 BitTargetNext 优化（我们不是存储给定弧的目标节点的地址，
	// 而是标记一个位，指出 byte[] 中的下一个节点是目标节点）
	lastFrozenNode int64

	// Reused temporarily while building the FST:
	numBytesPerArc      []int
	numLabelBytesPerArc []int

	fixedLengthArcsBuffer               *Buffer
	arcCount                            int
	nodeCount                           int
	binarySearchNodeCount               int
	directAddressingNodeCount           int
	allowFixedLengthArcs                bool
	directAddressingMaxOversizingFactor float64
	directAddressingExpansionCredit     int64
	bytes                               *ByteStore
}

// NewBuilder Instantiates an FST/FSA builder without any pruning.
// A shortcut to Builder(Fst.INPUT_TYPE, int, int, boolean, boolean, int, Output, boolean, int)
// with pruning options turned off.
func NewBuilder(inputType InputType, manager OutputManager, options ...BuilderOption) (*Builder, error) {
	opt := &builderOption{
		minSuffixCount1:          0,
		minSuffixCount2:          0,
		doShareSuffix:            true,
		doShareNonSingletonNodes: true,
		shareMaxTailLength:       math.MaxInt32,
		allowFixedLengthArcs:     true,
		bytesPageBits:            15,
	}

	for _, fn := range options {
		fn(opt)
	}

	return newBuilder(inputType, manager, opt.minSuffixCount1, opt.minSuffixCount2, opt.doShareSuffix,
		opt.doShareNonSingletonNodes, opt.shareMaxTailLength,
		opt.allowFixedLengthArcs, opt.bytesPageBits)
}

type BuilderOption func(option *builderOption)

type builderOption struct {
	minSuffixCount1          int
	minSuffixCount2          int
	doShareSuffix            bool // 是否共享后缀
	doShareNonSingletonNodes bool
	shareMaxTailLength       int
	allowFixedLengthArcs     bool
	bytesPageBits            int
}

func WithMinSuffixCount1(minSuffixCount1 int) BuilderOption {
	return func(option *builderOption) {
		option.minSuffixCount1 = minSuffixCount1
	}
}

func WithMinSuffixCount2(minSuffixCount2 int) BuilderOption {
	return func(option *builderOption) {
		option.minSuffixCount2 = minSuffixCount2
	}
}

func WithDoShareSuffix(doShareSuffix bool) BuilderOption {
	return func(option *builderOption) {
		option.doShareSuffix = doShareSuffix
	}
}

func WithDoShareNonSingletonNodes(doShareNonSingletonNodes bool) BuilderOption {
	return func(option *builderOption) {
		option.doShareNonSingletonNodes = doShareNonSingletonNodes
	}
}

func WithShareMaxTailLength(shareMaxTailLength int) BuilderOption {
	return func(option *builderOption) {
		option.shareMaxTailLength = shareMaxTailLength
	}
}

func WithAllowFixedLengthArcs(allowFixedLengthArcs bool) BuilderOption {
	return func(option *builderOption) {
		option.allowFixedLengthArcs = allowFixedLengthArcs
	}
}

func WithBytesPageBits(bytesPageBits int) BuilderOption {
	return func(option *builderOption) {
		option.bytesPageBits = bytesPageBits
	}
}

// newBuilder
// Instantiates an FST/FSA builder with all the possible tuning and construction tweaks.
// Read parameter documentation carefully.
//
// inputType: The input type (transition labels). Can be anything from Fst.INPUT_TYPE enumeration. Shorter types will consume less memory. Strings (character sequences) are represented as Fst.INPUT_TYPE.BYTE4 (full unicode codepoints).
// minSuffixCount1: If pruning the input graph during construction, this threshold is used for telling if a node is kept or pruned. If transition_count(node) >= minSuffixCount1, the node is kept.
// minSuffixCount2: (Note: only Mike McCandless knows what this one is really doing...)
// doShareSuffix: If true, the shared suffixes will be compacted into unique paths. This requires an additional RAM-intensive hash map for lookups in memory. Setting this parameter to false creates a single suffix path for all input sequences. This will result in a larger FST, but requires substantially less memory and CPU during building.
// doShareNonSingletonNodes: Only used if doShareSuffix is true. Set this to true to ensure FST is fully minimal, at cost of more CPU and more RAM during building.
// shareMaxTailLength: Only used if doShareSuffix is true. Set this to Integer.MAX_VALUE to ensure FST is fully minimal, at cost of more CPU and more RAM during building.
// output: The output type for each input sequence. Applies only if building an FST. For FSA, use NoOutputs.getSingleton() and NoOutputs.getNoOutput() as the singleton output object.
// allowFixedLengthArcs: Pass false to disable the fixed length arc optimization (binary search or direct addressing) while building the FST; this will make the resulting FST smaller but slower to traverse.
// bytesPageBits: How many bits wide to make each byte[] block in the BytesStore; if you know the FST will be large then make this larger. For example 15 bits = 32768 byte pages.
func newBuilder(inputType InputType, outputManager OutputManager, minSuffixCount1, minSuffixCount2 int,
	doShareSuffix, doShareNonSingletonNodes bool, shareMaxTailLength int,
	allowFixedLengthArcs bool, bytesPageBits int) (*Builder, error) {

	builder := &Builder{
		minSuffixCount1:          minSuffixCount1,
		minSuffixCount2:          minSuffixCount2,
		doShareNonSingletonNodes: doShareNonSingletonNodes,
		shareMaxTailLength:       shareMaxTailLength,
		allowFixedLengthArcs:     allowFixedLengthArcs,
		fst:                      NewFST(inputType, outputManager, bytesPageBits),
		frontier:                 make([]*UnCompiledNode, 0, 10),
		fixedLengthArcsBuffer:    NewBuffer(),
		noOutput:                 outputManager.EmptyOutput(),
	}

	builder.bytes = builder.fst.bytes
	// TODO: assert bytes != null;
	// pad: ensure no node gets address 0 which is reserved to mean
	// the stop state w/ no arcs
	if err := builder.bytes.WriteByte(0); err != nil {
		return nil, err
	}

	if doShareSuffix {
		reader, err := builder.bytes.getReverseReader(false)
		if err != nil {
			return nil, err
		}
		builder.dedupHash = NewNodeHash(builder.fst, reader)
	}

	for i := 0; i < 10; i++ {
		node := NewUnCompiledNode(builder, i)
		builder.frontier = append(builder.frontier, node)
	}
	return builder, nil
}

// SetDirectAddressingMaxOversizingFactor
// Overrides the default the maximum oversizing of fixed array
// allowed to enable direct addressing of arcs instead of binary search.
// Setting this factor to a negative value (e.g. -1) effectively disables direct addressing,
// only binary search nodes will be created.
// DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR
func (b *Builder) SetDirectAddressingMaxOversizingFactor(factor float64) *Builder {
	b.directAddressingMaxOversizingFactor = factor
	return b
}

func (b *Builder) GetDirectAddressingMaxOversizingFactor() float64 {
	return b.directAddressingMaxOversizingFactor
}

func (b *Builder) GetTermCount() int {
	return b.frontier[0].InputCount
}

func (b *Builder) GetNodeCount() int {
	// 1+ in order to count the -1 implicit final node
	return b.nodeCount + 1
}

func (b *Builder) GetArcCount() int {
	return b.arcCount
}

func (b *Builder) compileNode(ctx context.Context, nodeIn *UnCompiledNode, tailLength int) (*CompiledNode, error) {
	var node int64
	var err error
	bytesPosStart := b.bytes.GetPosition()
	if b.dedupHash != nil &&
		(b.doShareNonSingletonNodes || nodeIn.NumArcs() <= 1) &&
		tailLength <= b.shareMaxTailLength {

		if nodeIn.NumArcs() == 0 {
			// 尾部节点，直接加入到fst中
			node, err = b.fst.AddNode(ctx, b, nodeIn)
			if err != nil {
				return nil, err
			}
			b.lastFrozenNode = node
		} else {
			// 将节点加入到map中，如果存在后缀复用，会返回复用的node
			node, err = b.dedupHash.Add(ctx, b, nodeIn)
			if err != nil {
				return nil, err
			}
		}
	} else {
		node, err = b.fst.AddNode(ctx, b, nodeIn)
		if err != nil {
			return nil, err
		}
	}

	bytesPosEnd := b.bytes.GetPosition()
	if bytesPosEnd != bytesPosStart {
		// The Fst added a new node:
		b.lastFrozenNode = node
	}

	nodeIn.Clear()

	cNode := NewCompiledNode()
	cNode.node = node
	return cNode, nil
}

// 尾部冻结
// prefixLenPlus1: 相同的前缀长度
func (b *Builder) freezeTail(ctx context.Context, prefixLenPlus1 int) error {
	downTo := max(1, prefixLenPlus1)

	// idx := len(b.lastInput) 因为 b.frontier 长度比 b.lastInput 多一位
	for idx := len(b.lastInput); idx >= downTo; idx-- {
		doPrune := false   // 需要修剪
		doCompile := false // 需要编译

		node := b.frontier[idx]
		parent := b.frontier[idx-1]

		if node.InputCount < b.minSuffixCount1 {
			doPrune = true
			doCompile = true
		} else {
			if idx > prefixLenPlus1 {
				// prune if parent's inputCount is less than suffixMinCount2
				// 如果父级的 inputCount 小于 suffixMinCount2 则修剪
				if parent.InputCount < b.minSuffixCount2 ||
					(b.minSuffixCount2 == 1 && parent.InputCount == 1 && idx > 1) {

					// my parent, about to be compiled, doesn't make the cut,
					// so I'm definitely pruned
					// 我的父父节点，即将被编译，没有裁剪，所以我肯定被修剪了

					// if minSuffixCount2 is 1, we keep only up
					// until the 'distinguished edge', ie we keep only the
					// 'divergent' part of the Fst. if my parent, about to be
					// compiled, has inputCount 1 then we are already past the
					// distinguished edge.  NOTE: this only works if
					// the Fst output are not "compressible" (simple
					// ords ARE compressible).
					//
					// 如果 minSuffixCount2 为 1，我们只保留直到“可区分边缘”，
					// 即我们只保留 Fst 的“发散”部分。如果我的父级（即将编译）的 inputCount 为 1，
					// 那么我们已经超过了显着边缘。注意：这仅在 Fst 输出不可“压缩”（简单命令可压缩）时才有效。
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
		}

		if node.InputCount < b.minSuffixCount2 ||
			(b.minSuffixCount2 == 1 && node.InputCount == 1 && idx > 1) {

			// drop all arcs
			for arcIdx := 0; arcIdx < int(node.NumArcs()); arcIdx++ {
				target, ok := node.Arcs[arcIdx].Target.(*UnCompiledNode)
				if ok {
					target.Clear()
				} else {
					// TODO: 是否需要处理
				}
			}
			node.Arcs = node.Arcs[:0]
		}

		if doPrune {
			// this node doesn't make it -- deref it
			// 这个节点没有成功——取消引用它
			node.Clear()
			if err := parent.DeleteLast(ctx, b.lastInput[idx-1], node); err != nil {
				return err
			}
			continue
		}

		if b.minSuffixCount2 != 0 {
			if err := b.compileAllTargets(ctx, node, len(b.lastInput)-idx); err != nil {
				return err
			}
		}
		nextFinalOutput := node.Output

		// We "fake" the node as being final if it has no
		// outgoing arcs; in theory we could leave it
		// as non-final (the Fst can represent this), but
		// FstEnum, Util, etc., have trouble w/ non-final
		// dead-end states:
		isFinal := node.IsFinal || node.NumArcs() == 0

		if doCompile {
			// this node makes it and we now compile it.
			// first, compile any targets that were previously
			// undecided:
			tailLength := 1 + len(b.lastInput) - idx
			compileNode, err := b.compileNode(ctx, node, tailLength)
			if err != nil {
				return err
			}

			if err := parent.ReplaceLast(b.lastInput[idx-1], compileNode, nextFinalOutput, isFinal); err != nil {
				return err
			}
		} else {
			// replaceLast just to install
			// nextFinalOutput/isFinal onto the arc
			if err := parent.ReplaceLast(b.lastInput[idx-1], node, nextFinalOutput, isFinal); err != nil {
				return err
			}
			// this node will stay in play for now, since we are
			// undecided on whether to prune it.  later, it
			// will be either compiled or pruned, so we must
			// allocate a new node:
			b.frontier[idx] = NewUnCompiledNode(b, idx)
		}
	}

	return nil
}

// Add the next input/output pair. The provided input must be sorted after the previous one according
// to IntsRef.compareTo. It's also OK to add the same input twice in a row with different output,
// as long as Output implements the Output.merge method. Note that input is fully consumed after
// this method is returned (so caller is free to reuse), but output is not. So if your output are
// changeable (eg ByteSequenceOutputs or IntSequenceOutputs) then you cannot reuse across calls.
//
// 添加input/output。提供的input必须要先进行排序。如果输入相同的input+不同的output，output需要实现merge方法。

func (b *Builder) AddStr(ctx context.Context, input string, output Output) error {
	return b.Add(ctx, []rune(input), output)
}

func (b *Builder) Add(ctx context.Context, input []rune, output Output) error {
	newInput := make([]int, len(input))
	for i, v := range input {
		newInput[i] = int(v)
	}
	return b.AddInts(ctx, newInput, output)
}

func (b *Builder) AddInts(ctx context.Context, input []int, output Output) error {
	if output == nil {
		output = b.noOutput
	} else if output.IsNoOutput() {
		output = b.noOutput
	}

	if len(input) == 0 {
		// empty input: only allowed as first input.  we have
		// to special case this because the packed Fst
		// format cannot represent the empty input since
		// 'finalness' is stored on the incoming arc, not on
		// the node
		b.frontier[0].InputCount++
		b.frontier[0].IsFinal = true
		if err := b.fst.SetEmptyOutput(output); err != nil {
			return err
		}
		return nil
	}

	// compare shared prefix length
	pos := 0
	posStop := min(len(b.lastInput), len(input))
	b.frontier[pos].InputCount++
	for pos = 0; pos < posStop; pos++ {
		if b.lastInput[pos] != input[pos] {
			break
		}
		b.frontier[pos].InputCount++
	}

	// 计算公共前缀的长度
	prefixLenPlus1 := pos + 1

	// 如果frontier长度小于输入的长度，进行扩容
	inputLenPlus1 := len(input) + 1
	if len(b.frontier) < inputLenPlus1 {
		frontierSize := len(b.frontier)

		b.frontier = slices.Grow(b.frontier, inputLenPlus1)

		for i := frontierSize; i < inputLenPlus1; i++ {
			b.frontier[i] = NewUnCompiledNode(b, i)
		}
	}

	// minimize/compile states from previous input's orphan'd suffix
	if err := b.freezeTail(ctx, prefixLenPlus1); err != nil {
		return err
	}

	// init tail states for current input
	for idx := prefixLenPlus1; idx <= len(input); idx++ {
		b.frontier[idx-1].AddArc(input[idx-1], b.frontier[idx])
		b.frontier[idx].InputCount++
	}

	lastNode := b.frontier[len(input)]
	if len(b.lastInput) != len(input) || prefixLenPlus1 != len(input)+1 {
		lastNode.IsFinal = true
		lastNode.Output = b.noOutput
	}

	var err error

	// push conflicting output forward, only as far as needed
	// 仅根据需要将冲突的输出向前推进
	for idx := 1; idx < prefixLenPlus1; idx++ {
		node := b.frontier[idx]
		parentNode := b.frontier[idx-1]

		lastOutput := parentNode.GetLastOutput()

		var commonOutputPrefix Output
		var wordSuffix Output

		if !lastOutput.IsNoOutput() {
			commonOutputPrefix, err = output.Common(lastOutput)
			if err != nil {
				return err
			}

			wordSuffix, err = lastOutput.Sub(commonOutputPrefix)
			if err != nil {
				return err
			}

			if err := parentNode.SetLastOutput(ctx, input[idx-1], commonOutputPrefix); err != nil {
				return err
			}

			if err := node.PrependOutput(wordSuffix); err != nil {
				return err
			}
		}

		output, err = output.Sub(commonOutputPrefix)
		if err != nil {
			return err
		}
	}

	if len(b.lastInput) == len(input) && prefixLenPlus1 == 1+len(input) {
		// same input more than 1 time in a row, mapping to
		// multiple output
		mergeOutput, err := lastNode.Output.Merge(output)
		if err != nil {
			return err
		}
		lastNode.Output = mergeOutput
	} else {
		// this new arc is private to this new input; set its
		// arc output to the leftover output:
		if err := b.frontier[prefixLenPlus1-1].SetLastOutput(ctx, input[prefixLenPlus1-1], output); err != nil {
			return err
		}
	}

	// save last input
	b.lastInput = input

	return err
}

// Finish Returns final FST. NOTE: this will return null if nothing is accepted by the FST.
func (b *Builder) Finish(ctx context.Context) (*FST, error) {

	root := b.frontier[0]

	// minimize nodes in the last word's suffix
	if err := b.freezeTail(ctx, 0); err != nil {
		return nil, err
	}

	if root.InputCount < b.minSuffixCount1 || root.InputCount < b.minSuffixCount2 || root.NumArcs() == 0 {
		if b.fst.emptyOutput.IsNoOutput() {
			return nil, nil
		}

		if b.minSuffixCount1 > 0 || b.minSuffixCount2 > 0 {
			// empty string got pruned
			return nil, nil
		}
	} else {
		if b.minSuffixCount2 != 0 {
			if err := b.compileAllTargets(ctx, root, len(b.lastInput)); err != nil {
				return nil, err
			}
		}
	}

	compileNode, err := b.compileNode(ctx, root, len(b.lastInput))
	if err != nil {
		return nil, err
	}

	if err = b.fst.Finish(compileNode.node); err != nil {
		return nil, err
	}

	return b.fst, nil
}

func (b *Builder) compileAllTargets(ctx context.Context, node *UnCompiledNode, tailLength int) error {
	for arcIdx := 0; arcIdx < node.NumArcs(); arcIdx++ {
		arc := node.Arcs[arcIdx]
		if !arc.Target.IsCompiled() {
			// not yet compiled
			n := arc.Target.(*UnCompiledNode)
			if n.NumArcs() == 0 {
				//System.out.println("seg=" + segment + "        FORCE final arc=" + (char) arc.label);
				arc.IsFinal, n.IsFinal = true, true
			}
			target, err := b.compileNode(ctx, n, tailLength-1)
			if err != nil {
				return err
			}
			arc.Target = target
		}
	}
	return nil
}

// DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR
// Default oversizing factor used to decide whether to encode a node with direct addressing or binary search.
// Default is 1: ensure no oversizing on average.
// This factor does not determine whether to encode a node with a list of variable length arcs or with fixed length arcs.
// It only determines the effective encoding of a node that is already known to be encoded with fixed length arcs.
// See Fst.shouldExpandNodeWithFixedLengthArcs() and Fst.shouldExpandNodeWithDirectAddressing().
// For English words we measured 217K nodes, only 3.27% nodes are encoded with fixed length arcs,
// and 99.99% of them with direct addressing. Overall FST memory reduced by 1.67%.
// For worst case we measured 168K nodes, 50% of them are encoded with fixed length arcs,
// and 14% of them with direct encoding. Overall FST memory reduced by 0.8%.
// Use TestFstDirectAddressing.main() and TestFstDirectAddressing.testWorstCaseForDirectAddressing() to evaluate a change.
// see: setDirectAddressingMaxOversizingFactor
const DIRECT_ADDRESSING_MAX_OVERSIZING_FACTOR = 1.0

type Buffer struct {
	bytes.Buffer
	buff []byte
}

func NewBuffer() *Buffer {
	return &Buffer{
		Buffer: bytes.Buffer{},
		buff:   make([]byte, 48),
	}
}

func (b *Buffer) WriteUvarint(i uint64) error {
	num := binary.PutUvarint(b.buff, i)
	_, err := b.Write(b.buff[:num])
	return err
}
