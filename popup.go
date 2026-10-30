package main

import (
	"math"
	"strconv"
	"syscall"
	"time"
	"unsafe"
)

const (
	popupClass      = "BrightSlidePopup"
	deactivateGrace = 250 * time.Millisecond
	graceTimerID    = 1
	supersample     = 4

	colBg         = 0x202020
	colTitle      = 0xFFFFFF
	colText       = 0xCCCCCC
	colHint       = 0x999999
	colTrackEmpty = 0x414141
	colThumb      = 0xFFFFFF
)

type sliderRow struct{ nameY, sliderY int32 }

type popupState struct {
	hwnd                uintptr
	scale               float64
	w, h                int32
	pad                 int32
	sliderH             int32
	thumbR              int32
	trackH              int32
	x0, x1              int32
	rows                []sliderRow
	shownAt             time.Time
	drag                int
	titleFont, bodyFont uintptr

	// Drawing buffers live as long as the popup so a drag doesn't allocate GDI objects per mouse move.
	backDC, backBmp, backOld uintptr
	ssDC, ssBmp, ssOld       uintptr
}

func (p *popupState) s(n float64) int32 { return int32(math.Round(n * p.scale)) }

func (p *popupState) layout(count int) {
	p.pad = p.s(18)
	p.w = p.s(300)
	p.sliderH = p.s(20)
	p.thumbR = p.s(8)
	p.trackH = p.s(6)
	p.x0 = p.pad + p.thumbR
	p.x1 = p.w - p.pad - p.thumbR

	y := p.pad + p.s(20) + p.s(12)
	p.rows = p.rows[:0]
	for i := 0; i < count; i++ {
		nameY := y
		sliderY := nameY + p.s(16) + p.s(6)
		p.rows = append(p.rows, sliderRow{nameY, sliderY})
		y = sliderY + p.sliderH + p.s(14)
	}
	if count == 0 {
		y += p.s(16) + p.s(14)
	}
	p.h = y - p.s(14) + p.pad
}

