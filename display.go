package main

import (
	"fmt"
	"math"
	"strings"
	"syscall"
	"unsafe"
)

const vcpBrightness = 0x10

type controller interface {
	Get() (int, bool)
	Set(percent int) bool
	// Effective is the value Set will actually apply (some panels remap low values).
	Effective(percent int) int
	Close()
}

type display struct {
	name     string
	external bool
	ctl      controller
	value    int // slider position; touched only by the UI thread
	wr       *writer
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ---- DDC/CI (external monitors) ----

type ddcController struct {
	handle uintptr
	max    uint32
}

func (c *ddcController) Get() (int, bool) {
	var cur, max uint32
	if !getVCPFeature(c.handle, vcpBrightness, &cur, &max) {
		return 0, false
	}
	return int(math.Round(float64(cur) * 100 / float64(c.max))), true
}

// Some monitors render 1-8% darker than 0%, so that dead zone is skipped.
func (c *ddcController) Effective(p int) int {
	p = clamp(p, 0, 100)
	if p >= 1 && p <= 8 {
		return 9
	}
	return p
}

func (c *ddcController) Set(p int) bool {
	raw := uint32(math.Round(float64(c.Effective(p)) * float64(c.max) / 100))
	return setVCPFeature(c.handle, vcpBrightness, raw)
}

func (c *ddcController) Close() { destroyPhysicalMonitor(c.handle) }

var discovered []*display

var enumMonitorsCallback = syscall.NewCallback(func(hMonitor, hdc, lprc, data uintptr) uintptr {
	var count uint32
	if !getNumPhysicalMonitors(hMonitor, &count) || count == 0 {
		return 1
	}
	pms := make([]physicalMonitor, count)
	if !getPhysicalMonitors(hMonitor, count, &pms[0]) {
		return 1
	}
	model := resolveModelName(hMonitor)
	for _, pm := range pms {
		var cur, max uint32
		// Monitors (or cables) without DDC/CI fail this probe.
		if !getVCPFeature(pm.handle, vcpBrightness, &cur, &max) {
			destroyPhysicalMonitor(pm.handle)
			continue
		}
		if max == 0 {
			max = 100
		}
		desc := model
		if desc == "" {
			desc = strings.TrimSpace(utf16ToString(pm.desc[:]))
		}
		name := fmt.Sprintf("외부 모니터 %d", len(discovered)+1)
		if desc != "" {
			name += " (" + desc + ")"
		}
		discovered = append(discovered, &display{name: name, external: true, ctl: &ddcController{handle: pm.handle, max: max}})
	}
	return 1
})

func discoverExternal() []*display {
	discovered = nil
	enumDisplayMonitors(enumMonitorsCallback)
	out := discovered
	discovered = nil
	return out
}

// resolveModelName maps an HMONITOR to its EDID model name through the display
// device's hardware id ("MONITOR\SKG2712\{guid}\0000" -> "SKG2712").
func resolveModelName(hMonitor uintptr) string {
	var mi monitorInfoEx
	mi.cbSize = uint32(unsafe.Sizeof(mi))
	if !getMonitorInfo(hMonitor, &mi) {
		return ""
	}
	var dd displayDevice
	dd.cb = uint32(unsafe.Sizeof(dd))
	if !enumDisplayDevices(&mi.szDevice[0], 0, &dd) {
		return ""
	}
	parts := strings.Split(utf16ToString(dd.deviceID[:]), `\`)
	if len(parts) < 2 {
		return ""
	}
	return edidModelName(parts[1])
}

func edidModelName(hardwareID string) string {
	base := `SYSTEM\CurrentControlSet\Enum\DISPLAY\` + hardwareID
	h, ok := regOpen(hkeyLocalMachine, base, keyRead)
	if !ok {
		return ""
	}
	instances := regEnumKeys(h)
	regClose(h)
	for _, inst := range instances {
		ih, ok := regOpen(hkeyLocalMachine, base+`\`+inst+`\Device Parameters`, keyRead)
		if !ok {
			continue
		}
		edid, _, ok := regReadValue(ih, "EDID")
		regClose(ih)
		if !ok {
			continue
		}
		if name := parseEDIDName(edid); name != "" {
			return name
		}
	}
	return ""
}

func parseEDIDName(edid []byte) string {
	if len(edid) < 128 {
		return ""
	}
	for off := 54; off <= 108; off += 18 {
		d := edid[off : off+18]
		if d[0] == 0 && d[1] == 0 && d[2] == 0 && d[3] == 0xFC {
			return strings.TrimSpace(strings.SplitN(string(d[5:18]), "\n", 2)[0])
		}
	}
	mfg := uint16(edid[8])<<8 | uint16(edid[9])
	letters := []byte{byte('@' + (mfg>>10)&0x1F), byte('@' + (mfg>>5)&0x1F), byte('@' + mfg&0x1F)}
	product := uint16(edid[11])<<8 | uint16(edid[10])
	return fmt.Sprintf("%s %04X", letters, product)
}

func discoverAll() []*display {
	var out []*display
	out = append(out, discoverInternal()...)
	out = append(out, discoverExternal()...)
	for _, d := range out {
		if v, ok := d.ctl.Get(); ok {
			d.value = clamp(v, 0, 100)
		}
		d.wr = newWriter(d)
	}
	return out
}

func closeDisplays(ds []*display) {
	for _, d := range ds {
		d.wr.stop()
		d.ctl.Close()
	}
}
