package main

import "testing"

func TestPinRequestPreservesPositionSizeActivation(t *testing.T) {
	on, flags := pinRequest(true)
	off, flagsOff := pinRequest(false)
	if on != ^uintptr(0) || off != ^uintptr(1) {
		t.Fatal("must use HWND_TOPMOST and HWND_NOTOPMOST")
	}
	if flags != 0x13 || flagsOff != flags {
		t.Fatal("pin must use NOMOVE NOSIZE NOACTIVATE and allow Z order change")
	}
}
