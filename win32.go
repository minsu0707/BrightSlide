package main

import (
	"syscall"
	"unsafe"
)

// Every wrapper converts pointers with uintptr(unsafe.Pointer(..)) directly inside the
// SyscallN call so the referenced memory stays valid for the duration of the call.

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")
	dxva2    = syscall.NewLazyDLL("dxva2.dll")
	dwmapi   = syscall.NewLazyDLL("dwmapi.dll")
	shcore   = syscall.NewLazyDLL("shcore.dll")

	pRegisterClassEx     = user32.NewProc("RegisterClassExW")
	pCreateWindowEx      = user32.NewProc("CreateWindowExW")
	pDefWindowProc       = user32.NewProc("DefWindowProcW")
	pDestroyWindow       = user32.NewProc("DestroyWindow")
	pShowWindow          = user32.NewProc("ShowWindow")
	pSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	pGetMessage          = user32.NewProc("GetMessageW")
	pTranslateMessage    = user32.NewProc("TranslateMessage")
	pDispatchMessage     = user32.NewProc("DispatchMessageW")
	pPostQuitMessage     = user32.NewProc("PostQuitMessage")
	pPostMessage         = user32.NewProc("PostMessageW")
	pGetCursorPos        = user32.NewProc("GetCursorPos")
	pSetCapture          = user32.NewProc("SetCapture")
	pReleaseCapture      = user32.NewProc("ReleaseCapture")
	pLoadCursor          = user32.NewProc("LoadCursorW")
	pLoadImage           = user32.NewProc("LoadImageW")
	pInvalidateRect      = user32.NewProc("InvalidateRect")
	pSetTimer            = user32.NewProc("SetTimer")
	pKillTimer           = user32.NewProc("KillTimer")
	pGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	pGetWindowThreadPID  = user32.NewProc("GetWindowThreadProcessId")
	pAttachThreadInput   = user32.NewProc("AttachThreadInput")
	pGetCurrentThreadID  = kernel32.NewProc("GetCurrentThreadId")
	pBeginPaint          = user32.NewProc("BeginPaint")
	pEndPaint            = user32.NewProc("EndPaint")
	pFillRect            = user32.NewProc("FillRect")
	pDrawText            = user32.NewProc("DrawTextW")
	pCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	pAppendMenu          = user32.NewProc("AppendMenuW")
	pTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	pDestroyMenu         = user32.NewProc("DestroyMenu")
	pRegisterWindowMsg   = user32.NewProc("RegisterWindowMessageW")
	pMonitorFromPoint    = user32.NewProc("MonitorFromPoint")
	pGetMonitorInfo      = user32.NewProc("GetMonitorInfoW")
	pEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")
	pEnumDisplayDevices  = user32.NewProc("EnumDisplayDevicesW")
	pGetSystemMetrics    = user32.NewProc("GetSystemMetrics")
	pSetDpiAwarenessCtx  = user32.NewProc("SetProcessDpiAwarenessContext")

	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
	pBitBlt                 = gdi32.NewProc("BitBlt")
	pStretchBlt             = gdi32.NewProc("StretchBlt")
	pSetStretchBltMode      = gdi32.NewProc("SetStretchBltMode")
	pSetBrushOrgEx          = gdi32.NewProc("SetBrushOrgEx")
	pCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	pGetStockObject         = gdi32.NewProc("GetStockObject")
	pRoundRect              = gdi32.NewProc("RoundRect")
	pEllipse                = gdi32.NewProc("Ellipse")
	pCreateFont             = gdi32.NewProc("CreateFontW")
	pSetBkMode              = gdi32.NewProc("SetBkMode")
	pSetTextColor           = gdi32.NewProc("SetTextColor")

	pShellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")

	pGetModuleHandle      = kernel32.NewProc("GetModuleHandleW")
	pCreateMutex          = kernel32.NewProc("CreateMutexW")
	pGetCurrentProcess    = kernel32.NewProc("GetCurrentProcess")
	pSetProcessWorkingSet = kernel32.NewProc("SetProcessWorkingSetSize")
	pCreateFile           = kernel32.NewProc("CreateFileW")
	pDeviceIoControl      = kernel32.NewProc("DeviceIoControl")
	pCloseHandle          = kernel32.NewProc("CloseHandle")
	pGetSystemPowerStatus = kernel32.NewProc("GetSystemPowerStatus")

	pRegOpenKeyEx    = advapi32.NewProc("RegOpenKeyExW")
	pRegCreateKeyEx  = advapi32.NewProc("RegCreateKeyExW")
	pRegCloseKey     = advapi32.NewProc("RegCloseKey")
	pRegQueryValueEx = advapi32.NewProc("RegQueryValueExW")
	pRegSetValueEx   = advapi32.NewProc("RegSetValueExW")
	pRegDeleteValue  = advapi32.NewProc("RegDeleteValueW")
	pRegEnumKeyEx    = advapi32.NewProc("RegEnumKeyExW")

	pGetNumPhysMonitors = dxva2.NewProc("GetNumberOfPhysicalMonitorsFromHMONITOR")
	pGetPhysMonitors    = dxva2.NewProc("GetPhysicalMonitorsFromHMONITOR")
	pDestroyPhysMonitor = dxva2.NewProc("DestroyPhysicalMonitor")
	pGetVCPFeature      = dxva2.NewProc("GetVCPFeatureAndVCPFeatureReply")
	pSetVCPFeature      = dxva2.NewProc("SetVCPFeature")

	pDwmSetWindowAttribute = dwmapi.NewProc("DwmSetWindowAttribute")
	pGetDpiForMonitor      = shcore.NewProc("GetDpiForMonitor")
)

