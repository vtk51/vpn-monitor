//go:build windows

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var user32 = syscall.NewLazyDLL("user32.dll")
var gdi = syscall.NewLazyDLL("gdi32.dll")
var createWindow = user32.NewProc("CreateWindowExW")
var defWindow = user32.NewProc("DefWindowProcW")
var sendMessage = user32.NewProc("SendMessageW")
var setText = user32.NewProc("SetWindowTextW")
var guiConfig = newLiveConfig()
var guiState = newMonitorState()
var cancelMonitor context.CancelFunc
var statusLabel, rowsLabel, journalEdit, detailsButton, diagnostic uintptr
var pinButton, journalButton, visibleJournal, appSmallIcon, tooltip uintptr
var journalExpanded bool
var editBrush uintptr
var dpi = 96
var lastJournal string
var toolText = map[uintptr]*uint16{}
var toolValues = map[uintptr]string{}

type rect struct{ L, T, R, B int32 }
type windowClass struct {
	Size, Style                        uint32
	Proc                               uintptr
	ClassExtra, WindowExtra            int32
	Instance, Icon, Cursor, Background uintptr
	Menu, Name                         *uint16
	SmallIcon                          uintptr
}
type message struct {
	Window         uintptr
	ID             uint32
	WParam, LParam uintptr
	Time           uint32
	X, Y           int32
	Private        uint32
}
type toolInfo struct {
	Size, Flags     uint32
	Window, ID      uintptr
	Rect            rect
	Instance        uintptr
	Text            *uint16
	Param, Reserved uintptr
}

