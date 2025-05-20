package util

import (
	"crypto/aes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

var (
	offsets = []int{0xA94204, 0xB607C4, 0xB56BC4, 0xB75C9C, 0xB7AEF4, 0xBE6540, 0xBE7540, 0xC95FD8, 0xC5B33C, 0xC5B73C}
	aesKey  []byte
)

func FindAesKey(exeBytes []byte) {
	for _, offset := range offsets {
		key := exeBytes[offset : offset+32]
		if ValidateAesKey(key) {
			aesKey = key
			return
		}
	}
}

func ValidateAesKey(aesKey []byte) bool {
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

func Decrypt(data []byte) error {
	if aesKey == nil {
		return fmt.Errorf("AES key not set")
	}
	cipher, err := aes.NewCipher(aesKey)
	if err != nil {
		return err
	}

	blockSize := cipher.BlockSize()
	for i := 0; i < len(data); i += blockSize {
		if i+blockSize > len(data) {
			break
		}
		for range 16 {
			cipher.Decrypt(data[i:i+blockSize], data[i:i+blockSize])
		}
	}

	return nil
}