const (
	wmDestroy       = 0x0002
	wmActivate      = 0x0006
	wmPaint         = 0x000F
	wmEraseBkgnd    = 0x0014
	wmDisplayChange = 0x007E
	wmKeyDown       = 0x0100
	wmMouseMove     = 0x0200
	wmLButtonDown   = 0x0201
	wmLButtonUp     = 0x0202
	wmRButtonUp     = 0x0205
	wmApp           = 0x8000
	wmClose         = 0x0010
	wmTimer         = 0x0113
	wmPowerBroadcast = 0x0218
	wmLButtonDblClk = 0x0203

	wsPopup        = 0x80000000
	wsExToolWindow = 0x00000080
	wsExTopmost    = 0x00000008
	csDropShadow   = 0x00020000
	swShow         = 5
	vkEscape       = 0x1B
	mkLButton      = 0x0001
	srcCopy        = 0x00CC0020
	halftone       = 4
	transparentBk  = 1
	nullPen        = 8
	idcArrow       = 32512
	idiApplication = 32512
	imageIcon      = 1
	smCxSmIcon     = 49
	smCySmIcon     = 50

	mfString    = 0x0000
	mfChecked   = 0x0008
	mfSeparator = 0x0800

	tpmRightButton = 0x0002
	tpmReturnCmd   = 0x0100
	tpmNoNotify    = 0x0080
	tpmBottomAlign = 0x0020
	tpmRightAlign  = 0x0008

	dtLeft        = 0x0000
	dtRight       = 0x0002
	dtVCenter     = 0x0004
	dtSingleLine  = 0x0020
	dtNoPrefix    = 0x0800
	dtEndEllipsis = 0x8000

	nimAdd     = 0
	nimDelete  = 2
	nifMessage = 0x1
	nifIcon    = 0x2
	nifTip     = 0x4
)

type point struct{ x, y int32 }

type rect struct{ left, top, right, bottom int32 }

type msg struct {
	hwnd     uintptr
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     uint32
	pt       point
	lPrivate uint32
}

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type paintStruct struct {
	hdc         uintptr
	fErase      int32
	rcPaint     rect
	fRestore    int32
	fIncUpdate  int32
	rgbReserved [32]byte
}

type monitorInfoEx struct {
	cbSize    uint32
	rcMonitor rect
	rcWork    rect
	dwFlags   uint32
	szDevice  [32]uint16
}

