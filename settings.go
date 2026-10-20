package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type targetConfig struct {
	Name string `json:"name"`
	IP   string `json:"ipv4"`
}
type settings struct {
	Version   int             `json:"version"`
	Targets   [4]targetConfig `json:"targets"`
	DNSServer string          `json:"dns_server"`
	DNSDomain string          `json:"dns_domain"`
	TimeoutMS int             `json:"timeout_ms"`
}

func defaultSettings() settings {
	return settings{1, [4]targetConfig{{"Debian", "192.168.1.107"}, {"Router", "192.168.1.100"}, {"Yandex", "77.88.44.242"}, {"YouTube", "192.178.24.110"}}, "192.168.1.107:53", "youtube.com", 2000}
}
func (c settings) validate() error {
	if c.Version != 1 {
		return fmt.Errorf("Неподдерживаемая версия настроек: %d", c.Version)
	}
	for i, t := range c.Targets {
		if strings.TrimSpace(t.Name) == "" || utf8.RuneCountInString(t.Name) > 32 || strings.IndexFunc(t.Name, unicode.IsControl) >= 0 {
			return fmt.Errorf("Узел %d: введите имя (1–32 символа, без переводов строк)", i+1)
		}
		if ip := net.ParseIP(t.IP); ip == nil || ip.To4() == nil || strings.Contains(t.IP, ":") {
			return fmt.Errorf("Узел %d: введите IPv4, например 127.0.0.1", i+1)
		}
	}
	host, port, e := net.SplitHostPort(c.DNSServer)
	n, pe := strconv.Atoi(port)
	if ip := net.ParseIP(host); e != nil || ip == nil || ip.To4() == nil || strings.Contains(host, ":") || pe != nil || n < 1 || n > 65535 {
		return fmt.Errorf("DNS: введите IPv4:порт, например 192.168.1.107:53 (порт 1–65535)")
	}
	d := strings.TrimSuffix(c.DNSDomain, ".")
	if len(d) == 0 || len(d) > 253 || net.ParseIP(d) != nil {
		return fmt.Errorf("DNS: введите доменное имя, например youtube.com (не IP)")
	}
	for _, s := range strings.Split(d, ".") {
		if len(s) == 0 || len(s) > 63 || s[0] == '-' || s[len(s)-1] == '-' {
			return fmt.Errorf("DNS: неверное доменное имя; используйте ASCII / punycode")
		}
		for _, r := range s {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
				return fmt.Errorf("DNS: используйте ASCII / punycode, без пробелов и URL")
			}
		}
	}
	if c.TimeoutMS < 1 || c.TimeoutMS > 60000 {
		return fmt.Errorf("Введите целое число от 1 до 60000 мс")
	}
	return nil
}
func loadSettings(path string) (settings, error) {
	d := defaultSettings()
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		return d, nil
	}
	if e != nil {
		return d, e
	}
	var c settings
	if e = json.Unmarshal(b, &c); e != nil {
		return d, e
	}
	if e = c.validate(); e != nil {
		return d, e
	}
	return c, nil
}
func saveSettings(path string, c settings) error {
	if e := c.validate(); e != nil {
		return e
	}
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".vpn-monitor-*.tmp")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if e = f.Chmod(0600); e == nil {
		_, e = f.Write(append(b, '\n'))
	}
	if e == nil {
		e = f.Sync()
	}
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		return e
	}
	return os.Rename(tmp, path)
}
func (c settings) node(i int) targetConfig {
	if i < 4 {
		return c.Targets[i]
	}
	return targetConfig{"DNS", c.DNSServer}
}
func (c settings) identity(i int) string {
	n := c.node(i)
	s := n.Name + " (" + n.IP + ")"
	if i == 4 {
		s += " · " + c.DNSDomain
	}
	return s
}
