package blocktree

import (
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"sort"

	"github.com/samber/lo"

	"github.com/geange/lucene-go/core/codecs"
	"github.com/geange/lucene-go/core/codecs/types"
	"github.com/geange/lucene-go/core/document"
	"github.com/geange/lucene-go/core/interface/index"
	"github.com/geange/lucene-go/core/store"
	"github.com/geange/lucene-go/core/util"
	"github.com/geange/lucene-go/core/util/fst"
)

const (
	OUTPUT_FLAGS_NUM_BITS = 2
	OUTPUT_FLAGS_MASK     = 0x3
	OUTPUT_FLAG_IS_FLOOR  = 0x1
	OUTPUT_FLAG_HAS_TERMS = 0x2

	TERMS_EXTENSION             = "tim"
	TERMS_CODEC_NAME            = "BlockTreeTermsDict"
	VERSION_START               = 3
	VERSION_META_LONGS_REMOVED  = 4
	VERSION_COMPRESSED_SUFFIXES = 5
	VERSION_META_FILE           = 6
	VERSION_CURRENT             = VERSION_META_FILE
	TERMS_INDEX_EXTENSION       = "tip"
	TERMS_INDEX_CODEC_NAME      = "BlockTreeTermsIndex"
	TERMS_META_EXTENSION        = "tmd"
	TERMS_META_CODEC_NAME       = "BlockTreeTermsMeta"
)

var _ index.FieldsProducer = &TermsReader{}

type TermsReader struct {
	termsIn        store.IndexInput // Open input to the main terms dict file (_X.tib)
	indexIn        store.IndexInput // Open input to the terms index file (_X.tip)
	postingsReader types.PostingsReader
	fieldMap       map[string]*FieldReader
	fieldList      []string
	segment        string
	version        int
}

func NewTermsReader(ctx context.Context, postingsReader types.PostingsReader,
	state *index.SegmentReadState) (*TermsReader, error) {

	this := &TermsReader{
		postingsReader: postingsReader,
		segment:        state.SegmentInfo.Name(),
	}

	termsName := store.SegmentFileName(this.segment, state.SegmentSuffix, TERMS_EXTENSION)
	termsIn, err := state.Directory.OpenInput(ctx, termsName)
	if err != nil {
		return nil, err
	}
	this.termsIn = termsIn

	version, err := codecs.CheckIndexHeader(ctx, termsIn, TERMS_CODEC_NAME, VERSION_START, VERSION_CURRENT,
		state.SegmentInfo.GetID(), state.SegmentSuffix)
	if err != nil {
		return nil, err
	}
	this.version = version

	indexName := store.SegmentFileName(this.segment, state.SegmentSuffix, TERMS_INDEX_EXTENSION)
	indexIn, err := state.Directory.OpenInput(ctx, indexName)
	if err != nil {
		return nil, err
	}
	this.indexIn = indexIn

	_, err = codecs.CheckIndexHeader(ctx, indexIn, TERMS_INDEX_CODEC_NAME, version, version,
		state.SegmentInfo.GetID(), state.SegmentSuffix)
	if err != nil {
		return nil, err
	}

	if version < VERSION_META_FILE {
		// Have PostingsReader init itself
		if err := postingsReader.Init(ctx, termsIn, state); err != nil {
			return nil, err
		}

		// Verifying the checksum against all bytes would be too costly, but for now we at least
		// verify proper structure of the checksum footer. This is cheap and can detect some forms
		// of corruption such as file truncation.
		if _, err := codecs.RetrieveChecksum(ctx, indexIn); err != nil {
			return nil, err
		}
		if _, err := codecs.RetrieveChecksum(ctx, termsIn); err != nil {
			return nil, err
		}
	}

	// Read per-field details
	metaName := store.SegmentFileName(this.segment, state.SegmentSuffix, TERMS_META_EXTENSION)
	var fieldMap map[string]*FieldReader
	indexLength := -1
	termsLength := -1

	var metaIn store.ChecksumIndexInput

	if version >= VERSION_META_FILE {
		metaIn, err = store.OpenChecksumInput(ctx, state.Directory, metaName)
		if err != nil {
			return nil, err
		}
	}

	var indexMetaIn, termsMetaIn store.IndexInput
	if version >= VERSION_META_FILE {
		if _, err := codecs.CheckIndexHeader(ctx, metaIn, TERMS_META_CODEC_NAME, version, version,
			state.SegmentInfo.GetID(), state.SegmentSuffix); err != nil {
			return nil, err
		}
		indexMetaIn = metaIn
		termsMetaIn = metaIn
		if err := postingsReader.Init(ctx, metaIn, state); err != nil {
			return nil, err
		}
	} else {
		if err := seekDir(ctx, termsIn); err != nil {
			return nil, err
		}
		if err := seekDir(ctx, indexIn); err != nil {
			return nil, err
		}
		indexMetaIn = indexIn
		termsMetaIn = termsIn
	}

	numFields, err := termsMetaIn.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	if numFields < 0 {
		return nil, fmt.Errorf("invalid numFields: %d", numFields)
	}
	fieldMap = make(map[string]*FieldReader)

	for i := 0; i < int(numFields); i++ {
		field, err := termsMetaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		numTerms, err := termsMetaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		if numTerms <= 0 {
			return nil, fmt.Errorf("Illegal numTerms for field number: %d", field)
		}

		rootCode, err := readBytes(ctx, termsMetaIn)
		if err != nil {
			return nil, err
		}
		fieldInfo := state.FieldInfos.FieldInfoByNumber(int(field))
		if fieldInfo == nil {
			return nil, fmt.Errorf("invalid field number: %d", field)
		}
		sumTotalTermFreq, err := termsMetaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		// when frequencies are omitted, sumDocFreq=sumTotalTermFreq and only one value is written.
		sumDocFreq := 0
		if fieldInfo.GetIndexOptions() == document.INDEX_OPTIONS_DOCS {
			sumDocFreq = int(sumTotalTermFreq)
		} else {
			uvarint, err := termsMetaIn.ReadUvarint(ctx)
			if err != nil {
				return nil, err
			}
			sumDocFreq = int(uvarint)
		}
		docCountU64, err := termsMetaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		docCount := int(docCountU64)

		if version < VERSION_META_LONGS_REMOVED {
			longsSize, err := termsMetaIn.ReadUvarint(ctx)
			if err != nil {
				return nil, err
			}
			if longsSize < 0 {
				return nil, fmt.Errorf("invalid longsSize for field: %s, longsSize=%d", fieldInfo.Name(), longsSize)
			}
		}
		minTerm, err := readBytes(ctx, termsMetaIn)
		if err != nil {
			return nil, err
		}
		maxTerm, err := readBytes(ctx, termsMetaIn)
		if err != nil {
			return nil, err
		}

		maxDoc, err := state.SegmentInfo.MaxDoc()
		if err != nil {
			return nil, err
		}
		if docCount < 0 || docCount > maxDoc { // #docs with field must be <= #docs
			return nil, fmt.Errorf("invalid docCount: %d maxDoc: %d", docCount, maxDoc)
		}
		if sumDocFreq < docCount { // #postings must be >= #docs with field
			return nil, fmt.Errorf("invalid sumDocFreq: %d docCount: %d", sumDocFreq, docCount)
		}
		if int(sumTotalTermFreq) < sumDocFreq { // #positions must be >= #postings
			return nil, fmt.Errorf("invalid sumTotalTermFreq: %d sumDocFreq: %d", sumTotalTermFreq, sumDocFreq)
		}
		indexStartFP, err := indexMetaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}

		if _, ok := fieldMap[fieldInfo.Name()]; ok {
			return nil, fmt.Errorf("duplicate field: %s", fieldInfo.Name())
		}

		fieldReader, err := NewFieldReader(ctx, this, fieldInfo, int(numTerms), rootCode,
			int64(sumTotalTermFreq), int64(sumDocFreq), docCount,
			int64(indexStartFP), indexMetaIn, indexIn, minTerm, maxTerm)
		if err != nil {
			return nil, err
		}
		fieldMap[fieldInfo.Name()] = fieldReader

	}
	if version >= VERSION_META_FILE {
		n1, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		indexLength = int(n1)

		n2, err := metaIn.ReadUvarint(ctx)
		if err != nil {
			return nil, err
		}
		termsLength = int(n2)
	}

	if metaIn != nil {
		if _, err := codecs.CheckFooter(ctx, metaIn); err != nil {
			return nil, err
		}
	}

	if version >= VERSION_META_FILE {
		// At this point the checksum of the meta file has been verified so the lengths are likely correct
		if _, err := codecs.RetrieveChecksumWithLength(ctx, indexIn, indexLength); err != nil {
			return nil, err
		}
		if _, err := codecs.RetrieveChecksumWithLength(ctx, termsIn, termsLength); err != nil {
			return nil, err
		}
	}

	fieldList := lo.Keys(fieldMap)
	sort.Strings(fieldList)

	this.fieldMap = fieldMap
	this.fieldList = fieldList

	return this, nil
}

