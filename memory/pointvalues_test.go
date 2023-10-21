package memory

import (
	"testing"
)

func TestNewMemoryIndexPointValues(t *testing.T) {
	//set := analysis.NewCharArraySet()
	//set.Add(" ")
	//set.Add("\n")
	//set.Add("\t")
	//analyzer := standard.NewAnalyzer(set)
	//
	//memIndex, err := NewIndex(WithStorePayloads(true))
	//assert.Nil(t, err)
	//
	//points1, err := document.NewBinaryPoint("dim1", []byte{0, 0, 1}, []byte{0, 0, 2}, []byte{0, 0, 4})
	//err = memIndex.AddIndexAbleField(points1, nil)
	//
	//points2, err := document.NewBinaryPoint("dim2", []byte{0, 0, 1}, []byte{0, 0, 2}, []byte{0, 0, 5})
	//err = memIndex.AddIndexAbleField(points2, nil)
	//memIndex.Freeze()
	//
	//query, err := search.NewPointInSetQuery("dim1", 3, 3, [][]byte{{0, 0, 1}, {0, 0, 2}, {0, 0, 4}})
	//assert.Nil(t, err)
	//score := memIndex.Search(query)
	//assert.True(t, score > 0)
	//
	//points, err := document.NewLongPoint("name", 1, 2, 3)
	//
	//fInfo, err := memIndex.getInfo(points.Name(), points.FieldType())
	//
	//pointValues := newMemoryIndexPointValues(fInfo)
	//
	//count, err := pointValues.EstimatePointCount(nil)
	//assert.Nil(t, err)
	//assert.Equal(t, 1, count)
}
