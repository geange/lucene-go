package main

import (
	"fmt"
	"github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
)

func main() {
	dir, err := store.NewNIOFSDirectory("data")
	if err != nil {
		panic(err)
	}

	version := util.NewVersion(8, 11, 0)
	minVersion := util.NewVersion(8, 0, 0)
	segment := index.NewSegmentInfo(dir, version, minVersion, "0", 3,
		false, nil, map[string]string{}, []byte("1"), map[string]string{}, nil)

	fieldInfos := index.NewFieldInfos([]*types.FieldInfo{&types.FieldInfo{}})

	format := simpletext.NewSimpleTextFieldInfosFormat()
	err = format.Write(dir, segment, "", fieldInfos, nil)
	if err != nil {
		panic(err)
	}

	infos, err := format.Read(dir, segment, "", nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(infos.Size())
}
