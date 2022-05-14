package tokenattributes

// PayloadAttribute The payload of a Token.
// The payload is stored in the index at each position, and can be used to influence scoring when using
// Payload-based queries.
// NOTE: because the payload will be stored at each position, it's usually best to use the minimum number of
// bytes necessary. Some codec implementations may optimize payload storage when all payloads have the same length.
// See Also: org.apache.lucene.index.PostingsEnum
type PayloadAttribute interface {
	// GetPayload Returns this Token's payload.
	// See Also: setPayload(BytesRef)
	GetPayload() []byte

	// SetPayload Sets this Token's payload.
	// See Also: getPayload()
	SetPayload(payload []byte) error
}