func registerPopupClass() {
	wc := wndClassEx{
		style:         csDropShadow,
		lpfnWndProc:   syscall.NewCallback(popupProc),
		hInstance:     app.hInst,
		hCursor:       loadCursor(idcArrow),
		lpszClassName: u16(popupClass),
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	registerClassEx(&wc)
}

func showPopup() {
	if app.popup != nil {
		return
	}
	if app.needRescan.Load() {
		rescanDisplays()
	}

	var cur point
	getCursorPos(&cur)
	hMon := monitorFromPoint(cur, 2)
	var mi monitorInfoEx
	mi.cbSize = uint32(unsafe.Sizeof(mi))
	getMonitorInfo(hMon, &mi)
	dpi := uint32(96)
	var dx, dy uint32
	if getDpiForMonitor(hMon, &dx, &dy) && dx != 0 {
		dpi = dx
	}

	p := &popupState{scale: float64(dpi) / 96, drag: -1}
	p.layout(len(app.displays))

	x := cur.x - p.w
	y := cur.y - p.h - p.s(12)
	if y < mi.rcWork.top {
		y = cur.y + p.s(12)
	}
	x = int32(clamp(int(x), int(mi.rcWork.left), int(mi.rcWork.right-p.w)))
	y = int32(clamp(int(y), int(mi.rcWork.top), int(mi.rcWork.bottom-p.h)))

	face := u16("Segoe UI")
	p.titleFont = createFont(-p.s(14), 600, face)
	p.bodyFont = createFont(-p.s(12), 400, face)

	app.popup = p
	p.hwnd = createWindowEx(wsExTopmost|wsExToolWindow, u16(popupClass), u16("밝기 조절"), wsPopup, x, y, p.w, p.h, 0, 0, app.hInst, 0)
	if p.hwnd == 0 {
		app.popup = nil
		deleteObject(p.titleFont)
		deleteObject(p.bodyFont)
		return
	}
	round := uint32(2) // DWMWCP_ROUND
	dwmSetWindowAttribute(p.hwnd, 33, &round)
	border := uint32(0x404040)
	dwmSetWindowAttribute(p.hwnd, 34, &border)

	showWindow(p.hwnd, swShow)
	bringToForeground(p.hwnd)
	p.shownAt = time.Now()
	// Deactivation inside the grace period is ignored, so check once more afterwards: if the
	// popup never got (or already lost) focus, nothing else would ever dismiss it.
	setTimer(p.hwnd, graceTimerID, uint32(deactivateGrace/time.Millisecond)+50)
}

func hidePopup() {
	p := app.popup
	if p == nil {
		return
	}
	app.popup = nil
	releaseCapture()
	killTimer(p.hwnd, graceTimerID)
	destroyWindow(p.hwnd)
	p.freeBuffers()
	deleteObject(p.titleFont)
	deleteObject(p.bodyFont)
	app.lastHide = time.Now()
	trimMemoryAsync()
}

func (p *popupState) ensureBuffers(hdc uintptr) {
	if p.backDC != 0 {
		return
	}
	p.backDC = createCompatibleDC(hdc)
	p.backBmp = createCompatibleBitmap(hdc, p.w, p.h)
	p.backOld = selectObject(p.backDC, p.backBmp)

	sliderW := p.w - 2*p.pad
	p.ssDC = createCompatibleDC(hdc)
	p.ssBmp = createCompatibleBitmap(hdc, sliderW*supersample, p.sliderH*supersample)
	p.ssOld = selectObject(p.ssDC, p.ssBmp)
}

func (p *popupState) freeBuffers() {
	if p.backDC != 0 {
		selectObject(p.backDC, p.backOld)
		deleteObject(p.backBmp)
		deleteDC(p.backDC)
	}
	if p.ssDC != 0 {
		selectObject(p.ssDC, p.ssOld)
		deleteObject(p.ssBmp)
		deleteDC(p.ssDC)
	}
	p.backDC, p.ssDC = 0, 0
}

func popupProc(hwnd, m, wp, lp uintptr) uintptr {
	p := app.popup
	if p == nil {
		return defWindowProc(hwnd, m, wp, lp)
	}
	switch m {
	case wmEraseBkgnd:
		return 1
	case wmPaint:
		p.paint(hwnd)
		return 0
	case wmLButtonDown:
		x, y := int32(int16(lp&0xFFFF)), int32(int16(lp>>16&0xFFFF))
		for i, r := range p.rows {
			if y >= r.sliderY-p.s(6) && y <= r.sliderY+p.sliderH+p.s(6) {
				p.drag = i
				setCapture(hwnd)
				p.update(i, x)
				break
			}
		}
		return 0
	case wmMouseMove:
		if p.drag >= 0 && wp&mkLButton != 0 {
			p.update(p.drag, int32(int16(lp&0xFFFF)))
		}
		return 0
	case wmLButtonUp:
		if p.drag >= 0 {
			p.drag = -1
			releaseCapture()
		}
		return 0
	case wmKeyDown:
		if wp == vkEscape {
			hidePopup()
		}
		return 0
	case wmActivate:
		if wp&0xFFFF == 0 && time.Since(p.shownAt) > deactivateGrace {
			hidePopup()
		}
		return 0
	case wmTimer:
		if wp == graceTimerID {
			killTimer(hwnd, graceTimerID)
			if getForegroundWindow() != hwnd && !bringToForeground(hwnd) {
				hidePopup()
			}
			return 0
		}
	}
	return defWindowProc(hwnd, m, wp, lp)
}

func (p *popupState) update(i int, mouseX int32) {
	if i >= len(app.displays) {
		return
	}
	v := int(math.Round(float64(mouseX-p.x0) / float64(p.x1-p.x0) * 100))
	v = clamp(v, 0, 100)
	d := app.displays[i]
	if v == d.value {
		return
	}
	d.value = v
	d.wr.set(v)
	r := p.rows[i]
	row := rect{0, r.nameY, p.w, r.sliderY + p.sliderH}
	invalidateRect(p.hwnd, &row)
}

func fillSolid(dc uintptr, r rect, color uint32) {
	b := createSolidBrush(color)
	fillRect(dc, &r, b)
	deleteObject(b)
}

func (p *popupState) paint(hwnd uintptr) {
	var ps paintStruct
	hdc := beginPaint(hwnd, &ps)
	defer endPaint(hwnd, &ps)

	p.ensureBuffers(hdc)
	mem := p.backDC
	fillSolid(mem, rect{0, 0, p.w, p.h}, colBg)
	setBkMode(mem, transparentBk)

	oldFont := selectObject(mem, p.titleFont)
	setTextColor(mem, colTitle)
	title := rect{p.pad, p.pad, p.w - p.pad, p.pad + p.s(20)}
	drawText(mem, u16("디스플레이 밝기"), &title, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)

	selectObject(mem, p.bodyFont)
	if len(app.displays) == 0 {
		setTextColor(mem, colHint)
		y := p.pad + p.s(20) + p.s(12)
		hint := rect{p.pad, y, p.w - p.pad, y + p.s(16)}
		drawText(mem, u16("감지된 디스플레이가 없습니다"), &hint, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix)
	}
	setTextColor(mem, colText)
	for i, d := range app.displays {
		r := p.rows[i]
		name := rect{p.pad, r.nameY, p.w - p.pad - p.s(44), r.nameY + p.s(16)}
		drawText(mem, u16(d.name), &name, dtLeft|dtVCenter|dtSingleLine|dtNoPrefix|dtEndEllipsis)
		pct := rect{p.w - p.pad - p.s(44), r.nameY, p.w - p.pad, r.nameY + p.s(16)}
		drawText(mem, u16(strconv.Itoa(d.value)+"%"), &pct, dtRight|dtVCenter|dtSingleLine|dtNoPrefix)
		p.drawSlider(mem, p.pad, r.sliderY, p.w-2*p.pad, p.sliderH, d.value)
	}
	selectObject(mem, oldFont)

	bitBlt(hdc, 0, 0, p.w, p.h, mem, 0, 0, srcCopy)
}

// drawSlider renders at supersample x and averages down so the capsule and thumb are smooth.
func (p *popupState) drawSlider(dst uintptr, x, y, w, h int32, value int) {
	const ss = supersample
	sdc := p.ssDC
	fillSolid(sdc, rect{0, 0, w * ss, h * ss}, colBg)
	oldPen := selectObject(sdc, getStockObject(nullPen))

	r := p.thumbR * ss
	cy := h * ss / 2
	half := p.trackH * ss / 2
	tx0, tx1 := r, w*ss-r
	tx := tx0 + int32(int64(tx1-tx0)*int64(value)/100)

	empty := createSolidBrush(colTrackEmpty)
	old := selectObject(sdc, empty)
	roundRect(sdc, tx0-half, cy-half, tx1+half+1, cy+half+1, p.trackH*ss, p.trackH*ss)
	selectObject(sdc, old)
	deleteObject(empty)

	white := createSolidBrush(colThumb)
	old = selectObject(sdc, white)
	roundRect(sdc, tx0-half, cy-half, tx+1, cy+half+1, p.trackH*ss, p.trackH*ss)
	ellipse(sdc, tx-r, cy-r, tx+r+1, cy+r+1)
	selectObject(sdc, old)
	deleteObject(white)

	setStretchBltMode(dst, halftone)
	setBrushOrgEx(dst, 0, 0)
	stretchBlt(dst, x, y, w, h, sdc, 0, 0, w*ss, h*ss, srcCopy)

	selectObject(sdc, oldPen)
}
