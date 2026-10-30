package main

import (
	"unsafe"
)

// Built-in laptop panels are driven through the video miniport brightness IOCTLs on
// \\.\LCD. This path is not verified on real laptop hardware.
const (
	ioctlQuerySupportedBrightness = 0x00230496
	ioctlQueryDisplayBrightness   = 0x00230498
	ioctlSetDisplayBrightness     = 0x0023049C
	displayPolicyBoth             = 3
)

type displayBrightness struct {
	policy byte
	ac     byte
	dc     byte
}

type lcdController struct {
	handle uintptr
	levels []byte
}

func discoverInternal() []*display {
	h := createFile(u16(`\\.\LCD`))
	if h == 0 || h == ^uintptr(0) {
		return nil
	}
	buf := make([]byte, 256)
	var n uint32
	if !deviceIoControl(h, ioctlQuerySupportedBrightness, nil, 0, unsafe.Pointer(&buf[0]), uint32(len(buf)), &n) || n == 0 {
		closeHandle(h)
		return nil
	}
	c := &lcdController{handle: h, levels: buf[:n]}
	if _, ok := c.Get(); !ok {
		closeHandle(h)
		return nil
	}
	return []*display{{name: "내장 디스플레이", ctl: c}}
}

func (c *lcdController) nearest(p int) int {
	best, bestDiff := int(c.levels[0]), 1<<30
	for _, l := range c.levels {
		d := int(l) - p
		if d < 0 {
			d = -d
		}
		if d < bestDiff {
			best, bestDiff = int(l), d
		}
	}
	return best
}

func (c *lcdController) Effective(p int) int { return c.nearest(clamp(p, 0, 100)) }

func (c *lcdController) Get() (int, bool) {
	var b displayBrightness
	var n uint32
	if !deviceIoControl(c.handle, ioctlQueryDisplayBrightness, nil, 0, unsafe.Pointer(&b), uint32(unsafe.Sizeof(b)), &n) {
		return 0, false
	}
	var ps systemPowerStatus
	if getSystemPowerStatus(&ps) && ps.acLineStatus != 1 {
		return int(b.dc), true
	}
	return int(b.ac), true
}

func (c *lcdController) Set(p int) bool {
	level := byte(c.Effective(p))
	b := displayBrightness{policy: displayPolicyBoth, ac: level, dc: level}
	var n uint32
	return deviceIoControl(c.handle, ioctlSetDisplayBrightness, unsafe.Pointer(&b), uint32(unsafe.Sizeof(b)), nil, 0, &n)
}

func (c *lcdController) Close() { closeHandle(c.handle) }
