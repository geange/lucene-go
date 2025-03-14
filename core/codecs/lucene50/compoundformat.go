package lucene50

import (
	"context"

	"github.com/geange/lucene-go/core/codecs"
	index2 "github.com/geange/lucene-go/core/index"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
)

var _ index.CompoundFormat = &CompoundFormat{}

const (
	COMPOUND_FORMAT_DATA_EXTENSION    = "cfs"
	COMPOUND_FORMAT_ENTRIES_EXTENSION = "cfe"
	COMPOUND_FORMAT_DATA_CODEC        = "Lucene50CompoundData"
	COMPOUND_FORMAT_ENTRY_CODEC       = "Lucene50CompoundEntries"
	COMPOUND_FORMAT_VERSION_START     = 0
	COMPOUND_FORMAT_VERSION_CURRENT   = COMPOUND_FORMAT_VERSION_START
)

type CompoundFormat struct {
}

func (f *CompoundFormat) GetCompoundReader(ctx context.Context, dir store.Directory, si index.SegmentInfo, context *store.IOContext) (index.CompoundDirectory, error) {
	return NewCompoundReader(ctx, dir, si)
}

func (f *CompoundFormat) Write(ctx context.Context, dir store.Directory, si index.SegmentInfo, ioContext *store.IOContext) error {
	dataFile := store.SegmentFileName(si.Name(), "", COMPOUND_FORMAT_DATA_EXTENSION)
	entriesFile := store.SegmentFileName(si.Name(), "", COMPOUND_FORMAT_ENTRIES_EXTENSION)

	data, err := dir.CreateOutput(ctx, dataFile)
	if err != nil {
		return err
	}
	entries, err := dir.CreateOutput(ctx, entriesFile)
	if err != nil {
		return err
	}
	if err := codecs.WriteIndexHeader(ctx, data, COMPOUND_FORMAT_DATA_CODEC,
		COMPOUND_FORMAT_VERSION_CURRENT, si.GetID(), ""); err != nil {
		return err
	}
	if err := codecs.WriteIndexHeader(ctx, entries, COMPOUND_FORMAT_ENTRY_CODEC,
		COMPOUND_FORMAT_VERSION_CURRENT, si.GetID(), ""); err != nil {
		return err
	}

	// write number of files
	files := si.Files()
	if err := entries.WriteUvarint(ctx, uint64(len(files))); err != nil {
		return err
	}

	for file := range files {
		// write bytes for file
		startOffset := data.GetFilePointer()

		in, err := store.OpenChecksumInput(ctx, dir, file)
		if err != nil {
			return err
		}

		// just copies the index header, verifying that its id matches what we expect
		if err := codecs.VerifyAndCopyIndexHeader(ctx, in, data, si.GetID()); err != nil {
			return err
		}
		// copy all bytes except the footer
		numBytesToCopy := int(in.Length()) - codecs.FooterLength() - int(in.GetFilePointer())
		if err := data.CopyBytes(ctx, in, numBytesToCopy); err != nil {
			return err
		}

		// verify footer (checksum) matches for the incoming file we are copying
		checksum, err := codecs.CheckFooter(ctx, in)
		if err != nil {
			return err
		}

		// this is poached from CodecUtil.writeFooter, but we need to use our own checksum, not data.getChecksum(), but I think
		// adding a public method to CodecUtil to do that is somewhat dangerous:
		if err := data.WriteUint32(ctx, codecs.FOOTER_MAGIC); err != nil {
			return err
		}
		if err := data.WriteUint32(ctx, 0); err != nil {
			return err
		}
		if err := data.WriteUint64(ctx, uint64(checksum)); err != nil {
			return err
		}

		endOffset := data.GetFilePointer()

		length := endOffset - startOffset

		// write entry for file
		if err := entries.WriteString(ctx, index2.StripSegmentName(file)); err != nil {
			return err
		}
		if err := entries.WriteUint64(ctx, uint64(startOffset)); err != nil {
			return err
		}
		if err := entries.WriteUint64(ctx, uint64(length)); err != nil {
			return err
		}
	}
	if err := codecs.WriteFooter(ctx, data); err != nil {
		return err
	}
	if err := codecs.WriteFooter(ctx, entries); err != nil {
		return err
	}
	return nil
}
