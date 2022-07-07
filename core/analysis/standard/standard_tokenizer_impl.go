package standard

import (
	"bufio"
	"io"
	"unicode"
)

// StandardTokenizerImpl This class implements Word Break rules from the Unicode Text Segmentation algorithm, as
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
type StandardTokenizerImpl struct {
	reader io.RuneReader
	Slow   int
	Fast   int

	buff []rune
}

func NewStandardTokenizerImpl() *StandardTokenizerImpl {
	return &StandardTokenizerImpl{
		reader: nil,
		Slow:   0,
		Fast:   0,
		buff:   make([]rune, 0),
	}
}

func (r *StandardTokenizerImpl) SetReader(reader io.Reader) {
	r.reader = bufio.NewReader(reader)
	r.Slow, r.Fast = 0, 0
	r.buff = r.buff[:0]
}

func (r *StandardTokenizerImpl) GetNextToken() (string, error) {
	if r.Slow < r.Fast {
		r.Slow = r.Fast
		r.buff = r.buff[:0]
	}

	for {
		char, n, err := r.reader.ReadRune()
		if err != nil {
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
