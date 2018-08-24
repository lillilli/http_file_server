package arbitrary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"gitlab2.sqtools.ru/go-pkg/logging"
	"github.com/lillilli/dr_web_exercise/src/api"
	"github.com/lillilli/dr_web_exercise/src/config"
	"github.com/lillilli/dr_web_exercise/src/http/handler"
	handlerTypes "github.com/lillilli/dr_web_exercise/src/http/handler/types"
	"github.com/lillilli/dr_web_exercise/src/path"
	"github.com/lillilli/dr_web_exercise/src/types"
)

// Handler - структура обработчика произвольных файлов
type Handler struct {
	*handler.BaseHandler
	fs              handlerTypes.Fs
	log             logging.Logger
	cfg             *config.Config
	path            *path.Path
	staticServerApi api.StaticServer
	git             handlerTypes.Git
}

// New - создание нового обработчика произвольных файлов
func New(baseHandler *handler.BaseHandler, cfg *config.Config, fs handlerTypes.Fs, git handlerTypes.Git) *Handler {
	return &Handler{
		baseHandler,
		fs,
		logging.NewLogger("arbitrary files handler"),
		cfg,
		path.New(cfg),
		api.NewStaticServerApi(),
		git,
	}
}

// Upload - загрузка произвольного статического файла в кеш
func (h Handler) Upload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	filename := vars["filename"]

	r = h.SetRequestType(r, handler.UploadFileRequestType)

	file, _, err := r.FormFile("upload")
	if err != nil {
		h.SendBadRequestError(r.Context(), w, "can`t upload file", err)
		return
	}

	defer file.Close()

	manifest, err := h.saveFileAndGenerateManifest(file, project, filename)
	if err != nil {
		h.SendBadRequestError(r.Context(), w, "can`t save file", err)
		return
	}

	h.SendJSONResponse(r.Context(), w, manifest)
}

// Update - загрузка произвольного статического файла в кеш и обновление его на всех статик серверах
func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	filename := vars["filename"]

	r = h.SetRequestType(r, handler.UploadFileRequestType)

	file, _, err := r.FormFile("upload")
	if err != nil {
		h.SendBadRequestError(r.Context(), w, "can`t upload file", err)
		return
	}

	defer file.Close()

	staticFilepath := h.path.GetStaticFilepath(project, filename)
	backupFilepath := h.path.GetBackupFilepath(project, filename)

	if err := h.copyFileIfExists(staticFilepath, backupFilepath); err != nil {
		h.SendBadRequestError(r.Context(), w, "can`t backup static file", err)
		return
	}

	manifest, err := h.saveFileAndGenerateManifest(file, project, filename)
	if err != nil {
		h.SendBadRequestError(r.Context(), w, "can`t save file", err)
		return
	}

	staticServers := h.getStaticServers()
	cdnPath := h.path.GetCDNPath(project, filename)

	for _, val := range staticServers {
		url := h.path.GetPurgeFileURL(val, cdnPath)
		if err := h.staticServerApi.Purge(url); err != nil {
			if err := h.copyFileIfExists(backupFilepath, staticFilepath); err != nil {
				h.log.Infof("can`t revert backup file to static")
			}
			h.SendInternalServerError(r.Context(), w, "can't purge file from static server", err)
			return
		}

		url = h.path.GetReceiveFileURL(val, cdnPath)
		bytes, err := h.staticServerApi.GetFile(url)
		if err != nil {
			if err := h.copyFileIfExists(backupFilepath, staticFilepath); err != nil {
				h.log.Infof("can`t revert backup file to static")
			}
			h.SendInternalServerError(r.Context(), w, "can't get file from static server", err)
			return
		}

		// FIXME move this method to Git API
		hash := plumbing.ComputeHash(plumbing.BlobObject, bytes)

		if manifest[filename].Hash != hash.String() {
			h.SendInternalServerError(r.Context(), w, "hashes not equals", err)
			return
		}

	}

	if err := os.Remove(backupFilepath); err != nil {
		h.log.Infof("Can't delete backup file: %s", backupFilepath)
	}

	h.SendJSONResponse(r.Context(), w, manifest)
}