func wide(s string) *uint16          { return syscall.StringToUTF16Ptr(s) }
func sp(n int) int                   { return (n*dpi + 48) / 96 }
func windowText(h uintptr, s string) { setText.Call(h, uintptr(unsafe.Pointer(wide(s)))) }
func child(parent uintptr, class, s string, style uintptr, x, y, w, h, id int) uintptr {
	instance, _, _ := kernel.NewProc("GetModuleHandleW").Call(0)
	exStyle := uintptr(0)
	if class == "EDIT" && style&0x800000 != 0 {
		exStyle = 0x200
	}
	v, _, _ := createWindow.Call(exStyle, uintptr(unsafe.Pointer(wide(class))), uintptr(unsafe.Pointer(wide(s))), 0x50000000|style, uintptr(sp(x)), uintptr(sp(y)), uintptr(sp(w)), uintptr(sp(h)), parent, uintptr(id), instance, 0)
	if v == 0 {
		panic("Create child failed")
	}
	sendMessage.Call(v, 0x30, font(sp(12)), 1)
	return v
}
func hidden(parent uintptr, id int) uintptr {
	h := child(parent, "STATIC", "", 0, 0, 0, 0, 0, id)
	user32.NewProc("ShowWindow").Call(h, 0)
	return h
}
func tool(parent, id uintptr, r rect, s string, isChild, add bool) {
	if !add && toolValues[id] == s {
		return
	}
	toolValues[id] = s
	flags := uint32(0x10)
	if isChild {
		flags |= 1
	}
	r = rect{int32(sp(int(r.L))), int32(sp(int(r.T))), int32(sp(int(r.R))), int32(sp(int(r.B)))}
	ti := toolInfo{Flags: flags, Window: parent, ID: id, Rect: r, Text: wide(s)}
	// V2 TOOLINFO size works with unmanifested system comctl32 v5 as well as v6.
	ti.Size = uint32(unsafe.Offsetof(ti.Reserved))
	toolText[id] = ti.Text
	msg := uintptr(0x439)
	if add {
		msg = 0x432
	}
	sendMessage.Call(tooltip, msg, 0, uintptr(unsafe.Pointer(&ti)))
}
func isPinned(h uintptr) bool {
	v, _, _ := user32.NewProc("GetWindowLongPtrW").Call(h, ^uintptr(19))
	return v&8 != 0
}
func refresh(h uintptr) {
	v := guiState.view(time.Now())
	refreshSettingsPause()
	d := "Полные сообщения: выкл."
	if v.Verbose {
		d = "Полные сообщения: вкл."
	}
	windowText(detailsButton, d)
	pin := "Поверх окон: выключено"
	if isPinned(h) {
		pin = "Поверх окон: включено"
	}
	windowText(pinButton, pin)
	tool(h, pinButton, rect{}, pin, true, false)
	label := "› Журнал событий"
	if journalExpanded {
		label = "⌄ Журнал событий"
	}
	windowText(journalButton, label)
	rows, logs, _ := guiState.snapshot()
	windowText(rowsLabel, strings.Join(rows[:], "\r\n"))
	windowText(journalEdit, strings.Join(logs, "\r\n"))
	journal := strings.Join(v.Journal, "\r\n")
	if journal != lastJournal {
		windowText(visibleJournal, journal)
		end := uintptr(journalCaret(journal))
		sendMessage.Call(visibleJournal, 0xb1, end, end)
		sendMessage.Call(visibleJournal, 0xb7, 0, 0)
		lastJournal = journal
	}
	for i, n := range v.Nodes {
		s := v.Config.identity(i) + ": ожидание"
		if n.Seen {
			if n.OK {
				s = fmt.Sprintf("%s: доступен · %.1f мс", v.Config.identity(i), float64(n.Elapsed.Microseconds())/1000)
			} else {
				s = v.Config.identity(i) + ": нет ответа"
			}
		}
		tool(h, uintptr(500+i), rect{}, s, false, false)
	}
	tool(h, 599, rect{}, "Сбои "+v.Config.identity(3)+" · эпизоды с запуска / смены узла 4", false, false)
	tool(h, 598, rect{}, "График за 60 секунд: "+v.Config.identity(3), false, false)

	b, _ := json.Marshal(v)
	windowText(diagnostic, string(b))
	user32.NewProc("InvalidateRect").Call(h, 0, 0)
}
func setJournal(h uintptr) {
	journalExpanded = !journalExpanded
	size := compactDimensions(journalExpanded)
	user32.NewProc("SetWindowPos").Call(h, 0, 0, 0, uintptr(sp(size.Width)), uintptr(sp(size.Height)), 0x16)
	show := uintptr(0)
	if journalExpanded {
		show = 5
	}
	user32.NewProc("ShowWindow").Call(detailsButton, show)
	user32.NewProc("ShowWindow").Call(visibleJournal, show)
}
func windowProc(h uintptr, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case 1:
		editBrush, _, _ = gdi.NewProc("CreateSolidBrush").Call(rgb(0x10151c))
		rs := compactControls()
		makeControl := func(i int, class, s string, style uintptr, id int) uintptr {
			r := rs[i]
			return child(h, class, s, style, r.X, r.Y, r.W, r.H, id)
		}
		pinButton = makeControl(0, "BUTTON", "Поверх окон: выключено", 0x1000b, 110)
		child(h, "BUTTON", "Настройки", 0x1000b, 243, 6, 28, 27, 115)
		makeControl(1, "BUTTON", "Свернуть", 0x1000b, 113)
		makeControl(2, "BUTTON", "Закрыть", 0x1000b, 114)
		journalButton = makeControl(3, "BUTTON", "› Журнал событий", 0x1000b, 111)
		detailsButton = child(h, "BUTTON", "Полные сообщения: выкл.", 0x1000b, 14, 317, 204, 25, 108)
		visibleJournal = child(h, "EDIT", "", 0x00300844, 14, 351, 352, 108, 112)
		sendMessage.Call(visibleJournal, 0x30, font(sp(11)), 1)
		user32.NewProc("ShowWindow").Call(detailsButton, 0)
		user32.NewProc("ShowWindow").Call(visibleJournal, 0)
		statusLabel = hidden(h, 103)
		rowsLabel = hidden(h, 104)
		journalEdit = hidden(h, 105)
		diagnostic = hidden(h, 109)
		windowText(statusLabel, fmt.Sprintf("Действующий таймаут: %d мс", guiConfig.timeout().Milliseconds()))
		syscall.NewLazyDLL("comctl32.dll").NewProc("InitCommonControls").Call()
		tooltip, _, _ = createWindow.Call(8, uintptr(unsafe.Pointer(wide("tooltips_class32"))), 0, 0x80000003, 0, 0, 0, 0, h, 0, 0, 0)
		sendMessage.Call(tooltip, 0x418, 0, uintptr(sp(420)))
		for _, id := range []int{110, 111, 113, 114, 115} {
			ch, _, _ := user32.NewProc("GetDlgItem").Call(h, uintptr(id))
			var b [128]uint16
			user32.NewProc("GetWindowTextW").Call(ch, uintptr(unsafe.Pointer(&b[0])), 128)
			tool(h, ch, rect{}, syscall.UTF16ToString(b[:]), true, true)
		}
		tool(h, 599, rect{285, 47, 366, 72}, "Эпизоды сбоя узла 4", false, true)
		tool(h, 598, rect{14, 232, 366, 267}, "График узла 4", false, true)
		for i := 0; i < 5; i++ {
			tool(h, uintptr(500+i), rect{14, int32(80 + i*29), 366, int32(109 + i*29)}, "Ожидание", false, true)
		}
		user32.NewProc("SetTimer").Call(h, 1, 200, 0)
		return 0
	case 0x84:
		var p struct{ X, Y int32 }
		p.X = int32(int16(lp & 0xffff))
		p.Y = int32(int16(lp >> 16))
		user32.NewProc("ScreenToClient").Call(h, uintptr(unsafe.Pointer(&p)))
		if p.Y < int32(sp(38)) && p.X < int32(sp(240)) {
			return 2
		}
	case 0x111:
		if wp>>16 != 0 {
			break
		}
		switch wp & 0xffff {

		case 108:
			guiState.toggleVerbose()
		case 110:
			after, flags := pinRequest(!isPinned(h))
			user32.NewProc("SetWindowPos").Call(h, after, 0, 0, 0, 0, flags)
		case 111:
			setJournal(h)
		case 113:
			user32.NewProc("ShowWindow").Call(h, 6)
		case 114:
			user32.NewProc("PostMessageW").Call(h, 0x10, 0, 0)
		case 115:
			openSettings(h)
		}
		refresh(h)
		return 0
	case 0x133, 0x138:
		gdi.NewProc("SetTextColor").Call(wp, rgb(ink))
		gdi.NewProc("SetBkColor").Call(wp, rgb(0x10151c))
		return editBrush
	case 0x2b:
		type drawItem struct {
			Type, ID, Item, Action, State uint32
			Window, DC                    uintptr
			Rect                          rect
			Data                          uintptr
		}
		var d *drawItem
		*(*uintptr)(unsafe.Pointer(&d)) = lp
		saved, _, _ := gdi.NewProc("SaveDC").Call(d.DC)
		mapDC(d.DC)
		w, hh := int(d.Rect.R)*96/dpi, int(d.Rect.B)*96/dpi
		c := bg
		border := bg
		col := muted
		if d.ID == 110 || d.ID == 113 || d.ID == 114 || d.ID == 115 {
			c = 0x20252d
			border = c
		}

		if d.ID == 110 && isPinned(h) {
			c = 0x304462
			border = 0x486891
			col = 0x9cc5ff
		}
		if d.State&1 != 0 {
			c = 0x353e4b
		}
		if d.State&0x10 != 0 {
			border = blue
		}
		fill(d.DC, 0, 0, w, hh, c)
		box(d.DC, 0, 0, w, hh, c, border)
		switch d.ID {
		case 115:
			textCenter(d.DC, 0, 0, w, hh, "⚙", 18, col)
		case 110:
			for _, a := range [][4]int{{15, 5, 22, 12}, {15, 5, 13, 9}, {13, 9, 8, 10}, {8, 10, 18, 20}, {18, 20, 19, 15}, {19, 15, 22, 12}, {11, 15, 5, 22}, {12, 13, 17, 8}} {
				stroke(d.DC, a[0], a[1], a[2], a[3], col, 1)
			}
		case 113:
			stroke(d.DC, 9, 16, 19, 16, col, 1)
		case 114:
			stroke(d.DC, 9, 9, 18, 18, col, 1)
			stroke(d.DC, 18, 9, 9, 18, col, 1)
		default:
			var b [128]uint16
			user32.NewProc("GetWindowTextW").Call(d.Window, uintptr(unsafe.Pointer(&b[0])), 128)
			s := syscall.UTF16ToString(b[:])
			size := 11

			textCenter(d.DC, 0, 0, w, hh, s, size, col)
		}
		gdi.NewProc("RestoreDC").Call(d.DC, saved)
		return 1
	case 0x14:
		return 1
	case 0xf:
		type paintStruct struct {
			DC           uintptr
			Erase        int32
			Paint        rect
			Restore, Inc int32
			Reserved     [32]byte
		}
		var ps paintStruct
		dc, _, _ := user32.NewProc("BeginPaint").Call(h, uintptr(unsafe.Pointer(&ps)))
		mem, _, _ := gdi.NewProc("CreateCompatibleDC").Call(dc)
		size := compactDimensions(journalExpanded)
		bitmap, _, _ := gdi.NewProc("CreateCompatibleBitmap").Call(dc, uintptr(sp(size.Width)), uintptr(sp(size.Height)))
		old, _, _ := gdi.NewProc("SelectObject").Call(mem, bitmap)
		mapDC(mem)
		drawDashboard(mem, h)
		gdi.NewProc("SetMapMode").Call(mem, 1)
		gdi.NewProc("BitBlt").Call(dc, 0, 0, uintptr(sp(size.Width)), uintptr(sp(size.Height)), mem, 0, 0, 0xcc0020)
		gdi.NewProc("SelectObject").Call(mem, old)
		gdi.NewProc("DeleteObject").Call(bitmap)
		gdi.NewProc("DeleteDC").Call(mem)
		user32.NewProc("EndPaint").Call(h, uintptr(unsafe.Pointer(&ps)))
		return 0
	case 0x113:
		refresh(h)
		return 0
	case 0x10:
		if cancelMonitor != nil {
			cancelMonitor()
		}
		user32.NewProc("DestroyWindow").Call(h)
		return 0
	case 2:
		if cancelMonitor != nil {
			cancelMonitor()
		}
		user32.NewProc("KillTimer").Call(h, 1)
		user32.NewProc("PostQuitMessage").Call(0)
		return 0
	}
	r, _, _ := defWindow.Call(h, uintptr(msg), wp, lp)
	return r
}
func main() {
	runtime.LockOSThread()
	initializeSettings()
	user32.NewProc("SetProcessDPIAware").Call()
	dc, _, _ := user32.NewProc("GetDC").Call(0)
	v, _, _ := gdi.NewProc("GetDeviceCaps").Call(dc, 88)
	dpi = int(v)
	user32.NewProc("ReleaseDC").Call(0, dc)
	instance, _, _ := kernel.NewProc("GetModuleHandleW").Call(0)
	cursor, _, _ := user32.NewProc("LoadCursorW").Call(0, 32512)
	bigW, _, _ := user32.NewProc("GetSystemMetrics").Call(11)
	bigH, _, _ := user32.NewProc("GetSystemMetrics").Call(12)
	smallW, _, _ := user32.NewProc("GetSystemMetrics").Call(49)
	smallH, _, _ := user32.NewProc("GetSystemMetrics").Call(50)
	bigIcon, _, err := user32.NewProc("LoadImageW").Call(instance, 1, 1, bigW, bigH, 0)
	if bigIcon == 0 {
		showFatal(fmt.Sprint(err))
		return
	}
	defer user32.NewProc("DestroyIcon").Call(bigIcon)
	smallIcon, _, err := user32.NewProc("LoadImageW").Call(instance, 1, 1, smallW, smallH, 0)
	if smallIcon == 0 {
		showFatal(fmt.Sprint(err))
		return
	}
	defer user32.NewProc("DestroyIcon").Call(smallIcon)
	appSmallIcon = smallIcon
	wc := windowClass{Proc: syscall.NewCallback(windowProc), Instance: instance, Icon: bigIcon, SmallIcon: smallIcon, Cursor: cursor, Name: wide("VPNMonitorDashboard")}
	wc.Size = uint32(unsafe.Sizeof(wc))
	atom, _, err := user32.NewProc("RegisterClassExW").Call(uintptr(unsafe.Pointer(&wc)))
	if atom == 0 {
		showFatal(fmt.Sprint(err))
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancelMonitor = cancel
	defer cancel()
	size := compactDimensions(false)
	h, _, err := createWindow.Call(0, uintptr(unsafe.Pointer(wc.Name)), uintptr(unsafe.Pointer(wide("VPN монитор"))), 0x820a0000, 100, 80, uintptr(sp(size.Width)), uintptr(sp(size.Height)), 0, 0, instance, 0)
	if h == 0 {
		showFatal(fmt.Sprint(err))
		return
	}
	cfg, _ := guiConfig.snapshot()
	applyTransparency(h, cfg.TransparencyPercent)
	sendMessage.Call(h, 0x80, 1, bigIcon)
	sendMessage.Call(h, 0x80, 0, smallIcon)
	user32.NewProc("ShowWindow").Call(h, 5)
	user32.NewProc("UpdateWindow").Call(h)
	go monitor(ctx, guiConfig, guiState)
	var m message
	for {
		v, _, _ := user32.NewProc("GetMessageW").Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(v) <= 0 {
			break
		}
		dialog := h
		if settingsWindow != 0 {
			dialog = settingsWindow
		}
		handled, _, _ := user32.NewProc("IsDialogMessageW").Call(dialog, uintptr(unsafe.Pointer(&m)))
		if handled == 0 {
			user32.NewProc("TranslateMessage").Call(uintptr(unsafe.Pointer(&m)))
			user32.NewProc("DispatchMessageW").Call(uintptr(unsafe.Pointer(&m)))
		}
	}
}
func showFatal(s string) {
	user32.NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(wide(s))), uintptr(unsafe.Pointer(wide("VPN монитор — ошибка"))), 0x10)
}
