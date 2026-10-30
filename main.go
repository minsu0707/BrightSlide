package main

import (
	"os"
	"runtime"
	"runtime/debug"
	"syscall"
	"time"
	"unsafe"
)

const (
	trayClass     = "BrightSlideTray"
	wmTray        = wmApp + 1
	trayIconID    = 1
	menuAutostart = 1
	menuExit      = 2
)

var app struct {
	hInst          uintptr
	trayHwnd       uintptr
	icon           uintptr
	displays       []*display
	popup          *popupState
	lastHide       time.Time
	needRescan     bool
	taskbarCreated uint32
}

func trimMemory() {
	debug.FreeOSMemory()
	trimWorkingSet()
}

func main() {
	runtime.LockOSThread()
	debug.SetGCPercent(20)
	setProcessDpiAwarenessContext(^uintptr(3)) // per-monitor v2

	if createMutex(u16(`Local\BrightSlide-SingleInstance`)) {
		return
	}

	app.hInst = getModuleHandle()
	app.taskbarCreated = registerWindowMessage(u16("TaskbarCreated"))
	app.displays = discoverAll()

	registerPopupClass()
	wc := wndClassEx{
		lpfnWndProc:   syscall.NewCallback(trayProc),
		hInstance:     app.hInst,
		lpszClassName: u16(trayClass),
	}
	wc.cbSize = uint32(unsafe.Sizeof(wc))
	registerClassEx(&wc)
	app.trayHwnd = createWindowEx(wsExToolWindow, u16(trayClass), u16("BrightSlide"), wsPopup, 0, 0, 0, 0, 0, 0, app.hInst, 0)

	cx, cy := getSystemMetrics(smCxSmIcon), getSystemMetrics(smCySmIcon)
	app.icon = loadImage(app.hInst, 1, imageIcon, cx, cy, 0)
	if app.icon == 0 {
		app.icon = loadImage(0, idiApplication, imageIcon, cx, cy, 0x8000) // LR_SHARED
	}
	addTrayIcon()

	// Developer aid: open the popup right away instead of waiting for a tray click.
	if len(os.Args) > 1 && os.Args[1] == "--show" {
		showPopup()
	}
	trimMemory()

	var m msg
	for getMessage(&m) > 0 {
		translateMessage(&m)
		dispatchMessage(&m)
	}
	removeTrayIcon()
	closeDisplays(app.displays)
}

func newTrayData() notifyIconData {
	var nid notifyIconData
	nid.cbSize = uint32(unsafe.Sizeof(nid))
	nid.hWnd = app.trayHwnd
	nid.uID = trayIconID
	return nid
}

func addTrayIcon() {
	nid := newTrayData()
	nid.uFlags = nifMessage | nifIcon | nifTip
	nid.uCallbackMessage = wmTray
	nid.hIcon = app.icon
	tip, _ := syscall.UTF16FromString("BrightSlide — 모니터 밝기 조절")
	copy(nid.szTip[:len(nid.szTip)-1], tip)
	shellNotifyIcon(nimAdd, &nid)
}

func removeTrayIcon() {
	nid := newTrayData()
	shellNotifyIcon(nimDelete, &nid)
}

func rescanDisplays() {
	closeDisplays(app.displays)
	app.displays = discoverAll()
	app.needRescan = false
}

func trayProc(hwnd, m, wp, lp uintptr) uintptr {
	switch {
	case m == wmTray:
		switch lp {
		case wmLButtonUp:
			onTrayClick()
		case wmRButtonUp:
			showTrayMenu()
		}
		return 0
	case m == wmDisplayChange:
		app.needRescan = true
		return 0
	case uint32(m) == app.taskbarCreated && app.taskbarCreated != 0:
		addTrayIcon()
		return 0
	}
	return defWindowProc(hwnd, m, wp, lp)
}

func onTrayClick() {
	if app.popup != nil {
		hidePopup()
		return
	}
	// The click that just dismissed the popup (focus loss) must not reopen it.
	if time.Since(app.lastHide) < 300*time.Millisecond {
		return
	}
	showPopup()
}

func showTrayMenu() {
	var cur point
	getCursorPos(&cur)
	menu := createPopupMenu()
	flags := uint32(mfString)
	if autostartEnabled() {
		flags |= mfChecked
	}
	appendMenu(menu, flags, menuAutostart, u16("시작 시 자동 실행"))
	appendMenu(menu, mfSeparator, 0, nil)
	appendMenu(menu, mfString, menuExit, u16("종료"))

	setForegroundWindow(app.trayHwnd)
	cmd := trackPopupMenu(menu, tpmRightButton|tpmReturnCmd|tpmNoNotify|tpmBottomAlign|tpmRightAlign, cur.x, cur.y, app.trayHwnd)
	postMessage(app.trayHwnd, 0, 0, 0)
	destroyMenu(menu)

	switch cmd {
	case menuAutostart:
		setAutostart(!autostartEnabled())
	case menuExit:
		postQuitMessage(0)
	}
	trimMemory()
}
