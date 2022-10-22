package fst

import "testing"

func Test_bitTable_assertIsValid(t *testing.T) {
	type args struct {
		arc *FSTArc
		in  BytesReader
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bitTable{}
			got, err := b.assertIsValid(tt.args.arc, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("assertIsValid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("assertIsValid() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitTable_countBits(t *testing.T) {
	type args struct {
		arc *FSTArc
		in  BytesReader
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bitTable{}
			got, err := b.countBits(tt.args.arc, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("countBits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("countBits() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitTable_countBitsUpTo(t *testing.T) {
	type args struct {
		bitIndex int
		arc      *FSTArc
		in       BytesReader
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bitTable{}
			got, err := b.countBitsUpTo(tt.args.bitIndex, tt.args.arc, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("countBitsUpTo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("countBitsUpTo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitTable_isBitSet(t *testing.T) {
	type args struct {
		bitIndex int
		arc      *FSTArc
		in       BytesReader
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bitTable{}
			got, err := b.isBitSet(tt.args.bitIndex, tt.args.arc, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("isBitSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isBitSet() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitTable_nextBitSet(t *testing.T) {
	type args struct {
		bitIndex int
		arc      *FSTArc
		in       BytesReader
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bitTable{}
			got, err := b.nextBitSet(tt.args.bitIndex, tt.args.arc, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("nextBitSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("nextBitSet() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_bitTable_previousBitSet(t *testing.T) {
	type args struct {
		bitIndex int
		arc      *FSTArc
		in       BytesReader
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bitTable{}
			got, err := b.previousBitSet(tt.args.bitIndex, tt.args.arc, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("previousBitSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("previousBitSet() got = %v, want %v", got, tt.want)
			}
		})
	}
}
