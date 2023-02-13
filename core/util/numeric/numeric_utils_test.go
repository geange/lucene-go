package numeric

import (
	"testing"
)

func TestIllegalSubtract(t *testing.T) {
	type args struct {
		bytesPerDim int
		dim         int
		a           []byte
		b           []byte
		result      []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "IllegalSubtract",
			args: args{
				bytesPerDim: 4,
				dim:         0,
				a:           []byte{0, 0, 0, 0xf0},
				b:           []byte{0, 0, 0, 0xf1},
				result:      []byte{0, 0, 0, 0},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Subtract(tt.args.bytesPerDim, tt.args.dim, tt.args.a, tt.args.b, tt.args.result); (err != nil) != tt.wantErr {
				t.Errorf("Subtract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	//rand.Seed(time.Now().UnixNano())
	//numBytes := rand.Intn(100) + 1
	//
	//for i := 0; i < 1000; i++ {
	//	big.Int
	//}
}
