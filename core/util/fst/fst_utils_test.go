package fst

import "testing"

func Test_getNumPresenceBytes(t *testing.T) {
	type args struct {
		labelRange int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			args: args{labelRange: 10},
			want: 2,
		},
		{
			args: args{labelRange: 7},
			want: 1,
		},
		{
			args: args{labelRange: 16},
			want: 2,
		},
		{
			args: args{labelRange: 17},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNumPresenceBytes(tt.args.labelRange); got != tt.want {
				t.Errorf("getNumPresenceBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
