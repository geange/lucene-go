package main

import (
	"fmt"
	"github.com/geange/lucene-go/codecs/simpletext"
	"github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/types"
	"github.com/geange/lucene-go/core/util"
	"io"
	"strconv"
)

func main() {
	NUM_TERMS := 100
	terms := make([]*index.TermData, 0, NUM_TERMS)
	for i := 0; i < NUM_TERMS; i++ {
		docs := []int{i}
		text := strconv.FormatInt(int64(i), 36)

		terms = append(terms, index.NewTermData(text, docs, nil))
	}

	builder := index.NewFieldInfosBuilder(index.NewFieldNumbers(""))
	field, err := index.NewFieldData("field", builder, terms, true, true)
	if err != nil {
		panic(err)
	}

	fields := []*index.FieldData{field}
	fieldInfos := builder.Finish()

	dir, err := store.NewNIOFSDirectory("data")
	if err != nil {
		panic(err)
	}

	version := util.NewVersion(8, 11, 0)
	minVersion := util.NewVersion(8, 0, 0)
	segment := index.NewSegmentInfo(dir, version, minVersion, "0", 10000,
		false, nil, map[string]string{}, []byte("1"), map[string]string{}, nil)

	state := index.NewSegmentWriteState(dir, segment, fieldInfos, nil, nil)

	writer, err := simpletext.NewSimpleTextFieldsWriter(state)
	if err != nil {
		panic(err)
	}

	err = writer.Write(index.NewDataFields(fields), NewMyNormsProducer(segment))
	if err != nil {
		panic(err)
	}

	err = writer.Close()

	if err != nil {
		return
	}

	readState := index.NewSegmentReadState(dir, segment, fieldInfos, nil, "")
	reader, err := simpletext.NewSimpleTextFieldsReader(readState)
	if err != nil {
		panic(err)
	}
	term, err := reader.Terms("field")
	if err != nil {
		panic(err)
	}

	{
		iterator, err := term.Iterator()
		if err != nil {
			panic(err)
		}

		ok, err := iterator.SeekExact([]byte("11"))
		if err != nil {
			panic(err)
		}
		fmt.Println(ok)
	}

	{
		iterator, err := term.Iterator()
		if err != nil {
			panic(err)
		}

		ok, err := iterator.SeekExact([]byte("89"))
		if err != nil {
			panic(err)
		}
		fmt.Println(ok)
	}

	{
		iterator, err := term.Iterator()
		if err != nil {
			panic(err)
		}

		ok, err := iterator.SeekExact([]byte("5"))
		if err != nil {
			panic(err)
		}
		fmt.Println(ok)
	}
}

var _ index.NormsProducer = &MyNormsProducer{}

type MyNormsProducer struct {
	si *index.SegmentInfo
}

func (m *MyNormsProducer) Close() error {
	return m.si.Dir().Close()
}

func NewMyNormsProducer(si *index.SegmentInfo) *MyNormsProducer {
	return &MyNormsProducer{si: si}
}

func (m *MyNormsProducer) GetNorms(field *types.FieldInfo) (index.NumericDocValues, error) {
	return NewMyNumericDocValues(m.si), nil
}

func (m *MyNormsProducer) CheckIntegrity() error {
	return nil
}

func (m *MyNormsProducer) GetMergeInstance() index.NormsProducer {
	return m
}

var _ index.NumericDocValues = &MyNumericDocValues{}

type MyNumericDocValues struct {
	doc int
	si  *index.SegmentInfo
}

func NewMyNumericDocValues(si *index.SegmentInfo) *MyNumericDocValues {
	values := &MyNumericDocValues{
		doc: -1,
		si:  si,
	}
	return values
}

func (m *MyNumericDocValues) DocID() int {
	return m.doc
}

func (m *MyNumericDocValues) NextDoc() (int, error) {
	return m.Advance(m.doc + 1)
}

func (m *MyNumericDocValues) Advance(target int) (int, error) {
	maxDoc, err := m.si.MaxDoc()
	if err != nil {
		return 0, err
	}
	if target >= maxDoc {
		return 0, io.EOF
	} else {
		m.doc = target
		return target, nil
	}
}

func (m *MyNumericDocValues) SlowAdvance(target int) (int, error) {
	return index.SlowAdvance(m, target)
}

func (m *MyNumericDocValues) Cost() int64 {
	n, _ := m.si.MaxDoc()
	return int64(n)
}

func (m *MyNumericDocValues) AdvanceExact(target int) (bool, error) {
	m.doc = target
	return true, nil
}

func (m *MyNumericDocValues) LongValue() (int64, error) {
	return 1, nil
}
