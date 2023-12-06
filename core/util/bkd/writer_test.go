package bkd

import (
	"context"
	"io"
	"math/rand"
	"testing"
	"time"

	"encoding/binary"
	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
)

func TestWriterWriteField(t *testing.T) {
	doWriteField(t, 2000, 1, 1, 8)
	doWriteField(t, 2000, 2, 2, 8)
	doWriteField(t, 2000, 4, 4, 8)

	doWriteField(t, 4000, 1, 1, 8)
	doWriteField(t, 4000, 2, 2, 8)
	doWriteField(t, 4000, 4, 4, 8)

	doWriteField(t, 8000, 1, 1, 8)
	doWriteField(t, 8000, 2, 2, 8)
	doWriteField(t, 8000, 4, 4, 8)
}

func TestWriterReader2Dim(t *testing.T) {
	numDocs, numDims, numIndexDims, bytesPerDim := 20, 2, 2, 8

	err := mkEmptyDir("./test")
	assert.Nil(t, err)

	path := "./test"
	dir, err := store.NewNIOFSDirectory(path)
	assert.Nil(t, err)

	config, err := NewConfig(numDims, numIndexDims, bytesPerDim, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	assert.Nil(t, err)

	w, err := NewWriter(numDocs, dir, "_0", config, DEFAULT_MAX_MB_SORT_IN_HEAP, 26*numDocs)
	assert.Nil(t, err)

	counter := 0
	packedBytes := make([]byte, numDims*bytesPerDim)
	for docID := 0; docID < numDocs; docID++ {
		for j := 0; j < 26; j++ {
			binary.BigEndian.PutUint64(packedBytes, uint64((docID*j)%256))
			binary.BigEndian.PutUint64(packedBytes[8:], uint64((docID*j+1)%256))

			err := w.Add(packedBytes, docID)
			assert.Nil(t, err)
			counter++
		}
	}

	out, err := dir.CreateOutput("1d.bkd", nil)
	assert.Nil(t, err)

	finalizer, err := w.Finish(nil, out, out, out)
	assert.Nil(t, err)

	indexFP := out.GetFilePointer()

	err = finalizer(context.Background())
	assert.Nil(t, err)

	err = out.Close()
	assert.Nil(t, err)

	in, err := dir.OpenInput("1d.bkd", nil)
	assert.Nil(t, err)

	_, err = in.Seek(indexFP, io.SeekStart)
	assert.Nil(t, err)

	reader, err := NewReader(nil, in, in, in)
	assert.Nil(t, err)

	visitor, err := NewVerifyPointsVisitor("1d", numDocs, reader)
	assert.Nil(t, err)

	err = reader.Intersect(nil, visitor)
	assert.Nil(t, err)

	err = in.Close()
	assert.Nil(t, err)
	err = dir.Close()
	assert.Nil(t, err)
}

func TestJustWriter(t *testing.T) {
	numDocs, numDims, numIndexDims, bytesPerDim := 100, 2, 2, 4

	err := mkEmptyDir("./test")
	assert.Nil(t, err)

	path := "./test"
	dir, err := store.NewNIOFSDirectory(path)
	assert.Nil(t, err)

	config, err := NewConfig(numDims, numIndexDims, bytesPerDim, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	assert.Nil(t, err)

	w, err := NewWriter(numDocs, dir, "_0", config, DEFAULT_MAX_MB_SORT_IN_HEAP, 26*numDocs)
	assert.Nil(t, err)

	counter := 0
	packedBytes := make([]byte, numDims*bytesPerDim)
	for docID := 0; docID < numDocs; docID++ {
		for j := 0; j < 26; j++ {
			//nextBytes(rand.NewSource(time.Now().UnixNano()), packedBytes)

			binary.BigEndian.PutUint32(packedBytes, uint32((docID*j)%256))
			binary.BigEndian.PutUint32(packedBytes[4:], uint32((docID*j+1)%256))

			err := w.Add(packedBytes, docID)
			assert.Nil(t, err)
			counter++
		}
	}

	out, err := dir.CreateOutput("1d.bkd", nil)
	assert.Nil(t, err)

	finalizer, err := w.Finish(nil, out, out, out)
	assert.Nil(t, err)

	fp := out.GetFilePointer()
	t.Log(fp)

	err = finalizer(context.Background())
	assert.Nil(t, err)

	err = out.Close()
	assert.Nil(t, err)

}

func TestWriterReaderForDebug(t *testing.T) {
	numDocs, numDims, numIndexDims, bytesPerDim := 100, 2, 2, 4

	err := mkEmptyDir("./test")
	assert.Nil(t, err)

	path := "./test"
	dir, err := store.NewNIOFSDirectory(path)
	assert.Nil(t, err)

	config, err := NewConfig(numDims, numIndexDims, bytesPerDim, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	assert.Nil(t, err)

	w, err := NewWriter(numDocs, dir, "_0", config, DEFAULT_MAX_MB_SORT_IN_HEAP, 26*numDocs)
	assert.Nil(t, err)

	counter := 0
	packedBytes := make([]byte, numDims*bytesPerDim)
	for docID := 0; docID < numDocs; docID++ {
		for j := 0; j < 26; j++ {
			binary.BigEndian.PutUint32(packedBytes, uint32((docID*j)%256))
			binary.BigEndian.PutUint32(packedBytes[4:], uint32((docID*j+1)%256))

			err := w.Add(packedBytes, docID)
			assert.Nil(t, err)
			counter++
		}
	}

	out, err := dir.CreateOutput("1d.bkd", nil)
	assert.Nil(t, err)

	finalizer, err := w.Finish(nil, out, out, out)
	assert.Nil(t, err)

	fp := out.GetFilePointer()
	err = finalizer(context.Background())
	assert.Nil(t, err)

	err = out.Close()
	assert.Nil(t, err)
	err = dir.Close()
	assert.Nil(t, err)

	in, err := dir.OpenInput("1d.bkd", nil)
	assert.Nil(t, err)

	_, err = in.Seek(fp, io.SeekStart)
	assert.Nil(t, err)

	reader, err := NewReader(nil, in, in, in)
	assert.Nil(t, err)

	visitor, err := NewVerifyPointsVisitor("1d", numDocs, reader)
	assert.Nil(t, err)

	err = reader.Intersect(nil, visitor)
	assert.Nil(t, err)

	err = in.Close()
	assert.Nil(t, err)
	err = dir.Close()
	assert.Nil(t, err)
}

func doWriteField(t *testing.T, numDocs, numDims, numIndexDims, bytesPerDim int) {
	err := mkEmptyDir("./test")
	assert.Nil(t, err)

	path := "./test"
	dir, err := store.NewNIOFSDirectory(path)
	assert.Nil(t, err)

	config, err := NewConfig(numDims, numIndexDims, bytesPerDim, DEFAULT_MAX_POINTS_IN_LEAF_NODE)
	assert.Nil(t, err)

	w, err := NewWriter(numDocs, dir, "_0", config, DEFAULT_MAX_MB_SORT_IN_HEAP, 26*numDocs)
	assert.Nil(t, err)

	counter := 0
	packedBytes := make([]byte, numDims*bytesPerDim)
	for docID := 0; docID < numDocs; docID++ {
		for j := 0; j < 26; j++ {
			nextBytes(rand.NewSource(time.Now().Unix()), packedBytes)
			err := w.Add(packedBytes, docID)
			assert.Nil(t, err)
			counter++
		}
	}

	out, err := dir.CreateOutput("1d.bkd", nil)
	assert.Nil(t, err)

	finalizer, err := w.Finish(nil, out, out, out)
	assert.Nil(t, err)

	//indexFP := out.GetFilePointer()
	err = finalizer(context.Background())
	assert.Nil(t, err)

	out.Close()
}