func (h Handler) copyFileIfExists(sourcePath string, destPath string) error {
	exists, err := h.fs.Exists(sourcePath)
	if err != nil {
		return errors.Wrapf(err, `existence check for the source file %s failed`, sourcePath)
	}

	if !exists {
		h.log.Infof("Source file does not exists: %s", sourcePath)
		return nil
	}

	if err := h.fs.MkdirAll(destPath[:strings.LastIndex(destPath, "/")], 0774); err != nil {
		return errors.Wrap(err, "can`t create destination folders")
	}

	return h.fs.CopyFile(sourcePath, destPath, 0666)
}

// UpdateFromGit - загрузка произвольных файлов в кеш и обновление их на всех статик серверах
func (h Handler) UpdateFromGit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]

	r = h.SetRequestType(r, handler.UploadFileRequestType)

	_, ok := h.cfg.Projects[project]
	if !ok {
		h.SendBadRequestError(r.Context(), w, fmt.Sprintf("project %s not exists", project), nil)
		return
	}

	file, _, err := r.FormFile("manifest")
	if err != nil {
		h.SendBadRequestError(r.Context(), w, "can`t upload file", err)
		return
	}

	defer file.Close()
	manifest := make(types.Manifest)

	b, err := ioutil.ReadAll(file)
	if err != nil {
		h.SendBadRequestError(r.Context(), w, "can`t upload manifest", err)
		return
	}

	if err = json.Unmarshal(b, &manifest); err != nil {
		h.SendBadRequestError(r.Context(), w, "parse manifest failed", err)
		return
	}

	preparedManifest := manifest.GroupByCommits()

	repoPath := h.path.GetProjectRepoStorageDir(project)
	repoURL := h.cfg.Projects[project].Repository
	for commit, files := range preparedManifest {
		err := h.git.LoadCommitFiles(
			repoPath, repoURL, commit,
			files,
			func(fileName string, fileReader io.Reader) error {
				staticFilepath := h.path.GetStaticFilepath(project, fileName)
				backupFilepath := h.path.GetBackupFilepath(project, fileName)

				if err := h.copyFileIfExists(staticFilepath, backupFilepath); err != nil {
					h.SendBadRequestError(r.Context(), w, "can`t backup static file", err)
					return errors.Wrap(err, "can`t backup static file")
				}

				if err := h.fs.MkdirAll(staticFilepath[:strings.LastIndex(staticFilepath, "/")], 0774); err != nil {
					h.SendInternalServerError(r.Context(), w, "can`t create project folder", err)
					return errors.Wrap(err, "can`t create project folder")
				}

				err := h.fs.WriteReader(staticFilepath, fileReader, 0666)
				if err != nil {
					h.SendInternalServerError(r.Context(), w, "can't copy file to static dir", err)
					return errors.Wrap(err, "can't copy file to static dir")
				}

				// TODO check the error
				hash, _ := h.git.ComputeReaderHash(fileReader)

				staticServers := h.getStaticServers()
				cdnPath := h.path.GetCDNPath(project, fileName)

				for _, val := range staticServers {
					url := h.path.GetPurgeFileURL(val, cdnPath)
					if err := h.staticServerApi.Purge(url); err != nil {
						if err := h.copyFileIfExists(backupFilepath, staticFilepath); err != nil {
							h.log.Infof("can`t revert backup file to static")
						}
						h.SendInternalServerError(r.Context(), w, "can't purge file from static server", err)
						return errors.Wrap(err, "can't purge file from static server")
					}

					url = h.path.GetReceiveFileURL(val, cdnPath)
					bytes, err := h.staticServerApi.GetFile(url)
					if err != nil {
						if err := h.copyFileIfExists(backupFilepath, staticFilepath); err != nil {
							h.log.Infof("can`t revert backup file to static")
						}
						h.SendInternalServerError(r.Context(), w, "can't get file from static server", err)
						return errors.Wrap(err, "can't get file from static server")
					}

					uploadedHash := h.git.ComputeHash(bytes)

					if uploadedHash != hash {
						h.SendInternalServerError(r.Context(), w, "hashes not equals", err)
						return errors.Wrap(err, "hashes not equals")
					}

				}

				if err := os.Remove(backupFilepath); err != nil {
					h.log.Infof("Can't delete backup file: %s", backupFilepath)
				}

				return nil
			},
		)

		if err != nil {
			return
		}
	}

	h.SendJSONResponse(r.Context(), w, manifest)
}

