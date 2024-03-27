package bkd

import (
	"errors"
	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestNewOfflinePointWriter(t *testing.T) {
	path := "./test"

	config, err := getRandomConfig()
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	err = mkEmptyDir(path)
	assert.Nil(t, err)

	dir, err := store.NewNIOFSDirectory(path)
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	values := make([][]byte, 0)

	size := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(15000)
	writer := NewOfflinePointWriter(config, dir, "test", "data", size)
	for docId := 0; docId < size; docId++ {
		packedValue := getPackedValue(config)
		values = append(values, packedValue)
		err := writer.Append(nil, packedValue, docId)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
	}
	writer.Close()

	reader, err := writer.GetReader(0, size)
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	readValues := make([][]byte, 0)
	idx := 0
	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Error(err)
			t.FailNow()
		}
		if !next {
			break
		}

		bs := reader.PointValue().PackedValue()
		readValues = append(readValues, slices.Clone(bs))
		idx++
	}

	assert.Equal(t, len(values), len(readValues))
	for i, value := range values {
		assert.Equal(t, value, readValues[i])
	}
}

func TestNewOfflinePointOneDimWriter(t *testing.T) {
	path := "./test"

	config, err := getOneDimConfig()
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	err = mkEmptyDir(path)
	assert.Nil(t, err)

	dir, err := store.NewNIOFSDirectory(path)
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	values := make([][]byte, 0)

	size := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(15000)
	writer := NewOfflinePointWriter(config, dir, "test", "data", size)
	for docId := 0; docId < size; docId++ {
		packedValue := getPackedValue(config)
		values = append(values, packedValue)
		err := writer.Append(nil, packedValue, docId)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
	}
	writer.Close()

	reader, err := writer.GetReader(0, size)
	if !assert.Nil(t, err) {
		t.FailNow()
	}

	readValues := make([][]byte, 0)
	idx := 0
	for {
		next, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Error(err)
			t.FailNow()
		}
		if !next {
			break
		}

		bs := reader.PointValue().PackedValue()
		readValues = append(readValues, slices.Clone(bs))
		idx++
	}

	assert.Equal(t, len(values), len(readValues))
	for i, value := range values {
		assert.Equal(t, value, readValues[i])
	}
}

func mkEmptyDir(name string) error {
	entries, err := os.ReadDir(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return os.Mkdir(name, 0755)
		}
		return err
	}

	for _, entry := range entries {
		err := os.RemoveAll(filepath.Join(name, entry.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}