type displayDevice struct {
	cb           uint32
	deviceName   [32]uint16
	deviceString [128]uint16
	stateFlags   uint32
	deviceID     [128]uint16
	deviceKey    [128]uint16
}

type physicalMonitor struct {
	handle uintptr
	desc   [128]uint16
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         [16]byte
	hBalloonIcon     uintptr
}

type systemPowerStatus struct {
	acLineStatus        byte
	batteryFlag         byte
	batteryLifePercent  byte
	systemStatusFlag    byte
	batteryLifeTime     uint32
	batteryFullLifeTime uint32
}

func u16(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func utf16ToString(b []uint16) string {
	return syscall.UTF16ToString(b)
}

func b2u(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

// ---- user32 ----

func registerClassEx(wc *wndClassEx) bool {
	r, _, _ := syscall.SyscallN(pRegisterClassEx.Addr(), uintptr(unsafe.Pointer(wc)))
	return r != 0
}

func createWindowEx(exStyle uint32, class, title *uint16, style uint32, x, y, w, h int32, parent, menu, inst, param uintptr) uintptr {
	r, _, _ := syscall.SyscallN(pCreateWindowEx.Addr(), uintptr(exStyle), uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title)),
		uintptr(style), uintptr(x), uintptr(y), uintptr(w), uintptr(h), parent, menu, inst, param)
	return r
}

func defWindowProc(hwnd, m, wp, lp uintptr) uintptr {
	r, _, _ := syscall.SyscallN(pDefWindowProc.Addr(), hwnd, m, wp, lp)
	return r
}

func destroyWindow(hwnd uintptr) { syscall.SyscallN(pDestroyWindow.Addr(), hwnd) }

func showWindow(hwnd uintptr, cmd int) { syscall.SyscallN(pShowWindow.Addr(), hwnd, uintptr(cmd)) }

func setForegroundWindow(hwnd uintptr) { syscall.SyscallN(pSetForegroundWindow.Addr(), hwnd) }

func getMessage(m *msg) int32 {
	r, _, _ := syscall.SyscallN(pGetMessage.Addr(), uintptr(unsafe.Pointer(m)), 0, 0, 0)
	return int32(r)
}

func translateMessage(m *msg) { syscall.SyscallN(pTranslateMessage.Addr(), uintptr(unsafe.Pointer(m))) }

func dispatchMessage(m *msg) { syscall.SyscallN(pDispatchMessage.Addr(), uintptr(unsafe.Pointer(m))) }

func postQuitMessage(code int) { syscall.SyscallN(pPostQuitMessage.Addr(), uintptr(code)) }

func postMessage(hwnd uintptr, m uint32, wp, lp uintptr) {
	syscall.SyscallN(pPostMessage.Addr(), hwnd, uintptr(m), wp, lp)
}

func getCursorPos(p *point) { syscall.SyscallN(pGetCursorPos.Addr(), uintptr(unsafe.Pointer(p))) }

func setCapture(hwnd uintptr) { syscall.SyscallN(pSetCapture.Addr(), hwnd) }

func releaseCapture() { syscall.SyscallN(pReleaseCapture.Addr()) }

func loadCursor(id uintptr) uintptr {
	r, _, _ := syscall.SyscallN(pLoadCursor.Addr(), 0, id)
	return r
}

func loadImage(inst, name uintptr, typ uint32, cx, cy int32, flags uint32) uintptr {
	r, _, _ := syscall.SyscallN(pLoadImage.Addr(), inst, name, uintptr(typ), uintptr(cx), uintptr(cy), uintptr(flags))
	return r
}

func invalidateRect(hwnd uintptr, r *rect) {
	syscall.SyscallN(pInvalidateRect.Addr(), hwnd, uintptr(unsafe.Pointer(r)), 0)
}

func setTimer(hwnd, id uintptr, ms uint32) {
	syscall.SyscallN(pSetTimer.Addr(), hwnd, id, uintptr(ms), 0)
}

func killTimer(hwnd, id uintptr) { syscall.SyscallN(pKillTimer.Addr(), hwnd, id) }

