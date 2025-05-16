package main

import (
	"bytes"
	"crypto/aes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	HEADER_MAGIC_BYTES = 0xA94E2A52
)

var (
	imgPath string
	exePath string
	offsets = []int{0xA94204, 0xB607C4, 0xB56BC4, 0xB75C9C, 0xB7AEF4, 0xBE6540, 0xBE7540, 0xC95FD8, 0xC5B33C, 0xC5B73C}
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

func init() {
	flag.StringVar(&imgPath, "img", imgPath, "Path to the img file")
	flag.StringVar(&exePath, "exe", exePath, "Path to the exe file")
	flag.Parse()
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

func findAesKey(exeBytes []byte) []byte {
	for _, offset := range offsets {
		key := exeBytes[offset : offset+32]
		if validateAesKey(key) {
			return key
		}
	}
	return nil
}

func validateAesKey(aesKey []byte) bool {
	targetHash := "DEA375EF1E6EF2223A1221C2C575C47BF17EFA5E"
	expectedHash, err := hex.DecodeString(targetHash)
	if err != nil {
		panic(err)
	}

	hash := sha1.New()
	hash.Write(aesKey)
	digest := hash.Sum(nil)

	if len(digest) != len(expectedHash) {
		return false
	}
	for i := range len(digest) {
		if digest[i] != expectedHash[i] {
			return false
		}
	}
	return true
}

func decrypt(data, key []byte) error {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	blockSize := cipher.BlockSize()
	for i := 0; i < len(data); i += blockSize {
		if i+blockSize > len(data) {
			fmt.Printf("Data length is not a multiple of block size %d/%d\n", blockSize, len(data))
			break
		}
		for range 16 {
			cipher.Decrypt(data[i:i+blockSize], data[i:i+blockSize])
		}
	}

	return nil
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

	aesKey := findAesKey(exeBytes)
	if aesKey == nil {
		fmt.Println("AES key not found")
	}

	reader := bytes.NewReader(imgBytes)

	headerBytes := make([]byte, 20)
	_, err = reader.Read(headerBytes)
	if err != nil {
		panic(err)
	}

	if encrypted {
		err = decrypt(headerBytes, aesKey)
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
		err = decrypt(tocBytes, aesKey)
		if err != nil {
			panic(err)
		}
	}

	stringData := string(tocBytes[header.EntryCount*int32(header.TocEntrySize):])
	names := strings.Split(stringData, "\x00")
	fmt.Printf("String data: %v\n", names)
}
