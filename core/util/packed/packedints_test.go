package packed

import (
	"context"
	"github.com/geange/lucene-go/core/store"
	"github.com/stretchr/testify/assert"
	"io"
	"math/rand"
	"testing"
	"time"
)

func TestGetMutable(t *testing.T) {
	t.Run("num=100", func(t *testing.T) {
		num := 100

		m := GetMutable(num, 64, 0)
		for i := 0; i < num; i++ {
			m.Set(i, uint64(i+2))
		}

		for i := 0; i < num; i++ {
			v, _ := m.Get(i)
			assert.EqualValues(t, i+2, v)
		}
	})

	t.Run("num=1000", func(t *testing.T) {
		num := 1000

		m := GetMutable(num, 64, 0)
		for i := 0; i < num; i++ {
			m.Set(i, uint64(i+2))
		}

		for i := 0; i < num; i++ {
			v, _ := m.Get(i)
			assert.EqualValues(t, i+2, v)
		}
	})

	t.Run("num=10000", func(t *testing.T) {
		num := 10000

		m := GetMutable(num, 64, 0)
		for i := 0; i < num; i++ {
			m.Set(i, uint64(i+2))
		}

		for i := 0; i < num; i++ {
			v, _ := m.Get(i)
			assert.EqualValues(t, i+2, v)
		}
	})

	t.Run("num=100000", func(t *testing.T) {
		num := 100000

		m := GetMutable(num, 64, 0)
		for i := 0; i < num; i++ {
			m.Set(i, uint64(i+2))
		}

		for i := 0; i < num; i++ {
			v, _ := m.Get(i)
			assert.EqualValues(t, i+2, v)
		}
	})
}

func TestEndPointer(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	valueCount := 1 + r.Intn(999)

	out := store.NewBufferDataOutput()
	for i := 0; i < valueCount; i++ {
		err := out.WriteUint64(context.TODO(), 0)
		assert.Nil(t, err)
	}

	formats := []Format{FormatPacked, FormatPackedSingleBlock}

	in := store.NewBytesInput(out.Bytes())

	for version := VERSION_START; version <= VERSION_CURRENT; version++ {
		for bpv := 11; bpv <= 64; bpv++ {
			for _, format := range formats {
				if !format.IsSupported(bpv) {
					continue
				}

				byteCount := format.ByteCount(version, valueCount, bpv)

				// test iterator
				_, err := in.Seek(0, io.SeekStart)
				assert.Nil(t, err)
				directReader, err := getDirectReaderNoHeader(context.TODO(), in, format, version, valueCount, bpv)
				assert.Nil(t, err)
				_, err = directReader.Get(valueCount - 1)
				assert.Nil(t, err)
				assert.EqualValuesf(t, byteCount, in.GetFilePointer(), "valueCount=%d,bpv=%d", valueCount, bpv)

				// test reader
				_, err = in.Seek(0, io.SeekStart)
				assert.Nil(t, err)
				reader, err := getReaderNoHeader(context.TODO(), in, format, version, valueCount, bpv)
				_, err = reader.Get(valueCount - 1)
				assert.Nil(t, err)
				assert.EqualValuesf(t, byteCount, in.GetFilePointer(), "valueCount=%d,bpv=%d", valueCount, bpv)
			}
		}
	}
}
