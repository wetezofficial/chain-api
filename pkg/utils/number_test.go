package utils

import "testing"

func Test_ToUint64(t *testing.T) {
	tests := []struct {
		input string
		want  uint64
	}{
		{"1", 1},
		{"0x1", 1},
		{"3.28594425e+08", 328594425},
	}
	for _, test := range tests {
		got, err := ToUint64(test.input)
		if err != nil {
			t.Errorf("ToUint64(%s) = %v", test.input, err)
		}
		if got != test.want {
			t.Errorf("ToUint64(%s) = %d, want %d", test.input, got, test.want)
		}
	}
}
