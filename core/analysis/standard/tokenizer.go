package standard

import (
	"bufio"
	"errors"
	"io"
	"unicode"

	"github.com/geange/lucene-go/core/analysis"
)

type Tokenizer struct {
	*analysis.BaseTokenizer

	scanner          *stdTokenizer
	skippedPositions int
	maxTokenLength   int
}

func NewTokenizer() *Tokenizer {
	tokenizer := &Tokenizer{
		BaseTokenizer:    analysis.NewBaseTokenizer(),
		scanner:          newStdTokenizer(),
		skippedPositions: 0,
		maxTokenLength:   0,
	}
	return tokenizer
}

func (r *Tokenizer) IncrementToken() (bool, error) {
	if err := r.AttributeSource().Reset(); err != nil {
		return false, err
	}
	r.skippedPositions = 0

	text, err := r.scanner.GetNextToken()
	if err != nil {
		if errors.Is(err, io.EOF) {
			r.AttributeSource().Type().SetType("ALPHANUM")
			if err := r.AttributeSource().CharTerm().AppendString(text); err != nil {
				return false, err
			}
			if err := r.AttributeSource().Offset().
				SetOffset(r.scanner.slow, r.scanner.slow+len(text)); err != nil {
				return false, err
			}
			return false, nil
		}
		return false, err
	}

	r.AttributeSource().Type().SetType("ALPHANUM")
	if err := r.AttributeSource().CharTerm().AppendString(text); err != nil {
		return false, err
	}
	if err := r.AttributeSource().Offset().
		SetOffset(r.scanner.slow, r.scanner.slow+len(text)); err != nil {
		return false, err
	}
	return true, nil
}

func (r *Tokenizer) SetReader(reader io.Reader) error {
	r.scanner.SetReader(reader)
	return nil
}

func (r *Tokenizer) setMaxTokenLength(length int) {
	r.maxTokenLength = length
}

type TokenType int

const (
	ALPHANUM = TokenType(iota)
	NUM
	SOUTHEAST_ASIAN
	IDEOGRAPHIC
	HIRAGANA
	KATAKANA
	HANGUL
	EMOJI
)

// stdTokenizer
// This class implements Word Break rules from the Unicode Text Segmentation algorithm, as
// specified in Unicode Standard Annex #29 .
// Tokens produced are of the following types:
// * <ALPHANUM>: A sequence of alphabetic and numeric characters
// * <NUM>: A number
// * <SOUTHEAST_ASIAN>: A sequence of characters from South and Southeast Asian languages, including Thai, Lao, Myanmar, and Khmer
// * <IDEOGRAPHIC>: A single CJKV ideographic character
// * <HIRAGANA>: A single hiragana character
// * <KATAKANA>: A sequence of katakana characters
// * <HANGUL>: A sequence of Hangul characters
// * <EMOJI>: A sequence of Emoji characters
type stdTokenizer struct {
	reader io.RuneReader
	slow   int
	fast   int
	buff   []rune
	eof    bool
}

func newStdTokenizer() *stdTokenizer {
	return &stdTokenizer{
		reader: nil,
		slow:   0,
		fast:   0,
		buff:   make([]rune, 0),
	}
}

func (r *stdTokenizer) SetReader(reader io.Reader) {
	r.reader = bufio.NewReader(reader)
	r.slow, r.fast = 0, 0
	r.buff = r.buff[:0]
	r.eof = false
}

func (r *stdTokenizer) GetNextToken() (string, error) {
	if r.eof {
		return "", io.EOF
	}

	if r.slow < r.fast {
		r.slow = r.fast
		r.buff = r.buff[:0]
	}

	for {
		char, n, err := r.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				r.eof = true
				return string(r.buff), nil
			}
			return "", err
		}

		if !unicode.IsSpace(char) {
			r.fast += n
			r.buff = append(r.buff, char)
		} else {
			r.fast += n
			r.fast++
			break
		}
	}

	return string(r.buff), nil
}
