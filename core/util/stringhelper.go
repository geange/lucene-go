package util

import (
	"errors"
	"github.com/geange/lucene-go/core/util/array"
	"golang.org/x/exp/rand"
	"math/big"
	"sync"
)

func BytesDifference(priorTerm, currentTerm []byte) (int, error) {
	mismatch := array.Mismatch(priorTerm, currentTerm)
	if mismatch < 0 {
		return -1, errors.New("terms out of order")
	}
	return mismatch, nil
}

var (
	nextId = big.NewInt(rand.Int63())
	idLock sync.Mutex
	one    = big.NewInt(1)
)

const (
	ID_LENGTH = 16
)

// RandomId Generates a non-cryptographic globally unique id.
func RandomId() []byte {
	// NOTE: we don't use Java's UUID.randomUUID() implementation here because:
	//
	//   * It's overkill for our usage: it tries to be cryptographically
	//     secure, whereas for this use we don't care if someone can
	//     guess the IDs.
	//
	//   * It uses SecureRandom, which on Linux can easily take a long time
	//     (I saw ~ 10 seconds just running a Lucene test) when entropy
	//     harvesting is falling behind.
	//
	//   * It loses a few (6) bits to version and variant and it's not clear
	//     what impact that has on the period, whereas the simple ++ (mod 2^128)
	//     we use here is guaranteed to have the full period.

	idLock.Lock()
	defer idLock.Unlock()

	bits := nextId.Bytes()

	nextId.Add(nextId, one)

	result := make([]byte, ID_LENGTH)
	if len(bits) > ID_LENGTH {
		copy(result, bits[len(bits)-ID_LENGTH:])
	} else {
		copy(result[len(result)-len(bits):], bits)
	}
	return result
}

func StringRandomId(bs []byte) string {
	num := big.NewInt(0)
	return num.SetBytes(bs).String()
}
