package versioned

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"gitlab2.sqtools.ru/go-pkg/logging"
	"gitlab2.sqtools.ru/go-pkg/metric"
	"github.com/lillilli/dr_web_exercise/src/config"
	"github.com/lillilli/dr_web_exercise/src/http/handler"
	handlerTypes "github.com/lillilli/dr_web_exercise/src/http/handler/types"
	"github.com/lillilli/dr_web_exercise/src/path"
	"github.com/lillilli/dr_web_exercise/src/types"
)

// VersionFileRevision - ревизия версионного файла
var VersionFileRevision = regexp.MustCompile("^(.*).(?P<rev>[[:word:]]{7}?).[[:word:]]+$")

// Handler - структура обработчика запросов версионных файлов
type Handler struct {
	*handler.BaseHandler
	cfg       *config.Config
	log       logging.Logger
	collector metric.StatsdCollector

	fs   handlerTypes.Fs
	path *path.Path
	git  handlerTypes.Git
}

// New - создание нового обработчика запросов версионных файлов
func New(baseHandler *handler.BaseHandler, cfg *config.Config, fs handlerTypes.Fs, git handlerTypes.Git, collector metric.StatsdCollector) *Handler {
	return &Handler{baseHandler, cfg, logging.NewLogger("versioned files handler"), collector, fs, path.New(cfg), git}
}

// GetManifest - получение манифеста репозитория
func (h Handler) GetManifest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	tag := vars["tag"]

	r = h.SetRequestType(r, handler.GetManifestRequestType)

	timing := h.collector.StartTiming()
	repoPath := h.path.GetProjectRepoStorageDir(project)
	repoURL := h.cfg.Projects[project].Repository
	err := h.git.CloneOrFetch(repoPath, repoURL, tag)
	if err != nil {
		h.SendInternalServerError(r.Context(), w, "updating local repo copy failed", err)
		return
	}

	h.collector.Send(h.cfg.Stats.Metrics.Requests.GetManifest.Success.GitTime, timing)
	h.collector.Send(h.cfg.Stats.Metrics.Requests.GetManifest.Success.GitCount, 1)

	hash, err := h.git.ResolveRevision(repoPath, tag)
	if err != nil {
		h.SendInternalServerError(r.Context(), w, fmt.Sprintf("can`t get revision for ref: %s", tag), err)
		return
	}

	savedManifests := h.getSavedManifestHashs(project, hash)
	if _, ok := savedManifests[hash]; ok {
		http.ServeFile(w, r, h.path.GetManifestCachePath(project, hash))
		return
	}

	manifest, err := h.git.CreateManifestForFiles(repoPath, hash, func(fileName string) bool {
		return h.path.InStaticProjectPath(project, fileName) &&
			!h.path.IsProjectFileExclude(project, fileName)
	})

	h.SendJSONResponse(r.Context(), w, manifest)
	h.saveManifestToFileSystem(project, hash, manifest)
}

// WarmUp - прогрев манифеста
func (h Handler) WarmUp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]

	r = h.SetRequestType(r, handler.UploadFileRequestType)

	file, _, err := r.FormFile("manifest")
	if err != nil {
		h.SendBadRequestError(r.Context(), w, "can`t upload manifest", err)
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

	h.warmUpManifest(project, manifest)
	h.SendOK(r.Context(), w)
}

// GetUncachedFile - получение незакэшированного файла
func (h Handler) GetUncachedFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project"]
	filepath := vars["filepath"]

	r = h.SetRequestType(r, handler.GetFileRequestType)

	if matched := VersionFileRevision.MatchString(filepath); !matched {
		err := fmt.Errorf("filename validation failed, doesn`t match for %s", VersionFileRevision.String())
		h.SendBadRequestError(r.Context(), w, err.Error(), err)
		return
	}

	rev := VersionFileRevision.ReplaceAllString(filepath, "${rev}")
	filepath = strings.Replace(filepath, "."+rev, "", 1)

	repoPath := h.path.GetProjectRepoStorageDir(project)
	repoURL := h.cfg.Projects[project].Repository
	_ = h.git.CloneOrFetch(repoPath, repoURL, rev)

	files := []string{h.path.GetReversedFilePath(project, filepath)}
	// FIXME get rid of WG!
	wg := &sync.WaitGroup{}
	wg.Add(1)
	h.warmUpCommitFiles(wg, project, rev, files)
	wg.Wait()

	http.ServeFile(w, r, h.path.GetFileCachePath(project, filepath))
}
