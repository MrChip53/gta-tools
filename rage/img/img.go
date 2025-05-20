package img

import "encoding/binary"

type TocEntry struct {
	Size        int
	RscFlags    int
	RsourceType int
	OffsetBlock int
	UsedBlocks  int
	Flags       int
}

func NewTocEntry(data []byte) TocEntry {
	t := TocEntry{}
	temp := binary.LittleEndian.Uint32(data[0:4])
	isResourceFile := (temp & 0xc0000000) != 0
	if !isResourceFile {
		t.Size = int(temp)
	} else {
		t.RscFlags = int(temp)
	}

	t.RsourceType = int(binary.LittleEndian.Uint32(data[4:8]))
	t.OffsetBlock = int(binary.LittleEndian.Uint32(data[8:12]))
	t.UsedBlocks = int(binary.LittleEndian.Uint16(data[12:14]))
	t.Flags = int(binary.LittleEndian.Uint16(data[14:16]))

	if isResourceFile {
		t.Size = t.UsedBlocks*0x800 - (t.Flags & 0x7FF)
	}
	return t
}
