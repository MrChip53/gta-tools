package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mrchip53/gta-tools/models"
	"github.com/mrchip53/gta-tools/rage/img"
	"github.com/mrchip53/gta-tools/rage/util"
)

const (
	HEADER_MAGIC_BYTES = 0xA94E2A52
)

var (
	imgPath string
	exePath string

	// Globals that need to not be Globals
	entryNames []string
	files      []models.File
)

type TocEntry struct{}

type ImgHeader struct {
	Identifier   uint32
	Version      int32
	EntryCount   int32
	TocSize      int32
	TocEntrySize int16
	Uknown1      int16
}

func (h *ImgHeader) Read(d []byte) {
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

func readFileToBytes(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func init() {
	flag.StringVar(&imgPath, "img", imgPath, "Path to the img file")
	flag.StringVar(&exePath, "exe", exePath, "Path to the exe file")
	flag.Parse()
}

func main() {
	encrypted := false
	imgBytes, err := readFileToBytes(imgPath)
	if err != nil {
		panic(err)
	}
	magicBytes := binary.BigEndian.Uint32(imgBytes[0:4])
	if magicBytes != HEADER_MAGIC_BYTES {
		encrypted = true
	}

	exeBytes, err := readFileToBytes(exePath)
	if err != nil {
		panic(err)
	}

	util.FindAesKey(exeBytes)

	reader := bytes.NewReader(imgBytes)

	headerBytes := make([]byte, 20)
	_, err = reader.Read(headerBytes)
	if err != nil {
		panic(err)
	}

	if encrypted {
		err = util.Decrypt(headerBytes)
		if err != nil {
			panic(err)
		}
	}

	header := ImgHeader{}
	header.Read(headerBytes)

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

	stringData := string(tocBytes[header.EntryCount*int32(header.TocEntrySize):])
	entryNames = strings.Split(stringData, "\x00")

	for i := range int(header.EntryCount) {
		eb := tocBytes[i*int(header.TocEntrySize) : (i+1)*int(header.TocEntrySize)]
		e := img.NewTocEntry(eb)
		d := imgBytes[e.OffsetBlock*0x800 : e.OffsetBlock*0x800+e.Size]
		files = append(files, models.File{
			Name:     entryNames[i],
			Data:     d,
			TocEntry: e,
		})
	}

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}
