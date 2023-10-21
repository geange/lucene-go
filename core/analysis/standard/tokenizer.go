package standard

import (
	"bufio"
	"github.com/geange/lucene-go/core/analysis"
	"io"
	"unicode"
)

type Tokenizer struct {
	*analysis.TokenizerBase

	scanner *stdTokenizer

	skippedPositions int
	maxTokenLength   int
}

func NewTokenizer() *Tokenizer {
	tokenizer := &Tokenizer{
		TokenizerBase:    analysis.NewTokenizer(),
		scanner:          newStdTokenizer(),
		skippedPositions: 0,
		maxTokenLength:   0,
	}
	return tokenizer
}

func (r *Tokenizer) IncrementToken() (bool, error) {
	r.AttributeSource().Clear()
	r.skippedPositions = 0

	text, err := r.scanner.GetNextToken()
	if err != nil {
		if err == io.EOF {
			r.AttributeSource().Type().SetType("ALPHANUM")
			r.AttributeSource().CharTerm().Append(text)
			r.AttributeSource().Offset().SetOffset(r.scanner.Slow, r.scanner.Slow+len(text))
			return false, nil
		}
		return false, err
	}

	r.AttributeSource().Type().SetType("ALPHANUM")
	r.AttributeSource().CharTerm().Append(text)
	r.AttributeSource().Offset().SetOffset(r.scanner.Slow, r.scanner.Slow+len(text))
	return true, nil
}

func (r *Tokenizer) SetReader(reader io.Reader) error {
	r.scanner.SetReader(reader)
	return nil
}

func (r *Tokenizer) setMaxTokenLength(length int) {
	r.maxTokenLength = length
}

// stdTokenizer This class implements Word Break rules from the Unicode Text Segmentation algorithm, as
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
	Slow   int
	Fast   int

	buff []rune

	EOF bool
}

func newStdTokenizer() *stdTokenizer {
	return &stdTokenizer{
		reader: nil,
		Slow:   0,
		Fast:   0,
		buff:   make([]rune, 0),
	}
}

func (r *stdTokenizer) SetReader(reader io.Reader) {
	r.reader = bufio.NewReader(reader)
	r.Slow, r.Fast = 0, 0
	r.buff = r.buff[:0]
	r.EOF = false
}

func (r *stdTokenizer) GetNextToken() (string, error) {
	if r.EOF {
		return "", io.EOF
	}

	if r.Slow < r.Fast {
		r.Slow = r.Fast
		r.buff = r.buff[:0]
	}

	for {
		char, n, err := r.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				r.EOF = true
				return string(r.buff), nil
			}
			return "", err
		}

		if !unicode.IsSpace(char) {
			r.Fast += n
			r.buff = append(r.buff, char)
		} else {
			r.Fast += n
			r.Fast++
			break
		}
	}

	return string(r.buff), nil
}
