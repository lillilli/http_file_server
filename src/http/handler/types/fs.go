package types

import (
	"io"
	"os"
)

// Fs интерфейс файловой системы.
type Fs interface {
	CopyFile(sourcePath, destPath string, perm os.FileMode) error
	DirExists(path string) (bool, error)
	Exists(path string) (bool, error)
	Ls(path string) ([]string, error)
	MkdirAll(name string, mode os.FileMode) error
	GetFiles(dirname string, recursive bool) ([]*FileInfo, error)
	ReadFile(filename string) ([]byte, error)
	Remove(name string) error
	RemoveAll(path string) error
	WriteFile(filename string, data []byte, perm os.FileMode) error
	WriteReader(path string, reader io.Reader, perm os.FileMode) error
}
