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

func (r *CharArraySet) Add(key any) {
	r.RLock()
	defer r.RUnlock()

	switch key.(type) {
	case []byte:
		newKey := base64.StdEncoding.EncodeToString(key.([]byte))
		r.values[newKey] = struct{}{}
	case string:
		obj := key.(string)

		newKey := base64.StdEncoding.EncodeToString([]byte(obj))
		r.values[newKey] = struct{}{}
	default:
		return
	}
}

func (r *CharArraySet) Contain(key []byte) bool {
	r.RLock()
	defer r.RUnlock()

	newKey := base64.StdEncoding.EncodeToString(key)
	_, ok := r.values[newKey]
	return ok
}

func (r *CharArraySet) Clear() {
	r.Lock()
	defer r.Unlock()

	r.values = make(map[string]struct{})
}
