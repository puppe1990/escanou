package handlers

import "testing"

func TestNormalizeBarcode(t *testing.T) {
	if got := normalizeBarcode("789 6256-8010 11"); got != "7896256801011" {
		t.Fatalf("got %q", got)
	}
}

func TestValidBarcode_checksum(t *testing.T) {
	cases := []struct {
		code string
		ok   bool
	}{
		{"7894900011517", true},
		{"5901234123457", true},
		{"1234567890123", false},
		{"abc", false},
		{"123", false},
	}
	for _, tc := range cases {
		if got := validBarcode(tc.code); got != tc.ok {
			t.Errorf("validBarcode(%q) = %v, want %v", tc.code, got, tc.ok)
		}
	}
}
