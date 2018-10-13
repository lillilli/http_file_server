package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/lillilli/http_file_server/fs"
)

// FileHandler - files handler structure
type FileHandler struct {
	baseHandler
	fs        fs.Storage
	staticDir string
}

// NewFileHandler - return a new file handler instance
func NewFileHandler(staticDir string, fs fs.Storage) (*FileHandler, error) {
	if exist := fs.Exist(staticDir); !exist {
		return nil, errors.New("dir for static files doesn`t exist")
	}

	return &FileHandler{baseHandler{}, fs, staticDir}, nil
}

// Upload - upload file
func (h FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("upload")
	if err != nil {
		h.sendBadRequestError(w, fmt.Sprintf("can`t upload file: %s", err.Error()))
		return
	}

	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		h.sendBadRequestError(w, "read file bytes failed")
		return
	}

	fileHash, err := MD5String(b)
	if err != nil {
		h.sendInternalError(w, "generating file md5 hash string failed")
		return
	}

	if err := h.fs.WriteFile(getHashedFilepath(h.staticDir, fileHash), b, 0644); err != nil {
		h.sendBadRequestError(w, "file saving failed")
		return
	}

	w.Write([]byte(fileHash))
}

// GetFile - serve files by their hashes
func (h FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["filehash"]

	http.ServeFile(w, r, getHashedFilepath(h.staticDir, hash))
}

// Remove - remove file by his hash
func (h FileHandler) Remove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["filehash"]

	filepath := getHashedFilepath(h.staticDir, hash)

	if exist := h.fs.Exist(filepath); !exist {
		h.sendBadRequestError(w, "file with that name doesn`t exist")
		return
	}

	if err := h.fs.RemoveFile(getHashedFilepath(h.staticDir, hash)); err != nil {
		h.sendInternalError(w, "removing file failed")
		return
	}

	w.Write([]byte("ok"))
}

func getHashedFilepath(staticDir string, hash string) string {
	return fmt.Sprintf("%s/%s", staticDir, hash)
}
