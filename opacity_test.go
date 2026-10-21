package main

import "testing"

func TestTransparencyAlphaAndStyle(t *testing.T) {
	for _, tc := range []struct {
		percent int
		alpha   byte
	}{{0, 255}, {1, 252}, {30, 179}, {40, 153}, {70, 77}} {
		if got := transparencyAlpha(tc.percent); got != tc.alpha {
			t.Fatalf("%d: %d != %d", tc.percent, got, tc.alpha)
		}
	}
	for _, s := range []uintptr{0, 8, 0x200, 0x80008} {
		if got := layeredStyle(s); got != s|0x80000 {
			t.Fatalf("lost style: %x", got)
		}
	}
}