func getForegroundWindow() uintptr {
	r, _, _ := syscall.SyscallN(pGetForegroundWindow.Addr())
	return r
}

func getWindowThreadID(hwnd uintptr) uint32 {
	r, _, _ := syscall.SyscallN(pGetWindowThreadPID.Addr(), hwnd, 0)
	return uint32(r)
}

func getCurrentThreadID() uint32 {
	r, _, _ := syscall.SyscallN(pGetCurrentThreadID.Addr())
	return uint32(r)
}

func attachThreadInput(from, to uint32, attach bool) bool {
	r, _, _ := syscall.SyscallN(pAttachThreadInput.Addr(), uintptr(from), uintptr(to), b2u(attach))
	return r != 0
}

// bringToForeground activates hwnd. A plain SetForegroundWindow is refused when another
// process owns the foreground and we did not just receive input, so retry once while briefly
// sharing the foreground thread's input queue.
func bringToForeground(hwnd uintptr) bool {
	setForegroundWindow(hwnd)
	if getForegroundWindow() == hwnd {
		return true
	}
	fg := getForegroundWindow()
	if fg == 0 {
		return false
	}
	me, other := getCurrentThreadID(), getWindowThreadID(fg)
	if other == 0 || other == me {
		return false
	}
	attachThreadInput(me, other, true)
	setForegroundWindow(hwnd)
	attachThreadInput(me, other, false)
	return getForegroundWindow() == hwnd
}

func beginPaint(hwnd uintptr, ps *paintStruct) uintptr {
	r, _, _ := syscall.SyscallN(pBeginPaint.Addr(), hwnd, uintptr(unsafe.Pointer(ps)))
	return r
}

func endPaint(hwnd uintptr, ps *paintStruct) {
	syscall.SyscallN(pEndPaint.Addr(), hwnd, uintptr(unsafe.Pointer(ps)))
}

func fillRect(hdc uintptr, r *rect, brush uintptr) {
	syscall.SyscallN(pFillRect.Addr(), hdc, uintptr(unsafe.Pointer(r)), brush)
}

func drawText(hdc uintptr, text *uint16, r *rect, format uint32) {
	syscall.SyscallN(pDrawText.Addr(), hdc, uintptr(unsafe.Pointer(text)), ^uintptr(0), uintptr(unsafe.Pointer(r)), uintptr(format))
}

func createPopupMenu() uintptr {
	r, _, _ := syscall.SyscallN(pCreatePopupMenu.Addr())
	return r
}

func appendMenu(menu uintptr, flags uint32, id uintptr, text *uint16) {
	syscall.SyscallN(pAppendMenu.Addr(), menu, uintptr(flags), id, uintptr(unsafe.Pointer(text)))
}

func trackPopupMenu(menu uintptr, flags uint32, x, y int32, hwnd uintptr) uintptr {
	r, _, _ := syscall.SyscallN(pTrackPopupMenu.Addr(), menu, uintptr(flags), uintptr(x), uintptr(y), 0, hwnd, 0)
	return r
}

func destroyMenu(menu uintptr) { syscall.SyscallN(pDestroyMenu.Addr(), menu) }

func registerWindowMessage(name *uint16) uint32 {
	r, _, _ := syscall.SyscallN(pRegisterWindowMsg.Addr(), uintptr(unsafe.Pointer(name)))
	return uint32(r)
}

func monitorFromPoint(pt point, flags uint32) uintptr {
	packed := uintptr(uint32(pt.x)) | uintptr(uint32(pt.y))<<32
	r, _, _ := syscall.SyscallN(pMonitorFromPoint.Addr(), packed, uintptr(flags))
	return r
}

func getMonitorInfo(h uintptr, mi *monitorInfoEx) bool {
	r, _, _ := syscall.SyscallN(pGetMonitorInfo.Addr(), h, uintptr(unsafe.Pointer(mi)))
	return r != 0
}

func enumDisplayMonitors(cb uintptr) {
	syscall.SyscallN(pEnumDisplayMonitors.Addr(), 0, 0, cb, 0)
}

