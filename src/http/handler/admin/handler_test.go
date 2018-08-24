package admin_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"gitlab2.sqtools.ru/go-pkg/sqauth"

	"github.com/stretchr/testify/assert"
	"github.com/lillilli/dr_web_exercise/src/db/models"
	baseHandler "github.com/lillilli/dr_web_exercise/src/http/handler"
	. "github.com/lillilli/dr_web_exercise/src/http/handler/admin"
)

func TestTokenAuthShouldRespondWithTokenNotExistsWhenTokenNotExists(t *testing.T) {
	scene := NewAdminHandlerScene()
	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)

	notExistingToken := "no-such-a-token-exists"
	url := fmt.Sprintf(`/admin/auth?token=%s`, notExistingToken)
	r := httptest.NewRequest("GET", url, nil)

	w := httptest.NewRecorder()

	handler.TokenAuth(w, r)
	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Contains(t, string(body), "token not exists")
}

func TestTokenAuthShouldSetCookiesAndRespondWithJSONReportingProjectAndToken(t *testing.T) {
	scene := NewAdminHandlerScene()
	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)

	url := fmt.Sprintf(`/admin/auth?token=%s`, scene.ValidToken)
	r := httptest.NewRequest("GET", url, nil)

	w := httptest.NewRecorder()

	handler.TokenAuth(w, r)
	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	expectedBody := fmt.Sprintf(`{"token":"%s","project":"%s"}`, scene.ValidToken, scene.ProjectName)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, expectedBody, string(body))

	actualTokenCookie, err := scene.GetCookie("token", w)
	if assert.Nil(t, err) {
		assert.Equal(t, scene.ValidToken, actualTokenCookie)
	}

	actualProjectCookie, err := scene.GetCookie("project", w)
	if assert.Nil(t, err) {
		assert.Equal(t, scene.ProjectName, actualProjectCookie)
	}
}

func TestGetTokensReturnsTokensList(t *testing.T) {
	scene := NewAdminHandlerScene()

	description := "some description"
	project := "a project"
	counter := 100
	scene.UserTokenManagerMock.On("GetByUID", "test").Return([]*models.TokenAndCounter{
		&models.TokenAndCounter{
			Token:       "token",
			Description: &description,
			Project:     &project,
			Counter:     &counter,
		},
	}, nil)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)

	r := httptest.NewRequest("GET", "/admin/token", nil)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	handler.GetTokens(w, r.WithContext(ctx))
	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `[{"token":"token","description":"some description","project":"a project","counter":100}]`, string(body))
}

func TestGetTokensReturnsInternalErrorWhenDBQueryFails(t *testing.T) {
	scene := NewAdminHandlerScene()

	fakeDBError := errors.New("fake DB error")

	scene.UserTokenManagerMock.On("GetByUID", "test").Return(nil, fakeDBError)

	r := httptest.NewRequest("GET", "/admin/token", nil)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.GetTokens(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, string(body), "can't get tokens from db")
}

