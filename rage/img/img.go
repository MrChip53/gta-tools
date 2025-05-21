package img

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/mrchip53/gta-tools/rage/util"
)

const (
	HEADER_MAGIC_BYTES = 0xA94E2A52
)

type ImgEntry struct {
	idx  int
	name string
	toc  TocEntry
	data []byte
}

func (e ImgEntry) Name() string { return e.name }
func (e ImgEntry) Data() []byte {
	d := make([]byte, len(e.data))
	copy(d, e.data)
	return d
}
func (e ImgEntry) Toc() TocEntry { return e.toc }

type TocEntry struct {
	Size           int
	RscFlags       int
	RsourceType    int
	OffsetBlock    int
	UsedBlocks     int
	Flags          int
	IsResourceFile bool
	rawData        []byte
}

type ImgHeader struct {
	Identifier   uint32
	Version      int32
	EntryCount   int32
	TocSize      int32
	TocEntrySize int16
	Uknown1      int16
	rawData      []byte
}

type ImgFile struct {
	header    ImgHeader
	entries   []ImgEntry
	encrypted bool
	rawToc    []byte
}

func (f ImgFile) Entries() []ImgEntry { return f.entries }
func (f ImgFile) Bytes() []byte {
	var header []byte
	var entryNames string
	var tocEntries []byte
	var data []byte

	header = f.header.write()
	err := util.Encrypt(header)
	if err != nil {
		panic(err)
	}

	var names []string
	for _, e := range f.entries {
		tmp := make([]byte, e.toc.UsedBlocks*0x800)
		copy(tmp, e.Data())
		names = append(names, e.Name())
		tocEntries = append(tocEntries, e.toc.rawData...)
		data = append(data, tmp...)
	}
	entryNames = strings.Join(names, "\x00") + "\x00"
	tocEntries = append(tocEntries, []byte(entryNames)...)
	err = util.Encrypt(tocEntries)
	if err != nil {
		panic(err)
	}

	metadata := make([]byte, 0x800*f.entries[0].toc.OffsetBlock)
	copy(metadata, header)
	copy(metadata[len(header):], tocEntries)

	final := append(metadata, data...)
	return final
}

func ParseImgHeader(data []byte) ImgHeader {
	h := ImgHeader{}
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
	h.Uknown1 = int16(binary.LittleEndian.Uint16(d[18:20]))
}

func (h *ImgHeader) write() []byte {
	header := make([]byte, 20)
	binary.LittleEndian.PutUint32(header[0:4], h.Identifier)
	binary.LittleEndian.PutUint32(header[4:8], uint32(h.Version))
	binary.LittleEndian.PutUint32(header[8:12], uint32(h.EntryCount))
	binary.LittleEndian.PutUint32(header[12:16], uint32(h.TocSize))
	binary.LittleEndian.PutUint16(header[16:18], uint16(h.TocEntrySize))
	binary.LittleEndian.PutUint16(header[18:20], uint16(h.Uknown1))
	return header
}

func NewTocEntry(data []byte) TocEntry {
	t := TocEntry{}
	t.rawData = make([]byte, len(data))
	copy(t.rawData, data)
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
		t.Size = t.UsedBlocks*0x800 - (t.Flags & 0x7FF)
	}
	return t
}

func LoadImgFile(data []byte) ImgFile {
	var err error
	encrypted := false
	magicBytes := binary.BigEndian.Uint32(data[0:4])
	if magicBytes != HEADER_MAGIC_BYTES {
		encrypted = true
	}

	reader := bytes.NewReader(data)

	headerBytes := make([]byte, 20)
	_, err = reader.Read(headerBytes)
	if err != nil {
		panic(err)
	}
	rawHeader := make([]byte, 20)
	copy(rawHeader, headerBytes)

	if encrypted {
		err = util.Decrypt(headerBytes)
		if err != nil {
			panic(err)
		}
	}

	header := ParseImgHeader(headerBytes)
	header.rawData = rawHeader

	tocBytes := make([]byte, header.TocSize)
	_, err = reader.Read(tocBytes)
	if err != nil {
		panic(err)
	}

	rawTocBytes := make([]byte, header.TocSize)
	copy(rawTocBytes, tocBytes)

	if encrypted {
		err = util.Decrypt(tocBytes)
		if err != nil {
			panic(err)
		}
	}

	stringData := tocBytes[header.EntryCount*int32(header.TocEntrySize):]
	entryNames := strings.Split(string(stringData), "\x00")

	var entries []ImgEntry

	for i := range int(header.EntryCount) {
		eb := tocBytes[i*int(header.TocEntrySize) : (i+1)*int(header.TocEntrySize)]
		e := NewTocEntry(eb)
		d := data[e.OffsetBlock*0x800 : e.OffsetBlock*0x800+e.Size]
		entries = append(entries, ImgEntry{
			idx:  i,
			name: entryNames[i],
			data: d,
			toc:  e,
		})
	}

	return ImgFile{
		header:    header,
		entries:   entries,
		encrypted: encrypted,
		rawToc:    rawTocBytes,
	}
}
