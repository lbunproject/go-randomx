package randomx

import "testing"

func TestReciprocal(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		a uint32
		b uint64
	}{
		{3, 12297829382473034410},
		{13, 11351842506898185609},
		{33, 17887751829051686415},
		{65537, 18446462603027742720},
		{15000001, 10316166306300415204},
		{3845182035, 10302264209224146340},
		{0xffffffff, 9223372039002259456},
	}

	for i, tt := range tests {
		r := reciprocal(tt.a)
		if r != tt.b {
			t.Errorf("i=%d, a=%d", i, tt.a)
			t.Errorf("expected=%016x, actual=%016x", tt.b, r)
		}
	}
}
