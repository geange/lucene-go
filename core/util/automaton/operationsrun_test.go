package automaton

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRun(t *testing.T) {
	type args struct {
		a *Automaton
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Run(tt.args.a, tt.args.s), "Run(%v, %v)", tt.args.a, tt.args.s)
		})
	}
}
