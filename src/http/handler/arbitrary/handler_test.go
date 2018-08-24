package arbitrary_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	baseHandler "github.com/lillilli/dr_web_exercise/src/http/handler"
	. "github.com/lillilli/dr_web_exercise/src/http/handler/arbitrary"
	handlerTypes "github.com/lillilli/dr_web_exercise/src/http/handler/types"
)

func TestUploadUploadsTheFile(t *testing.T) {
	s := NewArbitraryHandlerScene()
	fileContent := `file content`

	s.Fs.On("MkdirAll", "/StaticDir/ValidProject", os.FileMode(0774)).Return(nil)
	s.Fs.On(
		"WriteFile",
		"/StaticDir/ValidProject/filename.ext",
		[]byte(fileContent),
		os.FileMode(0644),
	).Return(nil)

	s.Git.On("ComputeReaderHash", mock.Anything).Return("content-hash", nil)

	contentType, payload, err := s.CreatePayload("upload", "filename.ext", fileContent)
	if err != nil {
		t.Fatalf("can't create payload: %s", err.Error())
	}

	w := httptest.NewRecorder()
	r := mux.SetURLVars(
		httptest.NewRequest("POST", "/api/upload/ProjectName/filename.ext", payload),
		map[string]string{
			"project":  s.ProjectName,
			"filename": "filename.ext",
		},
	)
	r.Header.Set("Content-Type", contentType)

	baseHandler := baseHandler.New(s.Stats, s.Config)
	handler := New(baseHandler, s.Config, s.Fs, s.Git)
	handler.Upload(w, r)

	body, res := s.GetBodyAndResponse(w)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `{"filename.ext":{"hash":"content-hash"}}`, body)
}

func TestUpdateUpdatesTheFile(t *testing.T) {
	s := NewArbitraryHandlerScene()

	newFileContent := `file content`

	s.Fs.On(
		"Exists",
		"/StaticDir/ValidProject/filename.ext",
	).Return(true, nil)

	// first back up old version
	s.Fs.On("MkdirAll", "/BackupDir/ValidProject", os.FileMode(0774)).Return(nil)
	s.Fs.On(
		"CopyFile",
		"/StaticDir/ValidProject/filename.ext",
		"/BackupDir/ValidProject/filename.ext",
		os.FileMode(0666),
	).Return(nil)

	// now update content
	s.Fs.On("MkdirAll", "/StaticDir/ValidProject", os.FileMode(0774)).Return(nil)
	s.Fs.On(
		"WriteFile",
		"/StaticDir/ValidProject/filename.ext",
		[]byte(newFileContent),
		os.FileMode(0644),
	).Return(nil)

	s.Git.On("ComputeReaderHash", mock.Anything).Return("new-hash", nil)

	contentType, payload, err := s.CreatePayload("upload", "filename.ext", newFileContent)
	if err != nil {
		t.Fatalf("can't create payload: %s", err.Error())
	}

	w := httptest.NewRecorder()
	r := mux.SetURLVars(
		httptest.NewRequest("POST", "/api/update/{project}/{filename:.*}", payload),
		map[string]string{
			"project":  s.ProjectName,
			"filename": "filename.ext",
		},
	)
	r.Header.Set("Content-Type", contentType)

	baseHandler := baseHandler.New(s.Stats, s.Config)
	handler := New(baseHandler, s.Config, s.Fs, s.Git)
	handler.Update(w, r)

	body, res := s.GetBodyAndResponse(w)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `{"filename.ext":{"hash":"new-hash"}}`, body)
}

