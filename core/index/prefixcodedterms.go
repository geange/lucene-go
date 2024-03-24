package index

// PrefixCodedTerms
// Prefix codes term instances (prefixes are shared).
// This is expected to be faster to build than a FST and might also be more compact
// if there are no common suffixes.
// lucene.internal
type PrefixCodedTerms struct {
	//buffer *store.RAMFile
	size   int64
	delGen int64
}

/*

func NewPrefixCodedTerms(buffer *store.RAMFile, size int64) *PrefixCodedTerms {
	return &PrefixCodedTerms{buffer: buffer, size: size}
}

func (p *PrefixCodedTerms) Iterator() (*TermIterator, error) {
	return NewTermIterator(p.delGen, p.buffer)
}

// PrefixCodedTermsBuilder
// Builds a PrefixCodedTerms: call add repeatedly, then finish.
type PrefixCodedTermsBuilder struct {
	buffer        *store.RAMFile
	output        *store.RAMOutputStream
	lastTerm      *Term
	lastTermBytes *bytes.Buffer
	size          int64
}

func NewPrefixCodedTermsBuilder() *PrefixCodedTermsBuilder {
	buffer := store.NewRAMFile()
	output := store.NewRAMOutputStreamV1("", buffer, false)

	return &PrefixCodedTermsBuilder{
		buffer:        buffer,
		output:        output,
		lastTerm:      NewTerm("", nil),
		lastTermBytes: new(bytes.Buffer),
	}
}

func (p *PrefixCodedTermsBuilder) Add(ctx context.Context, term *Term) error {
	return p.AddBytes(ctx, term.field, term.NewBytes())
}

func (p *PrefixCodedTermsBuilder) AddBytes(ctx context.Context, field string, bs []byte) (err error) {
	var prefix int
	if p.size > 0 && field == p.lastTerm.field {
		// same field as the last term
		prefix, err = util.BytesDifference(p.lastTerm.bytes, bs)
		if err != nil {
			return err
		}
		err = p.output.WriteUvarint(ctx, uint64(prefix<<1))
		if err != nil {
			return err
		}
	} else {
		// field change
		prefix = 0
		err = p.output.WriteUvarint(ctx, 1)
		if err != nil {
			return err
		}
		err = p.output.WriteString(ctx, field)
		if err != nil {
			return err
		}
	}

	suffix := len(bs) - prefix
	err = p.output.WriteUvarint(ctx, uint64(suffix))
	if err != nil {
		return err
	}
	_, err = p.output.Write(bs[prefix:])
	if err != nil {
		return err
	}
	p.lastTermBytes.Reset()
	p.lastTermBytes.Write(bs)
	p.lastTerm.bytes = p.lastTermBytes.NewBytes()
	p.lastTerm.field = field
	p.size++

	return nil
}

func (p *PrefixCodedTermsBuilder) Finish() *PrefixCodedTerms {
	//err := p.output.Close()
	//if err != nil {
	//	return nil
	//}
	//return NewPrefixCodedTerms(p.buffer, p.size)
	// TODO: fix
}

var _ FieldTermIterator = &TermIterator{}

// TermIterator
// An iterator over the list of terms stored in a PrefixCodedTerms.
type TermIterator struct {
	input  store.IndexInput
	bytes  []byte
	end    int64
	delGen int64
	field  string
}

func NewTermIterator(delGen int64, buffer *store.RAMFile) (*TermIterator, error) {
	input, err := store.NewRAMIndexInput("PrefixCodedTermsIterator", buffer)
	if err != nil {
		return nil, err
	}
	return &TermIterator{
		input:  input,
		end:    input.Length(),
		delGen: delGen,
	}, nil
}

func (t *TermIterator) Next(context.Context) ([]byte, error) {
	if t.input.GetFilePointer() >= t.end {
		t.field = ""
		return nil, nil
	}

	code, err := t.input.ReadUvarint(context.Background())
	if err != nil {
		return nil, err
	}
	newField := (code & 1) != 0
	if newField {
		t.field, err = t.input.ReadString(context.Background())
		if err != nil {
			return nil, err
		}
	}

	prefix := code >> 1
	suffix, err := t.input.ReadUvarint(context.Background())
	if err != nil {
		return nil, err
	}
	return t.readTermBytes(int(prefix), int(suffix))
}

func (t *TermIterator) readTermBytes(prefix, suffix int) ([]byte, error) {
	t.bytes = array.Grow(t.bytes, prefix+suffix)
	_, err := t.input.Read(t.bytes[:suffix])
	if err != nil {
		return nil, err
	}
	return t.bytes[:suffix], nil
}

func (t *TermIterator) Field() string {
	return t.field
}

func (t *TermIterator) DelGen() int64 {
	return t.delGen
}


*/
