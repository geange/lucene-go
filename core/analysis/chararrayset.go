package analysis

import (
	"encoding/base64"
	"sync"
)

// CharArraySet A simple class that stores Strings as char[]'s in a hash table. Note that this is not a general
// purpose class. For example, it cannot remove items from the set, nor does it resize its hash table to be
// smaller, etc. It is designed to be quick to test if a char[] is in the set without the necessity of converting
// it to a String first.
// Please note: This class implements Set but does not behave like it should in all cases. The generic type
// is Set , because you can add any object to it, that has a string representation. The add methods will use
// Object.toString and store the result using a char[] buffer. The same behavior have the contains() methods.
// The iterator() returns an Iterator .
type CharArraySet struct {
	sync.RWMutex

	values map[string]struct{}
}

func NewCharArraySet() *CharArraySet {
	return &CharArraySet{
		RWMutex: sync.RWMutex{},
		values:  make(map[string]struct{}),
	}
}

var (
	EmptySet = &CharArraySet{
		RWMutex: sync.RWMutex{},
		values: map[string]struct{}{
			" ":  {},
			"\n": {},
			"\t": {},
		},
	}
)

func (r *CharArraySet) Add(key any) {
	r.RLock()
	defer r.RUnlock()

	switch v := key.(type) {
	case []byte:
		newKey := base64.StdEncoding.EncodeToString(v)
		r.values[newKey] = struct{}{}
	case string:
		newKey := base64.StdEncoding.EncodeToString([]byte(v))
		r.values[newKey] = struct{}{}
	default:
		return
	}
}

func (r *CharArraySet) Contain(key []byte) bool {
	r.RLock()
	defer r.RUnlock()

	b64Key := base64.StdEncoding.EncodeToString(key)
	if _, ok := r.values[b64Key]; ok {
		return true
	}
	return false
}

func (r *CharArraySet) Clear() {
	r.Lock()
	defer r.Unlock()

	clear(r.values)
}
