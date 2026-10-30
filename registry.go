package main

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	hkeyCurrentUser  = 0x80000001
	hkeyLocalMachine = 0x80000002
	keyRead          = 0x20019
	keySetValue      = 0x0002
	regSZ            = 1

	runKeyPath   = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValueName = "BrightSlide"
)

func regOpen(root uintptr, path string, access uint32) (uintptr, bool) {
	var h uintptr
	r, _, _ := syscall.SyscallN(pRegOpenKeyEx.Addr(), root, uintptr(unsafe.Pointer(u16(path))), 0, uintptr(access), uintptr(unsafe.Pointer(&h)))
	return h, r == 0
}

func regClose(h uintptr) { syscall.SyscallN(pRegCloseKey.Addr(), h) }

func regReadValue(h uintptr, name string) ([]byte, uint32, bool) {
	var typ, size uint32
	n := u16(name)
	r, _, _ := syscall.SyscallN(pRegQueryValueEx.Addr(), h, uintptr(unsafe.Pointer(n)), 0, uintptr(unsafe.Pointer(&typ)), 0, uintptr(unsafe.Pointer(&size)))
	if r != 0 || size == 0 {
		return nil, 0, false
	}
	buf := make([]byte, size)
	r, _, _ = syscall.SyscallN(pRegQueryValueEx.Addr(), h, uintptr(unsafe.Pointer(n)), 0, uintptr(unsafe.Pointer(&typ)), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if r != 0 {
		return nil, 0, false
	}
	return buf[:size], typ, true
}

func regEnumKeys(h uintptr) []string {
	var out []string
	for i := uint32(0); ; i++ {
		buf := make([]uint16, 256)
		n := uint32(len(buf))
		r, _, _ := syscall.SyscallN(pRegEnumKeyEx.Addr(), h, uintptr(i), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n)), 0, 0, 0, 0)
		if r != 0 {
			return out
		}
		out = append(out, syscall.UTF16ToString(buf[:n]))
	}
}

func exePath() string {
	p, err := os.Executable()
	if err != nil {
		return ""
	}
	return p
}

// autostartEnabled accepts the Run value with or without surrounding quotes, so values written
// by older builds (unquoted) and by this one (quoted, safe for paths with spaces) both count.
func autostartEnabled() bool {
	h, ok := regOpen(hkeyCurrentUser, runKeyPath, keyRead)
	if !ok {
		return false
	}
	defer regClose(h)
	data, typ, ok := regReadValue(h, runValueName)
	if !ok || typ != regSZ || len(data) < 2 {
		return false
	}
	u := unsafe.Slice((*uint16)(unsafe.Pointer(&data[0])), len(data)/2)
	return strings.EqualFold(strings.Trim(syscall.UTF16ToString(u), `"`), exePath())
}

func setAutostart(enable bool) {
	var h uintptr
	if enable {
		var disp uint32
		r, _, _ := syscall.SyscallN(pRegCreateKeyEx.Addr(), hkeyCurrentUser, uintptr(unsafe.Pointer(u16(runKeyPath))), 0, 0, 0,
			uintptr(keySetValue), 0, uintptr(unsafe.Pointer(&h)), uintptr(unsafe.Pointer(&disp)))
		if r != 0 {
			return
		}
		defer regClose(h)
		val, _ := syscall.UTF16FromString(`"` + exePath() + `"`)
		syscall.SyscallN(pRegSetValueEx.Addr(), h, uintptr(unsafe.Pointer(u16(runValueName))), 0, regSZ,
			uintptr(unsafe.Pointer(&val[0])), uintptr(len(val)*2))
		return
	}
	h, ok := regOpen(hkeyCurrentUser, runKeyPath, keySetValue)
	if !ok {
		return
	}
	defer regClose(h)
	syscall.SyscallN(pRegDeleteValue.Addr(), h, uintptr(unsafe.Pointer(u16(runValueName))))
}
