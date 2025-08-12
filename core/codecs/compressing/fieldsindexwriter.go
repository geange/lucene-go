package compressing

import (
	"context"
	"errors"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util/packed"
)

// FieldsIndexWriter
// Efficient index format for block-based Codecs.
// For each block of compressed stored fields, this stores the first document of the block and the start
// pointer of the block in a DirectMonotonicWriter. At read time, the docID is binary-searched in the
// DirectMonotonicReader that records doc IDS, and the returned index is used to look up the start pointer
// in the DirectMonotonicReader that records start pointers.
type FieldsIndexWriter struct {
	dir             store.Directory
	name            string
	suffix          string
	extension       string
	codecName       string
	id              []byte
	blockShift      int
	ioContext       *store.IOContext
	docsOut         store.IndexOutput
	filePointersOut store.IndexOutput
	totalDocs       int
	totalChunks     int
	previousFP      int64
}

func NewFieldsIndexWriter(ctx context.Context, dir store.Directory, name, suffix, extension,
	codecName string, id []byte, blockShift int, ioContext *store.IOContext) *FieldsIndexWriter {
	docsOut, err := dir.CreateTempOutput(ctx, name, codecName+"-doc_ids")
	if err != nil {
		return nil
	}

	err = codecs.WriteHeader(ctx, docsOut, codecName+"Docs", FIELDS_VERSION_CURRENT)
	if err != nil {
		return nil
	}

	filePointersOut, err := dir.CreateTempOutput(ctx, name, codecName+"file_pointers")
	if err != nil {
		return nil
	}
	err = codecs.WriteHeader(ctx, filePointersOut, codecName+"FilePointers", FIELDS_VERSION_CURRENT)
	if err != nil {
		return nil
	}

	return &FieldsIndexWriter{
		dir:             dir,
		name:            name,
		suffix:          suffix,
		extension:       extension,
		codecName:       codecName,
		id:              id,
		blockShift:      blockShift,
		ioContext:       ioContext,
		docsOut:         docsOut,
		filePointersOut: filePointersOut,
	}
}

func (f *FieldsIndexWriter) writeIndex(ctx context.Context, numDocs int, startPointer int64) error {
	if err := f.docsOut.WriteUvarint(ctx, uint64(numDocs)); err != nil {
		return err
	}
	if err := f.filePointersOut.WriteUvarint(ctx, uint64(startPointer-f.previousFP)); err != nil {
		return err
	}
	f.previousFP = startPointer
	f.totalDocs += numDocs
	f.totalChunks++
	return nil
}

func (f *FieldsIndexWriter) finish(ctx context.Context, numDocs int, maxPointer int64, metaOut store.IndexOutput) error {
	if numDocs != f.totalDocs {
		return errors.New("illegal state exception")
	}
	if err := codecs.WriteFooter(ctx, f.docsOut); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, f.filePointersOut); err != nil {
		return err
	}
	if err := f.docsOut.Close(); err != nil {
		return err
	}
	if err := f.filePointersOut.Close(); err != nil {
		return err
	}

	dataOut, err := f.dir.CreateOutput(ctx, store.SegmentFileName(f.name, f.suffix, f.extension))
	if err != nil {
		return err
	}

	if err := codecs.WriteIndexHeader(ctx, dataOut, f.codecName+"Idx", FIELDS_VERSION_CURRENT, f.id, f.suffix); err != nil {
		return err
	}
	if err := metaOut.WriteUint32(ctx, uint32(numDocs)); err != nil {
		return err
	}
	if err := metaOut.WriteUint32(ctx, uint32(f.blockShift)); err != nil {
		return err
	}
	if err := metaOut.WriteUint32(ctx, uint32(f.totalChunks+1)); err != nil {
		return err
	}
	if err := metaOut.WriteUint64(ctx, uint64(dataOut.GetFilePointer())); err != nil {
		return err
	}

	docsIn, err := store.OpenChecksumInput(ctx, f.dir, f.docsOut.GetName())
	if err != nil {
		return err
	}
	if _, err := codecs.CheckHeader(ctx, docsIn, f.codecName+"Docs", FIELDS_VERSION_CURRENT, FIELDS_VERSION_CURRENT); err != nil {
		return err
	}

	docs, err := packed.DirectMonotonicWriterGetInstance(metaOut, dataOut, int64(f.totalChunks+1), f.blockShift)
	if err != nil {
		return err
	}
	doc := uint64(0)
	if err := docs.Add(int64(doc)); err != nil {
		return err
	}
	for i := 0; i < f.totalChunks; i++ {
		num, err := docsIn.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		doc += num
		if err := docs.Add(int64(doc)); err != nil {
			return err
		}
	}
	if err := docs.Finish(); err != nil {
		return err
	}
	if doc != uint64(f.totalDocs) {
		return errors.New("docs don't add up")
	}
	if _, err := codecs.CheckFooter(ctx, docsIn); err != nil {
		return err
	}

	if err := f.dir.DeleteFile(ctx, f.docsOut.GetName()); err != nil {
		return err
	}
	f.docsOut = nil

	if err := metaOut.WriteUint64(ctx, uint64(dataOut.GetFilePointer())); err != nil {
		return err
	}

	filePointersIn, err := store.OpenChecksumInput(ctx, f.dir, f.filePointersOut.GetName())
	if err != nil {
		return err
	}
	if _, err := codecs.CheckHeader(ctx, filePointersIn,
		f.codecName+"FilePointers", FIELDS_VERSION_CURRENT, FIELDS_VERSION_CURRENT); err != nil {
		return err
	}
	filePointers, err := packed.DirectMonotonicWriterGetInstance(metaOut, dataOut, int64(f.totalChunks+1), f.blockShift)
	if err != nil {
		return err
	}
	fp := uint64(0)
	for i := 0; i < f.totalChunks; i++ {
		num, err := filePointersIn.ReadUvarint(ctx)
		if err != nil {
			return err
		}
		fp += num
		if err := filePointers.Add(int64(fp)); err != nil {
			return err
		}
	}
	if maxPointer < int64(fp) {
		return errors.New("file pointers don't add up")
	}
	if err := filePointers.Add(maxPointer); err != nil {
		return err
	}
	if err := filePointers.Finish(); err != nil {
		return err
	}
	if _, err := codecs.CheckFooter(ctx, filePointersIn); err != nil {
		return err
	}

	if err := f.dir.DeleteFile(ctx, f.filePointersOut.GetName()); err != nil {
		return err
	}
	f.filePointersOut = nil

	if err := metaOut.WriteUint64(ctx, uint64(dataOut.GetFilePointer())); err != nil {
		return err
	}
	if err := metaOut.WriteUint64(ctx, uint64(maxPointer)); err != nil {
		return err
	}

	if err := codecs.WriteFooter(ctx, dataOut); err != nil {
		return err
	}

	return nil
}

func CorruptIndexException(s string, in store.ChecksumIndexInput) {

}

func (f *FieldsIndexWriter) Close() error {

	err := func() error {
		if err := f.docsOut.Close(); err != nil {
			return err
		}
		if err := f.filePointersOut.Close(); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		fileNames := make([]string, 0)
		if f.docsOut != nil {
			fileNames = append(fileNames, f.docsOut.GetName())
		}
		if f.filePointersOut != nil {
			fileNames = append(fileNames, f.filePointersOut.GetName())
		}
		for _, name := range fileNames {
			if err := f.dir.DeleteFile(context.Background(), name); err != nil {
				return err
			}
		}
		f.docsOut = nil
		f.filePointersOut = nil
	}
	return nil
}
