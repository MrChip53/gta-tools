package img

import (
	"encoding/binary"
)

type TocEntry struct {
	Size           int
	RscFlags       int
	RsourceType    int
	OffsetBlock    int
	UsedBlocks     int
	Flags          int
	IsResourceFile bool

	entrySize int
}

func NewTocEntry(data []byte) TocEntry {
	t := TocEntry{entrySize: len(data)}
	temp := binary.LittleEndian.Uint32(data[0:4])
	t.IsResourceFile = (temp & 0xc0000000) != 0
	if !t.IsResourceFile {
		t.Size = int(temp)
	} else {
		t.RscFlags = int(temp)
	}

	t.RsourceType = int(binary.LittleEndian.Uint32(data[4:8]))
	t.OffsetBlock = int(binary.LittleEndian.Uint32(data[8:12]))
	t.UsedBlocks = int(binary.LittleEndian.Uint16(data[12:14]))
	t.Flags = int(binary.LittleEndian.Uint16(data[14:16]))

	if t.IsResourceFile {
		size := t.UsedBlocks*0x800 - (t.Flags & 0x7FF)
		if size < 0 {
			size = 0
		}
		t.Size = size
	}
	return t
}

func (t *TocEntry) Bytes() []byte {
	b := make([]byte, t.entrySize)
	var temp uint32
	if t.IsResourceFile {
		temp = uint32(t.RscFlags)
	} else {
		temp = uint32(t.Size)
	}
	binary.LittleEndian.PutUint32(b[0:4], temp)
	binary.LittleEndian.PutUint32(b[4:8], uint32(t.RsourceType))
	binary.LittleEndian.PutUint32(b[8:12], uint32(t.OffsetBlock))
	binary.LittleEndian.PutUint16(b[12:14], uint16(t.UsedBlocks))
	binary.LittleEndian.PutUint16(b[14:16], uint16(t.Flags))
	return b
}
