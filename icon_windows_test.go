//go:build windows

package main

import (
	"encoding/binary"
	"os"
	"syscall"
	"testing"
)

func TestEmbeddedApplicationIcon(t *testing.T) {
	b, err := os.ReadFile("app.ico")
	if err != nil {
		t.Fatal("custom icon missing:", err)
	}
	if len(b) < 6 || binary.LittleEndian.Uint16(b[2:]) != 1 {
		t.Fatal("not ICO")
	}
	count := int(binary.LittleEndian.Uint16(b[4:]))
	sizes := map[int]bool{}
	for i := 0; i < count; i++ {
		n := int(b[6+i*16])
		if n == 0 {
			n = 256
		}
		sizes[n] = true
	}
	for _, n := range []int{16, 32, 48, 256} {
		if !sizes[n] {
			t.Fatalf("missing %dpx frame", n)
		}
	}
	instance, _, _ := kernel.NewProc("GetModuleHandleW").Call(0)
	for _, n := range []uintptr{16, 32, 48, 256} {
		h, _, e := user32.NewProc("LoadImageW").Call(instance, 1, 1, n, n, 0)
		if h == 0 {
			t.Fatalf("embedded icon ID 1 at %dpx: %v", n, e)
		}
		user32.NewProc("DestroyIcon").Call(h)
	}
	defaultIcon, _, _ := user32.NewProc("LoadIconW").Call(0, 32512)
	if defaultIcon == 0 {
		t.Fatal(syscall.GetLastError())
	}
}
