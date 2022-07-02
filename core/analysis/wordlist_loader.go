package analysis

import (
	"bufio"
	"io"
	"strings"
)

const (
	INITIAL_CAPACITY = 16
)

// WordlistLoader Loader for text files that represent a list of stopwords.
// See Also: to obtain {@link Reader} instances
type WordlistLoader struct {
}

// GetWordSet Reads lines from a Reader and adds every line as an entry to a CharArraySet (omitting leading and trailing whitespace). Every line of the Reader should contain only one word. The words need to be in lowercase if you make use of an Analyzer which uses LowerCaseFilter (like StandardAnalyzer).
// Params: 	reader – Reader containing the wordlist
//			set – the CharArraySet to fill with the readers words
// Returns: the given CharArraySet with the reader's words
func (r *WordlistLoader) GetWordSet(reader io.Reader, set *CharArraySet) error {
	buff := bufio.NewReader(reader)

	for {
		line, _, err := buff.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		set.Add(strings.TrimSpace(string(line)))
	}
}