func (h Handler) getGitFilename(project string, filepath string) (string, error) {
	matched, err := h.fileInExcludeList(project, filepath)
	if err != nil {
		return "", errors.Wrapf(err, "can't match file: %s", filepath)
	}
	if matched {
		return "", errors.Errorf("File in blacklist: %s", filepath)
	}

	return h.path.GetFileCachePath(project, filepath), nil
}

func (h Handler) fileInExcludeList(project string, filename string) (bool, error) {
	for _, reg := range h.cfg.Projects[project].Exclude {
		matched, err := filepath.Match(reg, filename)
		if err != nil || matched {
			return matched, err
		}
	}
	return false, nil
}

func (h Handler) saveFileAndGenerateManifest(file multipart.File, project string, filename string) (map[string]handlerTypes.FileInfo, error) {
	filepath := h.path.GetStaticFilepath(project, filename)
	h.log.Debugf("Creating directories: %s", filepath[:strings.LastIndex(filepath, "/")])
	if err := h.fs.MkdirAll(filepath[:strings.LastIndex(filepath, "/")], 0774); err != nil {
		return nil, errors.Wrap(err, "can`t create project folder")
	}

	h.log.Debugf("Writing data to file")

	buf := bytes.NewBuffer(nil)
	_, _ = io.Copy(buf, file)

	// we need the buffer to compute the hash. That's why we don't use Fs.WriteReader.
	if err := h.fs.WriteFile(filepath, buf.Bytes(), 0644); err != nil {
		return nil, errors.Wrap(err, "can`t write a file")
	}

	//hash := plumbing.ComputeHash(plumbing.BlobObject, buf.Bytes())
	//manifest := GenerateUploadedFileInfo(filename, hash.String())
	hash, _ := h.git.ComputeReaderHash(buf)
	manifest := GenerateUploadedFileInfo(filename, hash)

	return manifest, nil
}

// List - возвращает список произвольных файлов
func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	path := h.path.GetProjectStaticDir(project)

	r = h.SetRequestType(r, handler.GetFileRequestType)

	h.log.Debugf("Reading directory: %s", path)
	files, _ := h.fs.GetFiles(path, true)
	h.SendJSONResponse(r.Context(), w, files)
}

// Remove - удаление произвольного файла
func (h Handler) Remove(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	project := vars["project"]
	filename := vars["filename"]

	r = h.SetRequestType(r, handler.UploadFileRequestType)

	recursive := r.URL.Query().Get("recursive")
	path := h.path.GetStaticFilepath(project, filename)

	if recursive != "" {
		h.log.Debugf("Recursively removing files: %s", path)
		err = h.fs.RemoveAll(path)
	} else {
		h.log.Debugf("Removing file: %s", path)
		err = h.fs.Remove(path)
	}

	if err != nil {
		h.SendInternalServerError(r.Context(), w, "can`t remove", err)
		return
	}

	h.SendOK(r.Context(), w)
}

func (h Handler) getStaticServers() []string {
	var res []string
	for _, val := range h.cfg.StaticServers {
		if _, err := net.LookupHost(val); err == nil {
			res = append(res, val)
		}
	}
	return res
}
