package index

// ImpactsEnum Extension of PostingsEnum which also provides information about upcoming impacts.
type ImpactsEnum interface {
	PostingsEnum
	ImpactsSource
}