func enumDisplayDevices(dev *uint16, idx uint32, dd *displayDevice) bool {
	r, _, _ := syscall.SyscallN(pEnumDisplayDevices.Addr(), uintptr(unsafe.Pointer(dev)), uintptr(idx), uintptr(unsafe.Pointer(dd)), 0)
	return r != 0
}

func getSystemMetrics(i int) int32 {
	r, _, _ := syscall.SyscallN(pGetSystemMetrics.Addr(), uintptr(i))
	return int32(r)
}

func setProcessDpiAwarenessContext(ctx uintptr) { syscall.SyscallN(pSetDpiAwarenessCtx.Addr(), ctx) }

// ---- gdi32 ----

func createCompatibleDC(hdc uintptr) uintptr {
	r, _, _ := syscall.SyscallN(pCreateCompatibleDC.Addr(), hdc)
	return r
}

func createCompatibleBitmap(hdc uintptr, w, h int32) uintptr {
	r, _, _ := syscall.SyscallN(pCreateCompatibleBitmap.Addr(), hdc, uintptr(w), uintptr(h))
	return r
}

func selectObject(dc, obj uintptr) uintptr {
	r, _, _ := syscall.SyscallN(pSelectObject.Addr(), dc, obj)
	return r
}

func deleteObject(obj uintptr) { syscall.SyscallN(pDeleteObject.Addr(), obj) }

func deleteDC(dc uintptr) { syscall.SyscallN(pDeleteDC.Addr(), dc) }

func bitBlt(dst uintptr, x, y, w, h int32, src uintptr, sx, sy int32, rop uint32) {
	syscall.SyscallN(pBitBlt.Addr(), dst, uintptr(x), uintptr(y), uintptr(w), uintptr(h), src, uintptr(sx), uintptr(sy), uintptr(rop))
}

func stretchBlt(dst uintptr, x, y, w, h int32, src uintptr, sx, sy, sw, sh int32, rop uint32) {
	syscall.SyscallN(pStretchBlt.Addr(), dst, uintptr(x), uintptr(y), uintptr(w), uintptr(h), src,
		uintptr(sx), uintptr(sy), uintptr(sw), uintptr(sh), uintptr(rop))
}

func setStretchBltMode(dc uintptr, mode int) {
	syscall.SyscallN(pSetStretchBltMode.Addr(), dc, uintptr(mode))
}

func setBrushOrgEx(dc uintptr, x, y int32) {
	syscall.SyscallN(pSetBrushOrgEx.Addr(), dc, uintptr(x), uintptr(y), 0)
}

func createSolidBrush(color uint32) uintptr {
	r, _, _ := syscall.SyscallN(pCreateSolidBrush.Addr(), uintptr(color))
	return r
}

func getStockObject(i int) uintptr {
	r, _, _ := syscall.SyscallN(pGetStockObject.Addr(), uintptr(i))
	return r
}

func roundRect(dc uintptr, l, t, r, b, w, h int32) {
	syscall.SyscallN(pRoundRect.Addr(), dc, uintptr(l), uintptr(t), uintptr(r), uintptr(b), uintptr(w), uintptr(h))
}

func ellipse(dc uintptr, l, t, r, b int32) {
	syscall.SyscallN(pEllipse.Addr(), dc, uintptr(l), uintptr(t), uintptr(r), uintptr(b))
}

func createFont(height int32, weight int32, face *uint16) uintptr {
	const (
		defaultCharset   = 1
		cleartypeQuality = 5
	)
	r, _, _ := syscall.SyscallN(pCreateFont.Addr(), uintptr(height), 0, 0, 0, uintptr(weight), 0, 0, 0,
		defaultCharset, 0, 0, cleartypeQuality, 0, uintptr(unsafe.Pointer(face)))
	return r
}

func setBkMode(dc uintptr, mode int) { syscall.SyscallN(pSetBkMode.Addr(), dc, uintptr(mode)) }

func setTextColor(dc uintptr, color uint32) {
	syscall.SyscallN(pSetTextColor.Addr(), dc, uintptr(color))
}