func TestAddTokenRespondsWithUserToken(t *testing.T) {
	scene := NewAdminHandlerScene()

	description := "some description"
	scene.UserTokenManagerMock.On(
		"Add",
		"test",
		mock.AnythingOfType("string"),
		"some description",
	).Return(
		&models.UserToken{
			ID:          1,
			Token:       "added token",
			UID:         "user id",
			Description: &description,
		},
		nil,
	)

	payload := strings.NewReader(`{"description":"some description"}`)
	r := httptest.NewRequest("POST", "/admin/token", payload)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.AddToken(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, `{"id":1,"token":"added token","uid":"user id","description":"some description"}`, string(body))
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestAddTokenRespondsWithInternalServerErrorOnBadPayload(t *testing.T) {
	scene := NewAdminHandlerScene()

	payload := strings.NewReader(`bad json`)
	r := httptest.NewRequest("POST", "/admin/token", payload)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.AddToken(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Contains(t, `can't parse request params`, string(body))
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestAddTokenRespondsWithInternalServerErrorOnDBError(t *testing.T) {
	scene := NewAdminHandlerScene()

	fakeDbError := errors.New("fake DB error")
	scene.UserTokenManagerMock.On(
		"Add",
		"test",
		mock.AnythingOfType("string"),
		"some description",
	).Return(
		nil,
		fakeDbError,
	)

	payload := strings.NewReader(`{"description":"some description"}`)
	r := httptest.NewRequest("POST", "/admin/token", payload)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.AddToken(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Contains(t, `can't add token to db`, string(body))
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestAddCounterRespondsWithInternalServerErrorOnBadPayload(t *testing.T) {
	scene := NewAdminHandlerScene()
	payload := strings.NewReader("bad json")

	r := httptest.NewRequest("POST", "/admin/counter", payload)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.AddCounter(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, "can't parse request params", string(body))
}

func TestAddCounterRespondsWithInternalServerErrorOnTokenRetrievalFailure(t *testing.T) {
	scene := NewAdminHandlerScene()
	payload := strings.NewReader(`{"token":"token","project":"project"}`)

	fakeDBError := errors.New("fake db error")
	scene.UserTokenManagerMock.On(
		"GetByToken",
		"token",
	).Return(nil, fakeDBError)

	r := httptest.NewRequest("POST", "/admin/counter", payload)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.AddCounter(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, "can't get token from db", string(body))
}

func TestAddCounterRespondsWithBadRequestWhenTokenDoesNotExist(t *testing.T) {
	scene := NewAdminHandlerScene()
	payload := strings.NewReader(`{"token":"token","project":"project"}`)

	scene.UserTokenManagerMock.On(
		"GetByToken",
		"token",
	).Return(nil, nil)

	r := httptest.NewRequest("POST", "/admin/counter", payload)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.AddCounter(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Contains(t, "token not exists", string(body))
}

func TestAddCounterRespondsWithInternalServerErrorOnCounterAddError(t *testing.T) {
	scene := NewAdminHandlerScene()
	payload := strings.NewReader(`{"token":"token","project":"project"}`)

	description := "description"
	project := "project"
	counter := 122
	tokenAndCounter := &models.TokenAndCounter{
		Token:       "token",
		Description: &description,
		Project:     &project,
		Counter:     &counter,
	}

	scene.UserTokenManagerMock.On(
		"GetByToken",
		"token",
	).Return(tokenAndCounter, nil)

	fakeDBError := errors.New("fake db error")
	scene.TokenCountersManager.On("Add", "token", "project").Return(nil, fakeDBError)

	r := httptest.NewRequest("POST", "/admin/counter", payload)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.AddCounter(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, "can't add counter to db", string(body))
}

func TestAddCounterRespondsWithCounter(t *testing.T) {
	scene := NewAdminHandlerScene()
	payload := strings.NewReader(`{"token":"token","project":"project"}`)

	description := "description"
	project := "project"
	counter := 122
	tokenAndCounter := &models.TokenAndCounter{
		Token:       "token",
		Description: &description,
		Project:     &project,
		Counter:     &counter,
	}

	scene.UserTokenManagerMock.On(
		"GetByToken",
		"token",
	).Return(tokenAndCounter, nil)

	resProject := "a project"
	resCounter := 123
	tokenCounter := &models.TokenCounter{
		ID:      23,
		Token:   "a token",
		Project: &resProject,
		Counter: &resCounter,
	}
	scene.TokenCountersManager.On("Add", "token", "project").Return(tokenCounter, nil)

	r := httptest.NewRequest("POST", "/admin/counter", payload)
	w := httptest.NewRecorder()

	user := &sqauth.User{
		UID:        "test",
		Roles:      []string{},
		RepoAccess: []string{},
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, sqauth.ContextKey, user)

	baseHandler := baseHandler.New(scene.Stats, scene.Config)
	handler := New(baseHandler, scene.Config, scene.DB)
	handler.AddCounter(w, r.WithContext(ctx))

	res := w.Result()
	body, _ := ioutil.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, `{"id":23,"token":"a token","project":"a project","counter":123}`, string(body))
}
