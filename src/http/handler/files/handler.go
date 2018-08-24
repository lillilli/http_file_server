package files

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/lillilli/http_file_server/src/config"
	"github.com/lillilli/http_file_server/src/http/handler"
)

// Handler - структура обработчика произвольных файлов
type Handler struct {
	*handler.BaseHandler
	log *logging.Logger
	cfg *config.Config
	fs  afero.Fs
}

// New - создание нового обработчика произвольных файлов
func New(baseHandler *handler.BaseHandler, cfg *config.Config) (*Handler, error) {
	AppFs := afero.NewOsFs()

	exist, err := afero.DirExists(AppFs, cfg.StaticDir)
	if !exist {
		return nil, errors.New("dir for static files doesn`t exist")
	}

	return &Handler{
		baseHandler,
		logging.MustGetLogger("files handler"),
		cfg,
		AppFs,
	}, err
}

// Upload - загрузка произвольного статического файла в кеш
func (h Handler) Upload(opts ...UploadModifier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("upload")
		if err != nil {
			h.SendBadRequestError(w, fmt.Sprintf("can`t upload file: %s", err.Error()), err)
			return
		}

		defer file.Close()

		fileBuffer := bytes.NewBuffer(nil)
		if _, err = io.Copy(fileBuffer, file); err != nil {
			h.SendInternalServerError(w, "file copying failed", err)
			return
		}

		modifiers := newModifiers(opts)

		if modifiers.PreCallbackEnabled {
			if err = modifiers.PreCallback(fileBuffer); err != nil {
				h.SendBadRequestError(w, "file modifcation failed", err)
				return
			}
		}

		hash, err := h.saveFile(fileBuffer)
		if err != nil {
			h.SendBadRequestError(w, "file saving failed", err)
			return
		}

		if modifiers.PostCallbackEnabled {
			if err := modifiers.PostCallback(getHashedFilepath(h.cfg.StaticDir, hash)); err != nil {
				h.SendBadRequestError(w, "file modifcation failed", err)
				return
			}
		}

		h.SendJSONResponse(w, hash)
	}
}

func (h Handler) saveFile(fileBuffer *bytes.Buffer) (hash string, err error) {
	if hash, err = h.computeMD5HashString(fileBuffer.Bytes()); err != nil {
		return "", errors.Wrap(err, "can`t compute file hash")
	}

	h.log.Debugf("Creating directories: %s", getHashedFiledir(h.cfg.StaticDir, hash))

	if err := h.fs.MkdirAll(getHashedFiledir(h.cfg.StaticDir, hash), 0774); err != nil {
		return "", errors.Wrap(err, "can`t create destination folders")
	}

	h.log.Debugf("Writing data to file")
	if err := afero.WriteFile(h.fs, getHashedFilepath(h.cfg.StaticDir, hash), fileBuffer.Bytes(), 0644); err != nil {
		return "", errors.Wrap(err, "can`t write a file")
	}

	return hash, nil
}

func (h Handler) computeMD5HashString(src []byte) (string, error) {
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, bytes.NewBuffer(src)); err != nil {
		return "", err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	return hex.EncodeToString(hashInBytes), nil
}

// GetFile - получение файла по его хэшу
func (h Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["filehash"]

	h.log.Debugf("client want to get file: %s", getHashedFilepath(h.cfg.StaticDir, hash))
	http.ServeFile(w, r, getHashedFilepath(h.cfg.StaticDir, hash))
}

// Remove - удаление файла по его хэшу
func (h Handler) Remove(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash := vars["filehash"]

	filepath := getHashedFilepath(h.cfg.StaticDir, hash)

	exist, err := afero.Exists(h.fs, filepath)
	if err != nil {
		h.SendInternalServerError(w, "removing file failed", err)
		return
	}

	if !exist {
		h.SendBadRequestError(w, "file with that name doesn`t exist", nil)
		return
	}

	h.log.Debugf("Removing file")
	if err := h.fs.Remove(getHashedFilepath(h.cfg.StaticDir, hash)); err != nil {
		h.SendInternalServerError(w, "removing file failed", err)
		return
	}

	h.SendOK(w)
}

func getHashedFiledir(staticDir string, hash string) string {
	return fmt.Sprintf("%s/%s", staticDir, hash[:2])
}

func getHashedFilepath(staticDir string, hash string) string {
	return fmt.Sprintf("%s/%s/%s", staticDir, hash[:2], hash)
}
