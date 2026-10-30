// genrsrc turns Assets/brightslide.ico into rsrc_windows_amd64.syso so `go build` embeds the
// icon (group icon id 1) in the exe. Run from the repo root: go run ./tools/genrsrc
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	rtIcon      = 3
	rtGroupIcon = 14
	langEnglish = 0x0409
)

type image struct {
	width, height, colors, reserved byte
	planes, bitCount                uint16
	data                            []byte
}

func main() {
	ico, err := os.ReadFile("Assets/brightslide.ico")
	check(err)
	if len(ico) < 6 || binary.LittleEndian.Uint16(ico[0:]) != 0 || binary.LittleEndian.Uint16(ico[2:]) != 1 {
		fail("Assets/brightslide.ico is not an ICO file")
	}
	count := int(binary.LittleEndian.Uint16(ico[4:]))
	if count == 0 || 6+16*count > len(ico) {
		fail("ICO header is truncated or has no images")
	}
	var images []image
	for i := 0; i < count; i++ {
		e := ico[6+16*i:]
		size := binary.LittleEndian.Uint32(e[8:])
		offset := binary.LittleEndian.Uint32(e[12:])
		if uint64(offset)+uint64(size) > uint64(len(ico)) {
			fail(fmt.Sprintf("ICO image %d points outside the file", i))
		}
		images = append(images, image{e[0], e[1], e[2], e[3], binary.LittleEndian.Uint16(e[4:]), binary.LittleEndian.Uint16(e[6:]), ico[offset : offset+size]})
	}

	var group bytes.Buffer
	w := func(buf *bytes.Buffer, v any) { check(binary.Write(buf, binary.LittleEndian, v)) }
	w(&group, [3]uint16{0, 1, uint16(count)})
	for i, im := range images {
		group.Write([]byte{im.width, im.height, im.colors, im.reserved})
		w(&group, im.planes)
		w(&group, im.bitCount)
		w(&group, uint32(len(im.data)))
		w(&group, uint16(i+1))
	}

	// Blobs: icon images (ids 1..n), then the group icon (id 1).
	blobs := make([][]byte, 0, count+1)
	for _, im := range images {
		blobs = append(blobs, im.data)
	}
	blobs = append(blobs, group.Bytes())
	n := len(blobs) // one leaf per blob

	const dirSize, entrySize, dataEntrySize = 16, 8, 16
	rootOff := 0
	iconDirOff := rootOff + dirSize + 2*entrySize
	groupDirOff := iconDirOff + dirSize + count*entrySize
	langDirsOff := groupDirOff + dirSize + 1*entrySize
	dataEntriesOff := langDirsOff + n*(dirSize+entrySize)
	blobsOff := dataEntriesOff + n*dataEntrySize
	blobOffsets := make([]int, n)
	off := blobsOff
	for i, b := range blobs {
		blobOffsets[i] = off
		off += (len(b) + 3) &^ 3
	}
	total := off

	rsrc := make([]byte, total)
	put32 := func(at int, v uint32) { binary.LittleEndian.PutUint32(rsrc[at:], v) }
	put16 := func(at int, v uint16) { binary.LittleEndian.PutUint16(rsrc[at:], v) }
	dir := func(at, idEntries int) { put16(at+14, uint16(idEntries)) }
	entry := func(at int, id uint32, target int, isDir bool) {
		put32(at, id)
		if isDir {
			target |= 0x80000000
		}
		put32(at+4, uint32(target))
	}

	dir(rootOff, 2)
	entry(rootOff+dirSize, rtIcon, iconDirOff, true)
	entry(rootOff+dirSize+entrySize, rtGroupIcon, groupDirOff, true)
	dir(iconDirOff, count)
	for i := 0; i < count; i++ {
		entry(iconDirOff+dirSize+i*entrySize, uint32(i+1), langDirsOff+i*(dirSize+entrySize), true)
	}
	dir(groupDirOff, 1)
	entry(groupDirOff+dirSize, 1, langDirsOff+count*(dirSize+entrySize), true)

	var relocs []int
	for i := 0; i < n; i++ {
		ld := langDirsOff + i*(dirSize+entrySize)
		de := dataEntriesOff + i*dataEntrySize
		dir(ld, 1)
		entry(ld+dirSize, langEnglish, de, false)
		put32(de, uint32(blobOffsets[i])) // RVA, fixed up by the relocation below
		put32(de+4, uint32(len(blobs[i])))
		copy(rsrc[blobOffsets[i]:], blobs[i])
		relocs = append(relocs, de)
	}

	// COFF object: header, one .rsrc section, relocations, symbol table (section symbol + aux).
	const headerSize, sectionHdrSize = 20, 40
	rawOff := headerSize + sectionHdrSize
	relocOff := rawOff + len(rsrc)
	symOff := relocOff + len(relocs)*10

	var out bytes.Buffer
	w(&out, uint16(0x8664))
	w(&out, uint16(1))
	w(&out, uint32(0))
	w(&out, uint32(symOff))
	w(&out, uint32(2))
	w(&out, uint16(0))
	w(&out, uint16(0))

	out.Write([]byte(".rsrc\x00\x00\x00"))
	w(&out, uint32(0))
	w(&out, uint32(0))
	w(&out, uint32(len(rsrc)))
	w(&out, uint32(rawOff))
	w(&out, uint32(relocOff))
	w(&out, uint32(0))
	w(&out, uint16(len(relocs)))
	w(&out, uint16(0))
	w(&out, uint32(0xC0300040))

	out.Write(rsrc)
	for _, r := range relocs {
		w(&out, uint32(r))
		w(&out, uint32(0))
		w(&out, uint16(3)) // IMAGE_REL_AMD64_ADDR32NB
	}

	out.Write([]byte(".rsrc\x00\x00\x00"))
	w(&out, uint32(0))
	w(&out, uint16(1))
	w(&out, uint16(0))
	out.WriteByte(3) // static
	out.WriteByte(1) // one aux record
	w(&out, uint32(len(rsrc)))
	w(&out, uint16(len(relocs)))
	w(&out, uint16(0))
	w(&out, uint32(0))
	w(&out, uint16(1))
	out.Write([]byte{0, 0, 0, 0})
	w(&out, uint32(4)) // empty string table

	check(os.WriteFile("rsrc_windows_amd64.syso", out.Bytes(), 0o644))
	fmt.Printf("wrote rsrc_windows_amd64.syso (%d icon images, %d bytes)\n", count, out.Len())
}

func check(err error) {
	if err != nil {
		fail(err.Error())
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "genrsrc:", msg+" (run from the repo root)")
	os.Exit(1)
}
