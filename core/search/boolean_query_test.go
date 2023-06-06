package search

import "testing"

func TestNewBooleanQueryBuilder(t *testing.T) {
	builder := NewBooleanQueryBuilder()
	builder.AddQuery()
}
