package img

import (
	"bytes"
	"encoding/binary"
	"sort"
	"strings"

	"github.com/mrchip53/gta-tools/rage/util"
)

const (
	HEADER_SIZE = 20
	BLOCK_SIZE  = 0x800
)

type ImgFile struct {
	header    *ImgHeader
	entries   []*ImgEntry
	encrypted bool
}

func (f ImgFile) Entries() []*ImgEntry { return f.entries }

func (f *ImgFile) AddEntry(e *ImgEntry) {
	f.entries = append(f.entries, e)
	sort.Slice(f.entries, func(i, j int) bool {
		return f.entries[i].Name() < f.entries[j].Name()
	})
	for i, e := range f.entries {
		e.idx = i
	}
	f.rebuild()
}

func (f *ImgFile) rebuild() {
	entryCount := len(f.entries)
	f.header.EntryCount = int32(entryCount)

	tocSize := int(f.header.TocEntrySize) * entryCount

	entryNames := make([]string, entryCount)
	for i, e := range f.entries {
		entryNames[i] = e.Name()
	}
	entryNamesStr := strings.Join(entryNames, "\x00") + "\x00"
	entryNamesBytes := []byte(entryNamesStr)
	tocSize += len(entryNamesBytes)

	f.header.TocSize = int32(tocSize)
	curBlock := (f.header.TocSize + HEADER_SIZE) / BLOCK_SIZE
	if (f.header.TocSize+HEADER_SIZE)%BLOCK_SIZE != 0 {
		curBlock++
	}
	for _, e := range f.entries {
		e.toc.OffsetBlock = int(curBlock)
		e.toc.Size = len(e.data)
		dataBlocks := len(e.data) / BLOCK_SIZE
		if len(e.data)%BLOCK_SIZE != 0 {
			dataBlocks++
		}
		e.toc.UsedBlocks = dataBlocks
		curBlock += int32(dataBlocks)
	}
}

func (f ImgFile) Bytes() []byte {
	f.rebuild()

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
		tmp := make([]byte, e.toc.UsedBlocks*BLOCK_SIZE)
		copy(tmp, e.Data())
		names = append(names, e.Name())
		tocEntries = append(tocEntries, e.toc.Bytes()...)
		data = append(data, tmp...)
	}
	entryNames = strings.Join(names, "\x00") + "\x00"
	tocEntries = append(tocEntries, []byte(entryNames)...)
	err = util.Encrypt(tocEntries)
	if err != nil {
		panic(err)
	}

	var metadata []byte
	if len(f.entries) > 0 {
		metadata = make([]byte, BLOCK_SIZE*f.entries[0].toc.OffsetBlock)
	} else {
		// Handle case with no entries, e.g., create minimal metadata or panic
		// For now, let's assume a minimal size if no entries, or decide on a specific behavior.
		// This might need further refinement based on expected IMG structure for empty files.
		metadata = make([]byte, BLOCK_SIZE) // Default to one block size if no entries
	}

	if len(metadata) < len(header)+len(tocEntries) {
		// Ensure metadata is large enough, otherwise, this could panic
		// This scenario suggests a mismatch in calculated sizes or an issue with OffsetBlock
		// For now, we'll allocate a larger buffer if needed, but this should be reviewed for correctness
		tempMetadata := make([]byte, len(header)+len(tocEntries))
		copy(tempMetadata, metadata)
		metadata = tempMetadata
	}

	copy(metadata, header)
	copy(metadata[len(header):], tocEntries)

	final := append(metadata, data...)
	return final
}

func LoadImgFile(data []byte) ImgFile {
	var err error
	encrypted := false
	magicBytes := binary.BigEndian.Uint32(data[0:4])
	if magicBytes != HEADER_MAGIC_BYTES {
		encrypted = true
	}

	reader := bytes.NewReader(data)

	headerBytes := make([]byte, HEADER_SIZE)
	_, err = reader.Read(headerBytes)
	if err != nil {
		panic(err)
	}
	rawHeader := make([]byte, HEADER_SIZE)
	copy(rawHeader, headerBytes)

	if encrypted {
		err = util.Decrypt(headerBytes)
		if err != nil {
			panic(err)
		}
	}

	header := ParseImgHeader(headerBytes)
	header.rawData = rawHeader

	if header.TocSize < 0 {
		panic("Invalid TocSize in header")
	}
	tocBytes := make([]byte, header.TocSize)
	_, err = reader.Read(tocBytes)
	if err != nil {
		panic(err)
	}

	if encrypted {
		err = util.Decrypt(tocBytes)
		if err != nil {
			panic(err)
		}
	}

	// Ensure EntryCount and TocEntrySize are not negative and won't cause overflow or out-of-bounds access
	if header.EntryCount < 0 || header.TocEntrySize < 0 {
		panic("Invalid EntryCount or TocEntrySize in header")
	}
	entryDataSize := int(header.EntryCount) * int(header.TocEntrySize)
	if entryDataSize < 0 || entryDataSize > len(tocBytes) { // Check for overflow and bounds
		panic("Calculated entry data size is invalid or out of bounds")
	}

	stringData := tocBytes[entryDataSize:]
	entryNames := strings.Split(string(stringData), "\x00")

	var entries []*ImgEntry

	for i := range int(header.EntryCount) {
		tocEntryStartIndex := i * int(header.TocEntrySize)
		tocEntryEndIndex := (i + 1) * int(header.TocEntrySize)
		if tocEntryEndIndex > len(tocBytes) || tocEntryStartIndex < 0 { // Bounds check
			panic("TocEntry index out of bounds")
		}
		eb := tocBytes[tocEntryStartIndex:tocEntryEndIndex]
		e := NewTocEntry(eb)

		if e.OffsetBlock < 0 || e.Size < 0 { // Sanity check for offset and size
			panic("Invalid entry offset or size")
		}
		dataStartIndex := e.OffsetBlock * BLOCK_SIZE
		dataEndIndex := dataStartIndex + e.Size

		if dataStartIndex < 0 || dataEndIndex > len(data) || dataStartIndex > dataEndIndex { // More robust bounds check for data slice
			panic("Entry data indices out of bounds or invalid")
		}

		d := data[dataStartIndex:dataEndIndex]
		entries = append(entries, &ImgEntry{
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
	}
}
