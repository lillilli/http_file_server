package test

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"

	"gitlab2.sqtools.ru/go-pkg/metric"
	"github.com/lillilli/dr_web_exercise/src/config"
	"github.com/lillilli/dr_web_exercise/src/mocks"
	"gopkg.in/alexcesaro/statsd.v2"
)

type HandlerScene struct {
	ValidToken  string
	Project     config.Project
	ProjectName string
	ProjectURL  string
	Config      *config.Config
	DB          *mocks.DB
	Stats       *mocks.StatsdCollector
	Fs          *mocks.Fs
	Git         *mocks.Git
}

func NewHandlerScene() *HandlerScene {
	s := &HandlerScene{}
	return s.Init()
}

func (s *HandlerScene) Init() *HandlerScene {
	s.Fs = &mocks.Fs{}
	s.Git = &mocks.Git{}

	s.ValidToken = "valid-project-token"
	s.ProjectURL = "ssh://git@repository.url:1022"
	s.ProjectName = "ValidProject"

	s.Project = config.Project{
		Token:      s.ValidToken,
		Repository: s.ProjectURL,
	}

	s.Config = &config.Config{
		StaticDir:      "/StaticDir",
		CacheDir:       "/CacheDir",
		BackupDir:      "/BackupDir",
		RepoStorageDir: "/RepoStorageDir",
		Projects: map[string]config.Project{
			s.ProjectName: s.Project,
		},
		Stats: config.Stats{
			Metrics: config.MetricsConfiguration{
				Requests: config.Requests{
					GetFile: config.Request{
						Success: config.SuccessRequest{
							GitCount: &metric.StatMetric{
								Type: metric.GaugeType,
								Path: "request.get_file.success.count",
							},
							GitTime: &metric.StatMetric{
								Type: metric.TimingType,
								Path: "request.get_file.success.git_time",
							},
						},
					},
					GetManifest: config.Request{
						Success: config.SuccessRequest{
							GitCount: &metric.StatMetric{
								Type: metric.GaugeType,
								Path: "request.get_manifest.success.count",
							},
							GitTime: &metric.StatMetric{
								Type: metric.TimingType,
								Path: "request.get_manifest.success.git_time",
							},
						},
					},
				},
			},
		},
	}

	s.Config.Projects[s.ProjectName] = s.Project

	s.DB = &mocks.DB{}

	s.Stats = &mocks.StatsdCollector{}

	return s
}

func (s *HandlerScene) GetCookie(name string, w *httptest.ResponseRecorder) (string, error) {
	auxRequest := &http.Request{Header: http.Header{"Cookie": w.HeaderMap["Set-Cookie"]}}
	cookie, err := auxRequest.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (s *HandlerScene) GetBodyAndResponse(w *httptest.ResponseRecorder) (string, *http.Response) {
	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)
	return string(body), res
}

func (s *HandlerScene) CreatePayload(fieldName, fileName, content string) (string, io.Reader, error) {
	payload := new(bytes.Buffer)
	writer := multipart.NewWriter(payload)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return "", nil, err
	}

	if _, err := part.Write([]byte(content)); err != nil {
		return "", nil, err
	}

	if err := writer.Close(); err != nil {
		return "", nil, err
	}

	return writer.FormDataContentType(), payload, nil
}

// Keep all beasts behind the scene
func (s *HandlerScene) GetTiming() statsd.Timing {
	return statsd.Timing{}
}
