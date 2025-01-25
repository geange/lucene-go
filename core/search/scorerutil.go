package search

import "github.com/geange/lucene-go/core/util/structure"

func costWithMinShouldMatch(costs []int64, numScorers, minShouldMatch int) int64 {
	// the idea here is the following: a boolean query c1,c2,...cn with minShouldMatch=m
	// could be rewritten to:
	// (c1 AND (c2..cn|msm=m-1)) OR (!c1 AND (c2..cn|msm=m))
	// if we assume that clauses come in ascending cost, then
	// the cost of the first part is the cost of c1 (because the cost of a conjunction is
	// the cost of the least costly clause)
	// the cost of the second part is the cost of finding m matches among the c2...cn
	// remaining clauses
	// since it is a disjunction overall, the total cost is the sum of the costs of these
	// two parts

	// 这里的想法如下：一个布尔查询c1，c2，。。。minShouldMatch=m的cn可以重写为：
	//（c1 AND（c2..cn|msm=m-1））OR（！c1 AND（c2..cn|msm=m）），
	// 如果我们假设子句以升序出现，那么第一部分的成本是c1的成本（因为连词的成本是成本最低的子句的成本）
	// 第二部分的成本就是在剩余的c2…cn子句中找到m个匹配项的成本，因为它总体上是一个析取，所以总成本是这两个部分的成本之和

	// If we recurse infinitely, we find out that the cost of a msm query is the sum of the
	// costs of the num_scorers - minShouldMatch + 1 least costly scorers
	maxSize := numScorers - minShouldMatch + 1
	pq := structure.NewPriorityQueue[int64](maxSize, func(a, b int64) bool {
		return a > b
	})

	for _, cost := range costs {
		pq.InsertWithOverflow(cost)
	}

	sum := int64(0)
	for v := range pq.Iterator() {
		sum += v
	}
	return sum
}
