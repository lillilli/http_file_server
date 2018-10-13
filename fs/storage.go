package fs

import (
	"io/ioutil"
	"os"
)

// Storage - interface of file system storage
type Storage interface {
	WriteFile(path string, data []byte, perm os.FileMode) error
	RemoveFile(filepath string) error
	Exist(path string) bool
}

type storage struct{}

// NewStorage - return a new storage instance
func NewStorage() Storage {
	return &storage{}
}

// Exist - return true if file or directory exist
func (s storage) Exist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// WriteFile - write a file to a file system
func (s storage) WriteFile(path string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(path, data, perm)
}

// RemoveFile - remove file from file system
func (s storage) RemoveFile(filepath string) error {
	return os.Remove(filepath)
}