func (t *TermsReader) Close() error {
	err := util.Close(t.indexIn, t.termsIn, t.postingsReader)
	if err != nil {
		return err
	}
	clear(t.fieldMap)
	return nil
}

func (t *TermsReader) Iterator() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, v := range t.fieldList {
			if !yield(v) {
				return
			}
		}
	}
}

func (t *TermsReader) Names() []string {
	//TODO implement me
	panic("implement me")
}

func (t *TermsReader) Terms(field string) (index.Terms, error) {
	reader, ok := t.fieldMap[field]
	if !ok {
		return nil, errors.New("not found")
	}
	return reader, nil
}

func (t *TermsReader) Size() int {
	return len(t.fieldMap)
}

func (t *TermsReader) CheckIntegrity() error {
	if _, err := codecs.ChecksumEntireFile(context.Background(), t.indexIn); err != nil {
		return err
	}
	if _, err := codecs.ChecksumEntireFile(context.Background(), t.termsIn); err != nil {
		return err
	}
	return t.postingsReader.CheckIntegrity()
}

func (t *TermsReader) GetMergeInstance() index.FieldsProducer {
	return t
}

var (
	FST_OUTPUTS = fst.NewByteSequenceOutputs()
)

func seekDir(ctx context.Context, input store.IndexInput) error {
	if _, err := input.Seek(input.Length()-int64(codecs.FooterLength())-8, io.SeekStart); err != nil {
		return err
	}
	offset, err := input.ReadUvarint(ctx)
	if err != nil {
		return err
	}
	if _, err := input.Seek(int64(offset), io.SeekStart); err != nil {
		return err
	}
	return nil
}

func readBytes(ctx context.Context, in store.IndexInput) ([]byte, error) {
	numBytes, err := in.ReadUvarint(ctx)
	if err != nil {
		return nil, err
	}
	if numBytes < 0 {
		return nil, fmt.Errorf("invalid bytes length: %d", numBytes)
	}

	bytes := make([]byte, numBytes)
	if _, err := in.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil

}
