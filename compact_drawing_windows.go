//go:build windows

package main

import (
	"fmt"
	"time"
	"unsafe"
)

const (
	bg        = 0x171b21
	panel     = 0x20252d
	lineColor = 0x303641
	ink       = 0xe4e9f1
	muted     = 0x8d9aab
	blue      = 0x89b7f1
	green     = 0x72d9aa
	red       = 0xff8888
)

var fonts = map[int]uintptr{}

func rgb(c int) uintptr { return uintptr((c&255)<<16 | (c & 0xff00) | (c>>16)&255) }
func mapDC(dc uintptr) {
	gdi.NewProc("SetMapMode").Call(dc, 8)
	gdi.NewProc("SetWindowExtEx").Call(dc, 96, 96, 0)
	gdi.NewProc("SetViewportExtEx").Call(dc, uintptr(dpi), uintptr(dpi), 0)
}
func font(size int) uintptr {
	if f := fonts[size]; f != 0 {
		return f
	}
	weight := 400
	n := size
	if n < 0 {
		weight = 600
		n = -n
	}
	height := int32(-n)
	f, _, _ := gdi.NewProc("CreateFontW").Call(uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(wide("Segoe UI"))))
	fonts[size] = f
	return f
}
func fill(dc uintptr, x, y, w, h, c int) {
	r := rect{int32(x), int32(y), int32(x + w), int32(y + h)}
	b, _, _ := gdi.NewProc("CreateSolidBrush").Call(rgb(c))
	user32.NewProc("FillRect").Call(dc, uintptr(unsafe.Pointer(&r)), b)
	gdi.NewProc("DeleteObject").Call(b)
}
func label(dc uintptr, x, y, w, h int, s string, size, c int, flags uintptr) {
	gdi.NewProc("SelectObject").Call(dc, font(size))
	gdi.NewProc("SetBkMode").Call(dc, 1)
	gdi.NewProc("SetTextColor").Call(dc, rgb(c))
	r := rect{int32(x), int32(y), int32(x + w), int32(y + h)}
	user32.NewProc("DrawTextW").Call(dc, uintptr(unsafe.Pointer(wide(s))), ^uintptr(0), uintptr(unsafe.Pointer(&r)), flags|0x24|0x8000)
}
func text(dc uintptr, x, y, w, h int, s string, size, c int) { label(dc, x, y, w, h, s, size, c, 0) }
func textCenter(dc uintptr, x, y, w, h int, s string, size, c int) {
	label(dc, x, y, w, h, s, size, c, 1)
}
func stroke(dc uintptr, x, y, x2, y2, c, width int) {
	pen, _, _ := gdi.NewProc("CreatePen").Call(0, uintptr(width), rgb(c))
	old, _, _ := gdi.NewProc("SelectObject").Call(dc, pen)
	gdi.NewProc("MoveToEx").Call(dc, uintptr(x), uintptr(y), 0)
	gdi.NewProc("LineTo").Call(dc, uintptr(x2), uintptr(y2))
	gdi.NewProc("SelectObject").Call(dc, old)
	gdi.NewProc("DeleteObject").Call(pen)
}
func box(dc uintptr, x, y, w, h, c, border int) {
	b, _, _ := gdi.NewProc("CreateSolidBrush").Call(rgb(c))
	p, _, _ := gdi.NewProc("CreatePen").Call(0, 1, rgb(border))
	ob, _, _ := gdi.NewProc("SelectObject").Call(dc, b)
	op, _, _ := gdi.NewProc("SelectObject").Call(dc, p)
	gdi.NewProc("RoundRect").Call(dc, uintptr(x), uintptr(y), uintptr(x+w), uintptr(y+h), 5, 5)
	gdi.NewProc("SelectObject").Call(dc, ob)
	gdi.NewProc("SelectObject").Call(dc, op)
	gdi.NewProc("DeleteObject").Call(b)
	gdi.NewProc("DeleteObject").Call(p)
}
func drawDashboard(dc, h uintptr) {
	v := guiState.view(time.Now())
	size := compactDimensions(journalExpanded)
	fill(dc, 0, 0, 380, size.Height, bg)
	fill(dc, 0, 0, 380, 38, panel)
	stroke(dc, 0, 38, 380, 38, lineColor, 1)
	user32.NewProc("DrawIconEx").Call(dc, 12, 9, appSmallIcon, 20, 20, 0, 0, 3)
	text(dc, 40, 0, 220, 38, "VPN монитор", -12, ink)
	seen, ok := 0, 0
	last := time.Time{}
	for _, n := range v.Nodes {
		if n.Seen {
			seen++
		}
		if n.OK {
			ok++
		}
		if n.At.After(last) {
			last = n.At
		}
	}
	headline := "Ожидание проверки…"
	col := muted
	if seen > 0 {
		headline = fmt.Sprintf("Доступны %d из 5 узлов", ok)
		col = red
		if ok == 5 {
			headline = "Все узлы доступны"
			col = green
		} else if v.Nodes[3].Seen && !v.Nodes[3].OK {
			headline = v.Config.Targets[3].Name + " не отвечает"
		}
	}
	if v.Paused {
		headline = "Мониторинг на паузе"
		col = muted
	}
	text(dc, 14, 48, 12, 22, "●", 10, col)
	text(dc, 28, 48, 251, 22, headline, -13, col)
	label(dc, 282, 48, 84, 22, fmt.Sprintf("Сбоев %d", v.YouTubeOutages), 11, ink, 2)
	for i, n := range v.Nodes {
		y := 80 + i*29
		c := muted
		s := "ожидание"
		if n.Seen {
			if n.OK {
				c = green
				s = fmt.Sprintf("%.0f мс", float64(n.Elapsed.Microseconds())/1000)
			} else {
				c = red
				s = "нет ответа"
			}
		}
		if i == 3 {
			row := 0x1c2632
			if n.Seen && !n.OK {
				row = 0x302126
			}
			fill(dc, 14, y, 352, 29, row)
		}
		stroke(dc, 14, y+28, 366, y+28, 0x272e38, 1)
		text(dc, 14, y, 12, 29, "●", 9, c)
		text(dc, 29, y, 82, 29, v.Config.node(i).Name, 12, ink)
		addr := v.Config.node(i).IP
		text(dc, 112, y, 157, 29, addr, 10, 0x788597)
		label(dc, 274, y, 92, 29, s, 12, c, 2)
	}
	text(dc, 14, 235, 64, 15, v.Config.Targets[3].Name, 10, muted)
	text(dc, 14, 250, 64, 15, "60 секунд", 10, 0xb4c0d2)
	stroke(dc, 87, 259, 366, 259, 0x29323e, 1)
	maxMS := float64(100)
	good := 0
	for _, s := range v.Samples {
		if s.OK {
			good++
			ms := float64(s.Elapsed.Microseconds()) / 1000
			if ms > maxMS {
				maxMS = ms
			}
		}
	}
	px, py := 0, 0
	prev := false
	now := time.Now()
	for _, s := range v.Samples {
		x := 366 - int(now.Sub(s.At).Seconds()*279/60)
		if x < 87 {
			x = 87
		}
		if !s.OK {
			stroke(dc, x, 258, x, 264, red, 1)
			prev = false
			continue
		}
		y := 259 - int(float64(s.Elapsed.Microseconds())/1000/maxMS*24)
		if prev {
			stroke(dc, px, py, x, y, blue, 1)
		}
		fill(dc, x-1, y-1, 2, 2, blue)
		px = x
		py = y
		prev = true
	}
	if good == 0 {
		text(dc, 89, 235, 277, 27, "Нет успешных ответов", 10, muted)
	}
	stroke(dc, 14, 273, 366, 273, lineColor, 1)
	text(dc, 14, 281, 58, 25, "Таймаут", 11, muted)
	box(dc, 75, 280, 62, 27, 0x10151c, 0x3c4758)
	text(dc, 142, 281, 18, 25, "мс", 11, muted)
	feedback := ""
	hc := muted
	if hintError {
		feedback = "Ошибка · см. подсказку"
		hc = red
	} else if settingsHint != "Применяется со следующей проверки" {
		feedback = "Сохранено"
		hc = green
	}
	text(dc, 201, 281, 125, 25, feedback, 10, hc)
	fill(dc, 0, 319, 380, 34, 0x1c2129)
	stroke(dc, 0, 319, 380, 319, lineColor, 1)
	stamp := "Ожидание…"
	if !last.IsZero() {
		stamp = "Обновлено " + last.Format("15:04:05")
	}
	label(dc, 190, 323, 175, 26, stamp, 10, muted, 2)
	if journalExpanded {
		fill(dc, 1, 353, 378, 183, 0x14191f)
		stroke(dc, 0, 353, 380, 353, lineColor, 1)
		text(dc, 14, 509, 352, 21, "Счётчик с запуска · лог: vpn-diag.log", 10, muted)
	}
	border := 0x3a424e
	if isPinned(h) {
		border = 0x5e83b5
	}
	stroke(dc, 0, 0, 379, 0, border, 1)
	stroke(dc, 0, 0, 0, size.Height-1, border, 1)
	stroke(dc, 379, 0, 379, size.Height-1, border, 1)
	stroke(dc, 0, size.Height-1, 380, size.Height-1, border, 1)
}
