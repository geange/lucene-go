package automaton

import (
	"reflect"
	"testing"
)

func TestNewFrozenIntSet(t *testing.T) {
	tests := []struct {
		name       string
		values     []int
		state      int
		hashCode   int64
		wantValues []int
		wantState  int
		wantCode   int64
	}{
		{
			name:       "Normal case",
			values:     []int{1, 2, 3},
			state:      0,
			hashCode:   123456789,
			wantValues: []int{1, 2, 3},
			wantState:  0,
			wantCode:   123456789,
		},
		{
			name:       "Nil slice",
			values:     nil,
			state:      -1,
			hashCode:   0,
			wantValues: nil,
			wantState:  -1,
			wantCode:   0,
		},
		{
			name:       "Empty slice",
			values:     []int{},
			state:      1,
			hashCode:   -987654321,
			wantValues: []int{},
			wantState:  1,
			wantCode:   -987654321,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFrozenIntSet(tt.values, uint64(tt.hashCode), tt.state)
			if !reflect.DeepEqual(got.GetArray(), tt.wantValues) {
				t.Errorf("Values mismatch: got %v, want %v", got.GetArray(), tt.wantValues)
			}
			if !reflect.DeepEqual(got.Size(), len(tt.wantValues)) {
				t.Errorf("Values size mismatch: got %v, want %v", got.Size(), len(tt.wantValues))
			}

			if got.state != tt.wantState {
				t.Errorf("State mismatch: got %d, want %d", got.state, tt.wantState)
			}
			if got.Hash() != uint64(tt.wantCode) {
				t.Errorf("HashCode mismatch: got %d, want %d", got.Hash(), tt.wantCode)
			}
		})
	}
}

func TestFrozenIntSet_Equals(t *testing.T) {
	tests := []struct {
		name     string
		f        *FrozenIntSet
		other    Hashable
		expected bool
	}{
		{
			name:     "TC01 - both nil",
			f:        nil,
			other:    (*FrozenIntSet)(nil),
			expected: true,
		},
		{
			name:     "TC02 - f not nil, other nil",
			f:        &FrozenIntSet{},
			other:    nil,
			expected: false,
		},
		{
			name: "TC03 - different type",
			f: &FrozenIntSet{
				values:   []int{1, 2, 3},
				state:    1,
				hashCode: 123,
			},
			other:    &MockIntSet{},
			expected: false,
		},
		{
			name: "TC04 - values differ",
			f: &FrozenIntSet{
				values:   []int{1, 2, 3},
				state:    1,
				hashCode: 123,
			},
			other: &FrozenIntSet{
				values:   []int{1, 2},
				state:    1,
				hashCode: 123,
			},
			expected: false,
		},
		{
			name: "TC05 - state differs",
			f: &FrozenIntSet{
				values:   []int{1, 2, 3},
				state:    1,
				hashCode: 123,
			},
			other: &FrozenIntSet{
				values:   []int{1, 2, 3},
				state:    2,
				hashCode: 123,
			},
			expected: false,
		},
		{
			name: "TC06 - hashCode differs",
			f: &FrozenIntSet{
				values:   []int{1, 2, 3},
				state:    1,
				hashCode: 123,
			},
			other: &FrozenIntSet{
				values:   []int{1, 2, 3},
				state:    1,
				hashCode: 456,
			},
			expected: false,
		},
		{
			name: "TC07 - all fields equal",
			f: &FrozenIntSet{
				values:   []int{1, 2, 3},
				state:    1,
				hashCode: 123,
			},
			other: &FrozenIntSet{
				values:   []int{1, 2, 3},
				state:    1,
				hashCode: 123,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.f.Equals(tt.other)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
