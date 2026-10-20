//go:build windows

package main

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNativePingLoopback(t *testing.T) {
	p, err := newPinger()
	if err != nil {
		t.Fatal(err)
	}
	defer p.close()
	if !p.ping("127.0.0.1", time.Second) {
		t.Fatal("loopback ICMP failed")
	}
	if p.ping("bad-address", time.Second) {
		t.Fatal("invalid IP accepted")
	}
}
func TestDNSUsesSpecifiedServer(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	go func() {
		b := make([]byte, 1500)
		for {
			n, addr, e := pc.ReadFrom(b)
			if e != nil {
				return
			}
			if n < 12 {
				continue
			}
			end := 12
			for end < n && b[end] != 0 {
				end += int(b[end]) + 1
			}
			end += 5
			if end > n {
				continue
			}
			reply := append([]byte(nil), b[:end]...)
			reply[2] = 0x81
			reply[3] = 0x80
			reply[6] = 0
			reply[7] = 1
			reply[10] = 0
			reply[11] = 0
			reply = append(reply, 0xc0, 0x0c, 0, 1, 0, 1, 0, 0, 0, 30, 0, 4, 127, 0, 0, 1)
			pc.WriteTo(reply, addr)
		}
	}()
	if !checkDNS(context.Background(), pc.LocalAddr().String(), time.Second) {
		t.Fatal("local DNS response not accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if checkDNS(ctx, pc.LocalAddr().String(), time.Second) {
		t.Fatal("cancel ignored")
	}
}
