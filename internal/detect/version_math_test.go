package detect

import "testing"

func TestAllocator(t *testing.T) {
	cases := []struct{ last, orig, want string }{
		{"5", "0002", "0006"},
		{"20240101235959", "20240101000000", "20240102000000"},
		{"99", "1", "100"},
	}
	for _, tc := range cases {
		a := &allocator{last: tc.last}
		if got := a.next(tc.orig); got != tc.want {
			t.Errorf("next(last=%s, orig=%s) = %s, want %s", tc.last, tc.orig, got, tc.want)
		}
	}
	if cmpDigits(trimZeros("0010"), "9") <= 0 || trimZeros("000") != "0" {
		t.Error("digit comparison broken")
	}
}
