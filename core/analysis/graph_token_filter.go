package analysis

// GraphTokenFilter An abstract TokenFilter that exposes its input stream as a graph Call incrementBaseToken()
// to move the root of the graph to the next position in the TokenStream, incrementGraphToken() to move along
// the current graph, and incrementGraph() to reset to the next graph based at the current root. For example,
// given the stream 'a b/c:2 d e`, then with the base token at 'a', incrementGraphToken() will produce the
// stream 'a b d e', and then after calling incrementGraph() will produce the stream 'a c e'.
type GraphTokenFilter struct {
}
