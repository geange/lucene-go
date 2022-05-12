package tokenattributes

// TermToBytesRefAttribute This attribute is requested by TermsHashPerField to index the contents.
// This attribute can be used to customize the final byte[] encoding of terms.
// Consumers of this attribute call getBytesRef() for each term. Example:
//
//     final TermToBytesRefAttribute termAtt = tokenStream.getAttribute(TermToBytesRefAttribute.class);
//
//     while (tokenStream.incrementToken() {
//       final BytesRef bytes = termAtt.getBytesRef();
//
//       if (isInteresting(bytes)) {
//
//         // because the bytes are reused by the attribute (like CharTermAttribute's char[] buffer),
//         // you should make a copy if you need persistent access to the bytes, otherwise they will
//         // be rewritten across calls to incrementToken()
//
//         doSomethingWith(BytesRef.deepCopyOf(bytes));
//       }
//     }
//     ...
type TermToBytesRefAttribute interface {
	GetBytesRef() []byte
}
