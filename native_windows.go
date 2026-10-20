//go:build windows

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

var iphlp = syscall.NewLazyDLL("iphlpapi.dll")
var icmpCreate = iphlp.NewProc("IcmpCreateFile")
var icmpSend = iphlp.NewProc("IcmpSendEcho")
var icmpClose = iphlp.NewProc("IcmpCloseHandle")
var kernel = syscall.NewLazyDLL("kernel32.dll")
var setColor = kernel.NewProc("SetConsoleTextAttribute")
var outputHandle = func() uintptr { h, _, _ := kernel.NewProc("GetStdHandle").Call(uintptr(0xfffffff5)); return h }()

type pinger struct{ handle uintptr }

func newPinger() (*pinger, error) {
	h, _, err := icmpCreate.Call()
	if h == ^uintptr(0) || h == 0 {
		return nil, fmt.Errorf("IcmpCreateFile: %v", err)
	}
	return &pinger{h}, nil
}
func (p *pinger) close() { icmpClose.Call(p.handle) }
func (p *pinger) ping(ip string, timeout time.Duration) bool {
	addr := net.ParseIP(ip).To4()
	if addr == nil {
		return false
	}
	payload := []byte("vpn-monitor")
	reply := make([]byte, 512)
	count, _, _ := icmpSend.Call(p.handle, uintptr(binary.LittleEndian.Uint32(addr)), uintptr(unsafe.Pointer(&payload[0])), uintptr(len(payload)), 0, uintptr(unsafe.Pointer(&reply[0])), uintptr(len(reply)), uintptr(timeout.Milliseconds()))
	runtime.KeepAlive(payload)
	runtime.KeepAlive(reply)
	return count > 0 && binary.LittleEndian.Uint32(reply[4:8]) == 0
}
func checkDNS(parent context.Context, server string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	resolver := &net.Resolver{PreferGo: true, StrictErrors: true, Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{Timeout: timeout}).DialContext(ctx, network, server)
	}}
	ips, err := resolver.LookupIP(ctx, "ip4", "youtube.com")
	return err == nil && len(ips) > 0
}
func consoleSetup() { kernel.NewProc("SetConsoleOutputCP").Call(65001) }
func printColor(line string, color uintptr) {
	setColor.Call(outputHandle, color)
	fmt.Println(line)
	setColor.Call(outputHandle, 7)
}