func TestUpdateFromGitDoesWhatItShouldDo(t *testing.T) {
	s := NewArbitraryHandlerScene()

	manifest := `{"name1":{"version":"1234567"},"name2":{"version":"1234567"}}`

	reader1 := bytes.NewBufferString("content1")
	reader2 := bytes.NewBufferString("content2")
	s.Git.On(
		"LoadCommitFiles",
		"/RepoStorageDir/ValidProject",
		"ssh://git@repository.url:1022",
		"1234567",
		mock.MatchedBy(func(names []string) bool {
			arg := strings.Join(names, ":")
			return arg == "name1:name2" || arg == "name2:name1"
		}),
		mock.MatchedBy(func(action func(string, io.Reader) error) bool {
			return true
		}),
	).Run(func(args mock.Arguments) {
		action, ok := args.Get(4).(func(string, io.Reader) error)
		if !ok {
			t.Fatalf("can't get LoadCommitFiles action")
		}

		_ = action("name1", reader1)
		_ = action("name2", reader2)
	}).Return(nil)

	s.Git.On(
		"ComputeReaderHash",
		reader1,
	).Return("111", nil)

	s.Git.On(
		"ComputeReaderHash",
		reader2,
	).Return("222", nil)

	s.Fs.On("Exists", "/StaticDir/ValidProject/name1").Return(true, nil)
	s.Fs.On("Exists", "/StaticDir/ValidProject/name2").Return(true, nil)

	s.Fs.On("MkdirAll", "/StaticDir/ValidProject", os.FileMode(0774)).Return(nil)
	s.Fs.On("MkdirAll", "/BackupDir/ValidProject", os.FileMode(0774)).Return(nil)

	s.Fs.On(
		"CopyFile",
		"/StaticDir/ValidProject/name1",
		"/BackupDir/ValidProject/name1",
		os.FileMode(0666),
	).Return(nil)

	s.Fs.On(
		"CopyFile",
		"/StaticDir/ValidProject/name2",
		"/BackupDir/ValidProject/name2",
		os.FileMode(0666),
	).Return(nil)

	s.Fs.On(
		"WriteReader",
		"/StaticDir/ValidProject/name1",
		reader1,
		os.FileMode(0666),
	).Return(nil)

	s.Fs.On(
		"WriteReader",
		"/StaticDir/ValidProject/name2",
		reader2,
		os.FileMode(0666),
	).Return(nil)

	contentType, payload, err := s.CreatePayload("manifest", "manifest.json", manifest)
	if err != nil {
		t.Fatalf("can't create payload: %s", err.Error())
	}

	w := httptest.NewRecorder()
	r := mux.SetURLVars(
		httptest.NewRequest("POST", "/api/update-from-git/{project}", payload),
		map[string]string{
			"project": s.ProjectName,
		},
	)
	r.Header.Set("Content-Type", contentType)

	baseHandler := baseHandler.New(s.Stats, s.Config)
	handler := New(baseHandler, s.Config, s.Fs, s.Git)
	handler.UpdateFromGit(w, r)

	body, res := s.GetBodyAndResponse(w)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, manifest, body)
}

func TestListListsArbitraryFiles(t *testing.T) {
	s := NewArbitraryHandlerScene()

	files := []*handlerTypes.FileInfo{
		&handlerTypes.FileInfo{
			Name: "file name",
		},
	}
	s.Fs.On(
		"GetFiles",
		"/StaticDir/ValidProject",
		true,
	).Return(files, nil)

	w := httptest.NewRecorder()
	r := mux.SetURLVars(
		httptest.NewRequest("GET", "/api/files/{project}", nil),
		map[string]string{
			"project": s.ProjectName,
		},
	)

	baseHandler := baseHandler.New(s.Stats, s.Config)
	handler := New(baseHandler, s.Config, s.Fs, s.Git)
	handler.List(w, r)

	body, res := s.GetBodyAndResponse(w)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `[{"name":"file name"}]`, body)
}

func TestRemoveRemovesTheFile(t *testing.T) {
	s := NewArbitraryHandlerScene()

	s.Fs.On(
		"Remove",
		"/StaticDir/ValidProject/filename.ext",
	).Return(nil)

	w := httptest.NewRecorder()
	r := mux.SetURLVars(
		httptest.NewRequest("DELETE", "/api/remove/{project}/{filename:.*}", nil),
		map[string]string{
			"project":  s.ProjectName,
			"filename": "filename.ext",
		},
	)

	baseHandler := baseHandler.New(s.Stats, s.Config)
	handler := New(baseHandler, s.Config, s.Fs, s.Git)
	handler.Remove(w, r)

	body, res := s.GetBodyAndResponse(w)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `ok`, body)
}
