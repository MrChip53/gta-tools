package rage

import "strings"

type FileType int

const (
	FileTypeNone FileType = iota
	FileTypeImg
	FileTypeRpf
	FileTypeScript
)

func GetFileType(path string) FileType {
	if strings.HasSuffix(path, ".img") {
		return FileTypeImg
	} else if strings.HasSuffix(path, ".rpf") {
		return FileTypeRpf
	} else if strings.HasSuffix(path, ".sco") {
		return FileTypeScript
	}
	return FileTypeNone
}
