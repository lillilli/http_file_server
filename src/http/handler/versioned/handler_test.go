package versioned_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	baseHandler "github.com/lillilli/dr_web_exercise/src/http/handler"
	"github.com/lillilli/dr_web_exercise/src/types"

	. "github.com/lillilli/dr_web_exercise/src/http/handler/versioned"
)

func TestGetManifestBuildsManifest(t *testing.T) {
	s := NewVersionedHandlerScene()

	tag := "tag"
	commitSHA := "commit-sha"
	repoPath := "/RepoStorageDir/" + s.ProjectName
	cachePath := "/CacheDir/" + s.ProjectName

	manifest := types.Manifest{
		"name1": types.Version{Version: "1234567"},
		"name2": types.Version{Version: "1234567"},
	}

	s.Fs.On(
		"MkdirAll",
		cachePath,
		mock.MatchedBy(func(m os.FileMode) bool {
			return uint32(m) == 0744
		}),
	).Return(nil)

	// Cache is empty
	s.Fs.On(
		"Ls",
		cachePath,
	).Return([]string{}, nil)

	// Cache the manifest for the revision "commit-sha"
	s.Fs.On(
		"WriteFile",
		fmt.Sprintf("%s/%s/%s.json", s.Config.CacheDir, s.ProjectName, commitSHA),
		mock.MatchedBy(func(bb []byte) bool {
			return string(bb) == `{"name1":{"version":"1234567"},"name2":{"version":"1234567"}}`
		}),
		os.FileMode(0644),
	).Return(nil)

	s.Git.On(
		"CloneOrFetch",
		repoPath,
		s.ProjectURL,
		tag,
	).Return(nil)

	// tag resolves to the revision commitSHA
	s.Git.On(
		"ResolveRevision",
		repoPath,
		tag,
	).Return(commitSHA, nil)

	s.Git.On(
		"CreateManifestForFiles",
		repoPath,
		commitSHA,
		mock.MatchedBy(func(f func(string) bool) bool {
			return f != nil
		}),
	).Return(manifest, nil)

	timing := s.GetTiming()
	s.Stats.On("StartTiming").Return(timing)
	s.Stats.On(
		"Send",
		s.Config.Stats.Metrics.Requests.GetManifest.Success.GitTime,
		timing,
	).Return()
	s.Stats.On("Send", s.Config.Stats.Metrics.Requests.GetManifest.Success.GitCount, 1).Return()

	w := httptest.NewRecorder()
	r := mux.SetURLVars(
		httptest.NewRequest("GET", "/api/manifest/project/tag", nil),
		map[string]string{
			"project": s.ProjectName,
			"tag":     tag,
		},
	)

	baseHandler := baseHandler.New(s.Stats, s.Config)
	handler := New(baseHandler, s.Config, s.Fs, s.Git, s.Stats)
	handler.GetManifest(w, r)

	body, res := s.GetBodyAndResponse(w)

	assert.NotNil(t, handler)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `{"name1":{"version":"1234567"},"name2":{"version":"1234567"}}`, body)
}

func TestWarmUpLoadsFilesFromLocalRepoAndWritesThemToCache(t *testing.T) {
	s := NewVersionedHandlerScene()
	repoPath := "/RepoStorageDir/" + s.ProjectName

	// Load files from local repo according to the manifest
	fileReader1 := bytes.NewBufferString("content1")
	fileReader2 := bytes.NewBufferString("content2")
	s.Git.On(
		"LoadCommitFiles",
		repoPath,
		s.ProjectURL,
		"1234567",
		mock.MatchedBy(func(names []string) bool {
			res := strings.Join(names, ":")
			return res == "name1:name2" || res == "name2:name1"
		}),
		mock.MatchedBy(func(fn func(string, io.Reader) error) bool {
			return true
		}),
	).Run(func(args mock.Arguments) {
		callback, ok := args.Get(4).(func(string, io.Reader) error)
		if !ok {
			t.Fatalf("Can't cast the LoadCommitFiles callback")
		}

		_ = callback("name1", fileReader1)
		_ = callback("name2", fileReader2)
	}).Return(nil)

	// write them to chache directory
	s.Fs.On(
		"WriteReader",
		"/CacheDir/"+s.ProjectName+"/name1",
		fileReader1,
		os.FileMode(0644),
	).Return(nil)

	s.Fs.On(
		"WriteReader",
		"/CacheDir/"+s.ProjectName+"/name2",
		fileReader2,
		os.FileMode(0644),
	).Return(nil)

	manifest := `{"name1":{"version":"1234567"},"name2":{"version":"1234567"}}`
	contentType, payload, err := s.CreatePayload("manifest", "manifest.json", manifest)
	if err != nil {
		t.Fatalf("can't create payload: %s", err.Error())
	}

	w := httptest.NewRecorder()
	r := mux.SetURLVars(
		httptest.NewRequest("POST", "/api/warmup/project", payload),
		map[string]string{
			"project": s.ProjectName,
		},
	)
	r.Header.Set("Content-Type", contentType)

	baseHandler := baseHandler.New(s.Stats, s.Config)
	handler := New(baseHandler, s.Config, s.Fs, s.Git, s.Stats)
	handler.WarmUp(w, r)

	body, res := s.GetBodyAndResponse(w)

	assert.NotNil(t, handler)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `ok`, body)
}

func TestGetUncachedFileCachesTheFileAndRespondsWithThat(t *testing.T) {
	s := NewVersionedHandlerScene()
	repoPath := "/RepoStorageDir/" + s.ProjectName
	s.Config.CacheDir = "./test"

	s.Git.On(
		"CloneOrFetch",
		repoPath,
		s.ProjectURL,
		"1234567",
	).Return(nil)

	fileContent := "content"
	fileReader := bytes.NewBufferString(fileContent)
	s.Git.On(
		"LoadCommitFiles",
		repoPath,
		s.ProjectURL,
		"1234567",
		[]string{"filename.ext"},
		mock.MatchedBy(func(fn func(string, io.Reader) error) bool {
			return true
		}),
	).Run(func(args mock.Arguments) {
		callback, ok := args.Get(4).(func(string, io.Reader) error)
		if !ok {
			t.Fatalf("Can't cast the LoadCommitFiles callback")
		}

		_ = callback("filename.ext", fileReader)
	}).Return(nil)

	s.Fs.On(
		"WriteReader",
		"./test/"+s.ProjectName+"/filename.ext",
		fileReader,
		os.FileMode(0644),
	).Return(nil)

	w := httptest.NewRecorder()
	r := mux.SetURLVars(
		httptest.NewRequest("GET", "/{project}/{filepath:.*}", nil),
		map[string]string{
			"project":  s.ProjectName,
			"filepath": "filename.1234567.ext",
		},
	)

	baseHandler := baseHandler.New(s.Stats, s.Config)
	handler := New(baseHandler, s.Config, s.Fs, s.Git, s.Stats)
	handler.GetUncachedFile(w, r)

	body, res := s.GetBodyAndResponse(w)

	assert.NotNil(t, handler)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `content`, body)
}
