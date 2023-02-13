package main

import (
	"encoding/binary"
	"github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

func main() {
	fieldInfo := types.NewFieldInfo(
		"field1",
		1,
		false,
		false,
		true,
		types.INDEX_OPTIONS_DOCS_AND_FREQS_AND_POSITIONS_AND_OFFSETS,
		types.DOC_VALUES_TYPE_NONE,
		-1,
		map[string]string{},
		2,
		2,
		4,
		true,
	)
	w := index.NewPointValuesWriter(fieldInfo)
	err := w.AddPackedValue(1, Point(5, 4))
	if err != nil {
		return
	}
	err = w.AddPackedValue(1, Point(5, 5))
	if err != nil {
		return
	}
	err = w.AddPackedValue(2, Point(2, 4))
	if err != nil {
		return
	}
	err = w.AddPackedValue(2, Point(3, 4))
	if err != nil {
		return
	}
	err = w.AddPackedValue(3, Point(7, 4))
	if err != nil {
		return
	}
	err = w.AddPackedValue(3, Point(5, 8))
	if err != nil {
		return
	}

	dir, err := store.NewNIOFSDirectory("data")
	if err != nil {
		panic(err)
	}

	version := util.NewVersion(8, 11, 0)
	minVersion := util.NewVersion(8, 0, 0)
	segment := index.NewSegmentInfo(dir, version, minVersion, "0", 3,
		false, nil, map[string]string{}, []byte("1"), map[string]string{}, nil)

	infos := index.NewFieldInfos([]*types.FieldInfo{fieldInfo})

	writeState := index.NewSegmentWriteState(dir, segment, infos, nil, nil)

	writer, err := simpletext.NewSimpleTextPointsWriter(writeState)
	if err != nil {
		panic(err)
	}

	err = w.Flush(writer)
	if err != nil {
		panic(err)
	}

}

func Point(values ...int) []byte {
	size := 4 * len(values)
	bs := make([]byte, size)
	for i := 0; i < len(values); i++ {
		binary.BigEndian.PutUint32(bs[i*4:], uint32(values[i]))
	}
	return bs
}
