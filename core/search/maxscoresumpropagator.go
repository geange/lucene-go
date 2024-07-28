package search

import (
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"math"
	"sort"
)

// MaxScoreSumPropagator
// Utility class to propagate scoring information in BooleanQuery, which compute the score as the sum of
// the scores of its matching clauses. This helps propagate information about the maximum produced score
// GPT3.5:
// 这段注释描述的是`MaxScoreSumPropagator`作为一个实用工具类的用途。它用于在布尔查询（BooleanQuery）中传播评分信息，
// 其中查询的得分是其匹配子查询得分的累加和。这有助于传播关于生成的最大得分的信息。
//
// 在布尔查询中，可能包含多个子查询（clauses），如"must"（必须匹配）子查询和"should"（可选匹配）子查询。
// 每个子查询都会计算出一个相关性得分，表示文档与该子查询的匹配程度。
//
// `MaxScoreSumPropagator`的作用是将这些子查询的得分信息进行传播。它会遍历所有子查询的匹配文档，
// 对于每个文档，将其得分与之前的最大得分进行比较，并保留最大得分。这样，最终的文档得分就是所有子查询中的最大得分之和。
//
// 通过这种传播机制，`MaxScoreSumPropagator`能够确保布尔查询的文档得分是基于各个子查询中最相关的得分进行计算的。
// 这对于提高搜索结果的质量和排序准确性非常有帮助，因为它可以将最相关的子查询的得分信息传递给最终的文档得分。
//
// 总而言之，`MaxScoreSumPropagator`是一个用于在布尔查询中传播评分信息的实用工具类，用于确保最终文档得分是各个子查询中最相关得分的累加和。
type MaxScoreSumPropagator struct {
	numClauses          int
	scorers             []index.Scorer
	sumOfOtherMaxScores []float64
}

func NewMaxScoreSumPropagator(scorerList []index.Scorer) (*MaxScoreSumPropagator, error) {
	propagator := &MaxScoreSumPropagator{
		numClauses: len(scorerList),
		scorers:    scorerList,
	}

	// We'll need max scores multiple times so we cache them
	maxScores := make([]float64, propagator.numClauses)
	for i := 0; i < propagator.numClauses; i++ {
		_, err := propagator.scorers[i].AdvanceShallow(0)
		if err != nil {
			return nil, err
		}
		maxScores[i], err = propagator.scorers[i].GetMaxScore(types.NO_MORE_DOCS)
		if err != nil {
			return nil, err
		}
	}

	// Sort by decreasing max score
	sort.Sort(InPlaceMergeSorter{
		scorers:   propagator.scorers,
		maxScores: maxScores,
	})

	propagator.sumOfOtherMaxScores = computeSumOfComplement(maxScores)
	return propagator, nil
}

var _ sort.Interface = &InPlaceMergeSorter{}

type InPlaceMergeSorter struct {
	scorers   []index.Scorer
	maxScores []float64
}

func (r InPlaceMergeSorter) Len() int {
	return len(r.maxScores)
}

func (r InPlaceMergeSorter) Less(i, j int) bool {
	return r.maxScores[j] < r.maxScores[i]
}

func (r InPlaceMergeSorter) Swap(i, j int) {
	r.scorers[i], r.scorers[j] = r.scorers[j], r.scorers[i]
	r.maxScores[i], r.maxScores[j] = r.maxScores[j], r.maxScores[i]
}

// Return an array which, at index i, stores the sum of all entries of v except the one at index i.
func computeSumOfComplement(v []float64) []float64 {
	// We do not use subtraction on purpose because it would defeat the
	// upperbound formula that we use for sums.
	// Naive approach would be O(n^2), but we can do O(n) by computing the
	// sum for i<j and i>j and then sum them.
	sum1 := make([]float64, len(v))
	for i := 1; i < len(sum1); i++ {
		sum1[i] = sum1[i-1] + v[i-1]
	}

	sum2 := make([]float64, len(v))
	size := len(sum2) - 2
	for i := size; i >= 0; i-- {
		sum2[i] = sum2[i+1] + v[i+1]
	}

	result := make([]float64, len(v))
	for i := 0; i < len(result); i++ {
		result[i] = sum1[i] + sum2[i]
	}
	return result
}

func (m *MaxScoreSumPropagator) SetMinCompetitiveScore(minScore float64) error {
	if minScore == 0 {
		return nil
	}
	// A double that is less than 'minScore' might still be converted to 'minScore'
	// when casted to a float, so we go to the previous float to avoid this issue
	// minScoreDown := Math.nextDown(minScore);
	minScoreDown := math.Nextafter(minScore, math.Inf(-1))
	for i := 0; i < m.numClauses; i++ {
		sumOfOtherMaxScores := m.sumOfOtherMaxScores[i]
		minCompetitiveScore := m.getMinCompetitiveScore(minScoreDown, sumOfOtherMaxScores)
		if minCompetitiveScore <= 0 {
			// given that scorers are sorted by decreasing max score, next scorers will
			// have 0 as a minimum competitive score too
			break
		}
		err := m.scorers[i].SetMinCompetitiveScore(minCompetitiveScore)
		if err != nil {
			return err
		}
	}
	return nil
}