// ---- shell32 / dwmapi / shcore ----

func shellNotifyIcon(message uint32, d *notifyIconData) bool {
	r, _, _ := syscall.SyscallN(pShellNotifyIcon.Addr(), uintptr(message), uintptr(unsafe.Pointer(d)))
	return r != 0
}

func dwmSetWindowAttribute(hwnd uintptr, attr uint32, v *uint32) {
	syscall.SyscallN(pDwmSetWindowAttribute.Addr(), hwnd, uintptr(attr), uintptr(unsafe.Pointer(v)), 4)
}

func getDpiForMonitor(h uintptr, dx, dy *uint32) bool {
	r, _, _ := syscall.SyscallN(pGetDpiForMonitor.Addr(), h, 0, uintptr(unsafe.Pointer(dx)), uintptr(unsafe.Pointer(dy)))
	return r == 0
}

// ---- kernel32 ----

func getModuleHandle() uintptr {
	r, _, _ := syscall.SyscallN(pGetModuleHandle.Addr(), 0)
	return r
}

// createMutex reports whether another instance already owns the named mutex.
func createMutex(name *uint16) (alreadyExists bool) {
	_, _, e := syscall.SyscallN(pCreateMutex.Addr(), 0, 0, uintptr(unsafe.Pointer(name)))
	return e == syscall.Errno(183)
}

func getCurrentProcess() uintptr {
	r, _, _ := syscall.SyscallN(pGetCurrentProcess.Addr())
	return r
}

func trimWorkingSet() {
	syscall.SyscallN(pSetProcessWorkingSet.Addr(), getCurrentProcess(), ^uintptr(0), ^uintptr(0))
}

func createFile(path *uint16) uintptr {
	const (
		genericReadWrite = 0xC0000000
		shareReadWrite   = 0x3
		openExisting     = 3
	)
	r, _, _ := syscall.SyscallN(pCreateFile.Addr(), uintptr(unsafe.Pointer(path)), genericReadWrite, shareReadWrite, 0, openExisting, 0, 0)
	return r
}

func deviceIoControl(h uintptr, code uint32, in unsafe.Pointer, inSize uint32, out unsafe.Pointer, outSize uint32, returned *uint32) bool {
	r, _, _ := syscall.SyscallN(pDeviceIoControl.Addr(), h, uintptr(code), uintptr(in), uintptr(inSize), uintptr(out), uintptr(outSize),
		uintptr(unsafe.Pointer(returned)), 0)
	return r != 0
}

func closeHandle(h uintptr) { syscall.SyscallN(pCloseHandle.Addr(), h) }

func getSystemPowerStatus(s *systemPowerStatus) bool {
	r, _, _ := syscall.SyscallN(pGetSystemPowerStatus.Addr(), uintptr(unsafe.Pointer(s)))
	return r != 0
}

// ---- dxva2 ----

func getNumPhysicalMonitors(h uintptr, n *uint32) bool {
	r, _, _ := syscall.SyscallN(pGetNumPhysMonitors.Addr(), h, uintptr(unsafe.Pointer(n)))
	return r != 0
}

func getPhysicalMonitors(h uintptr, n uint32, arr *physicalMonitor) bool {
	r, _, _ := syscall.SyscallN(pGetPhysMonitors.Addr(), h, uintptr(n), uintptr(unsafe.Pointer(arr)))
	return r != 0
}

func destroyPhysicalMonitor(h uintptr) { syscall.SyscallN(pDestroyPhysMonitor.Addr(), h) }

func getVCPFeature(h uintptr, code byte, cur, max *uint32) bool {
	r, _, _ := syscall.SyscallN(pGetVCPFeature.Addr(), h, uintptr(code), 0, uintptr(unsafe.Pointer(cur)), uintptr(unsafe.Pointer(max)))
	return r != 0
}

func setVCPFeature(h uintptr, code byte, value uint32) bool {
	r, _, _ := syscall.SyscallN(pSetVCPFeature.Addr(), h, uintptr(code), uintptr(value))
	return r != 0
}
