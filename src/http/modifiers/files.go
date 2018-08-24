package modifiers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// DublicateFileData - дублиррование данных, находящихся внутри файла
func DublicateFileData(buf *bytes.Buffer) error {
	bufCopy := bytes.NewBuffer(buf.Bytes())
	_, err := io.Copy(buf, bufCopy)
	return err
}

// DublicateFileName - переименование файла (дублирование его названия)
func DublicateFileName(filepath string) error {
	filename := filepath[strings.LastIndex(filepath, "/")+1:]
	return os.Rename(filepath, fmt.Sprintf("%s%s", filepath, filename))
}