// Return the minimum score that a Scorer must produce in order for a hit to be competitive.
// GPT:
// 这段代码是一个名为`getMinCompetitiveScore`的私有方法。它用于计算一个`Scorer`必须产生的最低分数，以便判断一个命中是否具有竞争力。
// 以下是该方法的工作原理的解释：
//  1. 首先，它检查`minScoreSum`是否小于或等于`sumOfOtherMaxScores`。如果条件成立，意味着最小分数已经具有竞争力，因此方法返回0。
//  2. 如果上述条件为假，则进入一个循环，以找到满足要求`minScore + sumOfOtherMaxScores <= minScoreSum`的最小分数值。
//  3. 初始的`minScore`值计算为`minScoreSum`减去`sumOfOtherMaxScores`的差值。
//  4. 在循环中，它调用`scoreSumUpperBound`方法，并将当前的`minScore + sumOfOtherMaxScores`作为参数传递进去。
//     这个方法的目的在代码片段中没有提供，因此我们可以假设它在代码库的其他地方定义。
//  5. 如果`scoreSumUpperBound`的结果大于`minScoreSum`，则意味着当前的`minScore`值不足，需要减小。
//  6. 在每次迭代中，`minScore`减去最小可表示的正增量，该增量由`Math.ulp(minScoreSum)`给出。将该值加到`minScore`上，使其变小并接近目标值。
//  7. 循环继续，直到条件`scoreSumUpperBound(minScore + sumOfOtherMaxScores) > minScoreSum`不再成立，表示当前的`minScore`值是满意的。
//  8. 最后，方法返回`minScore`和0之间的较大值（确保最小分数非负）。
//
// 代码中包含一个断言，即迭代次数（`iters`）应最多为2，表示预期循环将在很少的迭代次数内收敛。
//
// 代码片段还包含了一个TODO注释，暗示编写者不确定是否存在更高效的方法来找到所需的`minScore`值。
func (m *MaxScoreSumPropagator) getMinCompetitiveScore(minScoreSum, sumOfOtherMaxScores float64) float64 {
	if minScoreSum <= sumOfOtherMaxScores {
		return 0
	}

	// We need to find a value 'minScore' so that 'minScore + sumOfOtherMaxScores <= minScoreSum'
	// TODO: is there an efficient way to find the greatest value that meets this requirement?
	minScore := minScoreSum - sumOfOtherMaxScores
	iters := 0
	for m.scoreSumUpperBound(minScore+sumOfOtherMaxScores) > minScoreSum {
		// Important: use ulp of minScoreSum and not minScore to make sure that we
		// converge quickly.
		minScore -= minScoreSum
		// this should converge in at most two iterations:
		//  - one because of the subtraction rounding error
		//  - one because of the error introduced by sumUpperBound
		//assert ++iters <= 2 : iters;
		iters++
		if iters <= 2 {
			continue
		}
	}
	return max(minScore, 0)
}

// GPT3.5:
// 以下是该方法的解释：
//
// 1. 首先，它检查`numClauses`是否小于或等于2。如果是这种情况，那么无论顺序如何，总和始终相同。因此，方法直接返回总和的浮点值。
//
// 2. 如果`numClauses`大于2，那么总和的计算误差将取决于值相加的顺序。为了避免这个问题，方法计算总和可能达到的上界。如果最大相对误差为b，那么这意味着两个总和始终在彼此之间相差2*b。
//
// 3. 对于合取操作（conjunctions），我们可以跳过这个误差因子，因为分数相加的顺序是可预测的。但在实践中，这并没有太大帮助，因为这个误差因子引入的差值通常会被浮点数转换所抵消。
//
// 4. 最后，方法返回总和的上界值，计算方式为`(1.0 + 2 * b) * sum`，并将其转换为浮点数。
//
// `getMinCompetitiveScore`方法使用了`scoreSumUpperBound`方法来判断给定的`minScore + sumOfOtherMaxScores`是否满足要求，
// 以便寻找最小的满足条件的`minScore`值。该值用于确定是否达到了具有竞争力的最低分数。
func (m *MaxScoreSumPropagator) scoreSumUpperBound(sum float64) float64 {
	if m.numClauses <= 2 {
		// When there are only two clauses, the sum is always the same regardless
		// of the order.
		return sum
	}

	// The error of sums depends on the order in which values are summed up. In
	// order to avoid this issue, we compute an upper bound of the value that
	// the sum may take. If the max relative error is b, then it means that two
	// sums are always within 2*b of each other.
	// For conjunctions, we could skip this error factor since the order in which
	// scores are summed up is predictable, but in practice, this wouldn't help
	// much since the delta that is introduced by this error factor is usually
	// cancelled by the float cast.
	//double b = MathUtil.sumRelativeErrorBound(numClauses);
	b := util.SumRelativeErrorBound(m.numClauses)
	return (1.0 + 2*b) * sum
}
