package main

import (
	"os"
	"runtime"
	"runtime/debug"
	"sync/atomic"
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
	suppressUp     bool        // the pending tray button-up belongs to a click that just closed the popup
	needRescan     atomic.Bool // display handles may be stale; rediscover before the next popup
	taskbarCreated uint32
}

func trimMemory() {
	debug.FreeOSMemory()
	trimWorkingSet()
}

// trimMemoryAsync keeps the forced GC and working-set trim off the UI thread's hot path.
func trimMemoryAsync() {
	go func() {
		time.Sleep(200 * time.Millisecond)
		trimMemory()
	}()
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

	// Probing monitors over DDC/CI can take a while; the tray icon is already up by now.
	app.displays = discoverAll()

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
	app.needRescan.Store(false)
}

const (
	pbtApmResumeSuspend   = 0x0007
	pbtApmResumeAutomatic = 0x0012
)

func trayProc(hwnd, m, wp, lp uintptr) uintptr {
	switch {
	case m == wmTray:
		switch lp {
		case wmLButtonDown:
			onTrayButtonDown()
		case wmLButtonUp:
			onTrayButtonUp()
		case wmRButtonUp:
			showTrayMenu()
		}
		return 0
	case m == wmDisplayChange:
		app.needRescan.Store(true)
		return 0
	case m == wmPowerBroadcast:
		if wp == pbtApmResumeSuspend || wp == pbtApmResumeAutomatic {
			app.needRescan.Store(true) // monitors may have dropped their DDC state while asleep
		}
		return 1
	case m == wmClose:
		// Graceful shutdown used by the installer: drop the tray icon, then leave the message loop.
		postQuitMessage(0)
		return 0
	case uint32(m) == app.taskbarCreated && app.taskbarCreated != 0:
		addTrayIcon()
		return 0
	}
	return defWindowProc(hwnd, m, wp, lp)
}

// Pressing the tray icon steals focus from the popup, so by the time the button comes up the
// popup is already gone. Remember on button-down that this click closed it, and swallow the
// matching button-up however long the press lasts.
func onTrayButtonDown() {
	app.suppressUp = app.popup != nil || time.Since(app.lastHide) < 200*time.Millisecond
	if app.popup != nil {
		hidePopup()
	}
}

func onTrayButtonUp() {
	if app.suppressUp {
		app.suppressUp = false
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
