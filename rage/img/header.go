package img

import (
	"encoding/binary"
)

const (
	HEADER_MAGIC_BYTES = 0xA94E2A52
)

type ImgHeader struct {
	Identifier   uint32
	Version      int32
	EntryCount   int32
	TocSize      int32
	TocEntrySize int16
	Unknown1     int16
	rawData      []byte
}

func ParseImgHeader(data []byte) *ImgHeader {
	h := &ImgHeader{}
	h.read(data)
	return h
}

func (h *ImgHeader) read(d []byte) {
	if len(d) < 20 {
		panic("Not enough data to read header")
	}
	h.Identifier = binary.LittleEndian.Uint32(d[0:4])
	h.Version = int32(binary.LittleEndian.Uint32(d[4:8]))
	h.EntryCount = int32(binary.LittleEndian.Uint32(d[8:12]))
	h.TocSize = int32(binary.LittleEndian.Uint32(d[12:16]))
	h.TocEntrySize = int16(binary.LittleEndian.Uint16(d[16:18]))
	h.Unknown1 = int16(binary.LittleEndian.Uint16(d[18:20]))
}

func (h *ImgHeader) write() []byte {
	header := make([]byte, 20)
	binary.LittleEndian.PutUint32(header[0:4], h.Identifier)
	binary.LittleEndian.PutUint32(header[4:8], uint32(h.Version))
	binary.LittleEndian.PutUint32(header[8:12], uint32(h.EntryCount))
	binary.LittleEndian.PutUint32(header[12:16], uint32(h.TocSize))
	binary.LittleEndian.PutUint16(header[16:18], uint16(h.TocEntrySize))
	binary.LittleEndian.PutUint16(header[18:20], uint16(h.Unknown1))
	return header
}
