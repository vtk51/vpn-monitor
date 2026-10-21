//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

var settingsWindow, settingsOwner, settingsError uintptr
var settingsEdits [11]uintptr
var configPath, configLoadError string
var settingsClass uint16
var settingsSlider, settingsPercent, settingsPause uintptr

func applyTransparency(h uintptr, percent int) {
	style, _, _ := user32.NewProc("GetWindowLongPtrW").Call(h, ^uintptr(19))
	user32.NewProc("SetWindowLongPtrW").Call(h, ^uintptr(19), layeredStyle(style))
	user32.NewProc("SetLayeredWindowAttributes").Call(h, 0, uintptr(transparencyAlpha(percent)), 2)
}
func refreshSettingsPause() {
	if settingsPause == 0 {
		return
	}
	label := "Пауза"
	if guiState.view(time.Now()).Paused {
		label = "Продолжить"
	}
	windowText(settingsPause, label)
}
func refreshSettingsPercent() {
	n, _, _ := sendMessage.Call(settingsSlider, 0x400, 0, 0)
	windowText(settingsPercent, fmt.Sprintf("%d%%", n))
}

func readText(h uintptr) string {
	var b [1024]uint16
	user32.NewProc("GetWindowTextW").Call(h, uintptr(unsafe.Pointer(&b[0])), 1024)
	return syscall.UTF16ToString(b[:])
}
func initializeSettings() {
	exe, e := os.Executable()
	if e != nil {
		configLoadError = e.Error()
		return
	}
	configPath = filepath.Join(filepath.Dir(exe), "vpn-monitor.json")
	cfg, e := loadSettings(configPath)
	guiConfig.value = cfg
	guiState.config = cfg
	if e != nil {
		configLoadError = "Не удалось загрузить " + configPath + ": " + e.Error() + ". Используются значения по умолчанию; исходный файл не изменён."
		guiState.event(configLoadError)

	}
}
func settingsFields(c settings) {
	for i, t := range c.Targets {
		windowText(settingsEdits[i*2], t.Name)
		windowText(settingsEdits[i*2+1], t.IP)
	}
	windowText(settingsEdits[8], c.DNSServer)
	windowText(settingsEdits[9], c.DNSDomain)
	windowText(settingsEdits[10], strconv.Itoa(c.TimeoutMS))
	sendMessage.Call(settingsSlider, 0x405, 1, uintptr(c.TransparencyPercent))
	refreshSettingsPercent()
}
func openSettings(owner uintptr) {
	if settingsWindow != 0 {
		user32.NewProc("SetForegroundWindow").Call(settingsWindow)
		return
	}
	instance, _, _ := kernel.NewProc("GetModuleHandleW").Call(0)
	if settingsClass == 0 {
		cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
		wc := windowClass{Proc: syscall.NewCallback(settingsProc), Instance: instance, Cursor: cursor, Background: editBrush, Icon: appSmallIcon, SmallIcon: appSmallIcon, Name: wide("VPNMonitorSettings")}
		wc.Size = uint32(unsafe.Sizeof(wc))
		a, _, _ := user32.NewProc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&wc)))
		settingsClass = uint16(a)
		if a == 0 {
			showFatal("Не удалось создать окно настроек")
			return
		}
	}
	settingsOwner = owner
	var r rect
	user32.NewProc("GetWindowRect").Call(owner, uintptr(unsafe.Pointer(&r)))
	// Fixed client size, adjusted for native caption at the active system DPI.
	wr := rect{0, 0, int32(sp(540)), int32(sp(610))}
	user32.NewProc("AdjustWindowRectEx").Call(uintptr(unsafe.Pointer(&wr)), 0x80c80000, 0, 0)
	h, _, _ := createWindow.Call(0, uintptr(unsafe.Pointer(wide("VPNMonitorSettings"))), uintptr(unsafe.Pointer(wide("VPN монитор — Настройки"))), 0x80c80000, uintptr(r.L+int32(sp(20))), uintptr(r.T+int32(sp(20))), uintptr(wr.R-wr.L), uintptr(wr.B-wr.T), owner, 0, instance, 0)
	if h == 0 {
		showFatal("Не удалось открыть настройки")
		return
	}
	settingsWindow = h
	user32.NewProc("EnableWindow").Call(owner, 0)
	user32.NewProc("ShowWindow").Call(h, 5)
	user32.NewProc("SetForegroundWindow").Call(h)
	user32.NewProc("SetFocus").Call(settingsEdits[0])
}
func closeSettings(h uintptr) {
	user32.NewProc("EnableWindow").Call(settingsOwner, 1)
	user32.NewProc("DestroyWindow").Call(h)
	settingsWindow = 0
	settingsPause = 0
	user32.NewProc("SetForegroundWindow").Call(settingsOwner)
}
func settingsProc(h uintptr, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case 1:
		child(h, "STATIC", "Четыре ICMP-узла · только IPv4", 0, 18, 12, 500, 22, 0)
		child(h, "STATIC", "Имя (до 32 символов)", 0, 42, 40, 235, 20, 0)
		child(h, "STATIC", "IPv4-адрес", 0, 293, 40, 220, 20, 0)
		for i := 0; i < 4; i++ {
			y := 65 + i*34
			child(h, "STATIC", strconv.Itoa(i+1), 0, 18, y+3, 20, 22, 0)
			settingsEdits[i*2] = child(h, "EDIT", "", 0x810080, 42, y, 235, 26, 201+i*2)
			settingsEdits[i*2+1] = child(h, "EDIT", "", 0x810080, 293, y, 225, 26, 202+i*2)
			sendMessage.Call(settingsEdits[i*2], 0xc5, 32, 0)
			sendMessage.Call(settingsEdits[i*2+1], 0xc5, 15, 0)
		}
		for i, t := range []string{"DNS-сервер (IPv4:порт)", "DNS-домен (ASCII / punycode)", "Таймаут (1–60000 мс)"} {
			y := 213 + i*34
			child(h, "STATIC", t, 0, 18, y+3, 260, 24, 0)
			settingsEdits[8+i] = child(h, "EDIT", "", 0x810080, 293, y, 225, 26, 209+i)
		}
		child(h, "STATIC", "Прозрачность главного окна", 0, 18, 319, 275, 24, 0)
		settingsSlider = child(h, "msctls_trackbar32", "", 0x10000, 293, 312, 180, 32, 224)
		sendMessage.Call(settingsSlider, 0x406, 1, 70<<16)
		settingsPercent = child(h, "STATIC", "0%", 0, 480, 319, 48, 24, 225)
		child(h, "STATIC", "0% — непрозрачно · 70% — максимум · после сохранения", 0, 18, 348, 500, 24, 0)
		settingsPause = child(h, "BUTTON", "Пауза", 0x10000, 18, 380, 145, 28, 226)
		child(h, "STATIC", "Сразу; не зависит от «Сохранить» / «Отмена»", 0, 177, 384, 345, 24, 0)
		refreshSettingsPause()
		child(h, "STATIC", "Узел 4 управляет графиком и счётчиком сбоев.\r\nИзменение его имени / IP сбрасывает график и счётчик.\r\nСохранение — рядом с EXE; сеть — со следующего цикла.", 0, 18, 425, 500, 56, 0)
		settingsError = child(h, "STATIC", configLoadError, 0, 18, 487, 500, 77, 220)
		child(h, "BUTTON", "По умолчанию", 0x10000, 18, 572, 145, 27, 223)
		child(h, "BUTTON", "Сохранить", 0x10001, 283, 572, 112, 27, 221)
		child(h, "BUTTON", "Отмена", 0x10000, 406, 572, 112, 27, 222)
		cfg, _ := guiConfig.snapshot()
		settingsFields(cfg)
		return 0
	case 0x111:
		if wp>>16 != 0 {
			break
		}
		switch wp & 0xffff {
		case 226:
			guiState.togglePause()
			refreshSettingsPause()
			refresh(settingsOwner)
		case 2, 222:
			closeSettings(h)
		case 223:
			settingsFields(defaultSettings())
			windowText(settingsError, "Значения по умолчанию. Нажмите «Сохранить» для применения.")
		case 1, 221:
			cfg, _ := guiConfig.snapshot()
			for i := range cfg.Targets {
				cfg.Targets[i] = targetConfig{readText(settingsEdits[i*2]), readText(settingsEdits[i*2+1])}
			}
			cfg.DNSServer = readText(settingsEdits[8])
			cfg.DNSDomain = readText(settingsEdits[9])
			n, e := parseTimeout(readText(settingsEdits[10]))
			cfg.TimeoutMS = n
			percent, _, _ := sendMessage.Call(settingsSlider, 0x400, 0, 0)
			cfg.TransparencyPercent = int(percent)
			if e == nil {
				e = cfg.validate()
			}
			if e == nil {
				e = guiConfig.commit(configPath, cfg, guiState)
			}
			if e != nil {
				windowText(settingsError, "Не сохранено: "+e.Error())
				return 0
			}
			configLoadError = ""
			applyTransparency(settingsOwner, cfg.TransparencyPercent)

			windowText(statusLabel, fmt.Sprintf("Действующий таймаут: %d мс", n))
			closeSettings(h)
			refresh(settingsOwner)
		}
		return 0
	case 0x114:
		refreshSettingsPercent()
		return 0
	case 0x133, 0x138:
		gdi.NewProc("SetTextColor").Call(wp, rgb(ink))
		gdi.NewProc("SetBkColor").Call(wp, rgb(0x10151c))
		return editBrush
	case 0x10:
		closeSettings(h)
		return 0
	}
	r, _, _ := defWindow.Call(h, uintptr(msg), wp, lp)
	return r
}
