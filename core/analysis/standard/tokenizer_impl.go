package standard

// TokenizerImpl This class implements Word Break rules from the Unicode Text Segmentation algorithm, as specified in Unicode Standard Annex #29 .
//Tokens produced are of the following types:
// * <ALPHANUM>: A sequence of alphabetic and numeric characters
// * <NUM>: A number
// * <SOUTHEAST_ASIAN>: A sequence of characters from South and Southeast Asian languages, including Thai, Lao, Myanmar, and Khmer
// * <IDEOGRAPHIC>: A single CJKV ideographic character
// * <HIRAGANA>: A single hiragana character
// * <KATAKANA>: A sequence of katakana characters
// * <HANGUL>: A sequence of Hangul characters
// * <EMOJI>: A sequence of Emoji characters
type TokenizerImpl struct {

	// initial size of the lookahead buffer
	ZZ_BUFFERSIZE int
}

const (
	// YYEOF This character denotes the end of file
	YYEOF = -1

	// YYINITIAL lexical states
	YYINITIAL = 0
)

var (
	// ZZ_LEXSTATE ZZ_LEXSTATE[l] is the state in the DFA for the lexical state l ZZ_LEXSTATE[l+1] is the state
	// in the DFA for the lexical state l at the beginning of a line l is of the form l = 2*k, k a non negative
	// integer
	ZZ_LEXSTATE = []int{0, 0}
)
