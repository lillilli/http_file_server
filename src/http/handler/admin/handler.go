package admin

import (
	"fmt"
	"net/http"

	"github.com/satori/go.uuid"
	"gitlab2.sqtools.ru/go-pkg/logging"
	"gitlab2.sqtools.ru/go-pkg/sqauth"
	"github.com/lillilli/dr_web_exercise/src/config"
	"github.com/lillilli/dr_web_exercise/src/http/handler"
	"github.com/lillilli/dr_web_exercise/src/types"
)

type Handler struct {
	*handler.BaseHandler

	config *config.Config
	db     types.DB
	log    logging.Logger
}

func New(baseHandler *handler.BaseHandler, cfg *config.Config, db types.DB) *Handler {
	return &Handler{baseHandler, cfg, db, logging.NewLogger("admin handler")}
}

func (h Handler) TokenAuth(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	for key, val := range h.config.Projects {
		if token == val.Token {
			projectCookie := &http.Cookie{Name: "project", Value: key}
			tokenCookie := &http.Cookie{Name: "token", Value: token}
			http.SetCookie(w, projectCookie)
			http.SetCookie(w, tokenCookie)
			h.SendJSONResponse(r.Context(), w, struct {
				Token   string `json:"token"`
				Project string `json:"project"`
			}{token, key})
			return
		}
	}

	h.SendBadRequestError(r.Context(), w, "token not exists", nil)
}

func (h Handler) GetTokens(w http.ResponseWriter, r *http.Request) {
	user, err := sqauth.GetUser(r)
	if err != nil {
		h.SendInternalServerError(r.Context(), w, "can't get sqauth user", err)
		return
	}

	counters, err := h.db.UserTokens().GetByUID(user.UID)
	if err != nil {
		h.SendInternalServerError(r.Context(), w, "can't get tokens from db", err)
		return
	}

	h.SendJSONResponse(r.Context(), w, counters)
}

type addTokenParams struct {
	Description string `json:"description"`
}

func (h Handler) AddToken(w http.ResponseWriter, r *http.Request) {
	var params addTokenParams

	if err := h.ParseReqParams(r, &params); err != nil {
		h.SendInternalServerError(r.Context(), w, "can't parse request params", err)
		return
	}

	user, err := sqauth.GetUser(r)
	if err != nil {
		h.SendInternalServerError(r.Context(), w, "can't get sqauth user", err)
		return
	}

	token := uuid.NewV4()

	newToken, err := h.db.UserTokens().Add(user.UID, fmt.Sprintf("%s", token), params.Description)
	if err != nil {
		h.SendInternalServerError(r.Context(), w, "can't add token to db", err)
		return
	}

	h.SendJSONResponse(r.Context(), w, newToken)
}

type addCounterParams struct {
	Token   string `json:"token"`
	Project string `json:"project"`
}

func (h Handler) AddCounter(w http.ResponseWriter, r *http.Request) {
	var params addCounterParams

	if err := h.ParseReqParams(r, &params); err != nil {
		h.SendInternalServerError(r.Context(), w, "can't parse request params", err)
		return
	}

	token, err := h.db.UserTokens().GetByToken(params.Token)
	if err != nil {
		h.SendInternalServerError(r.Context(), w, "can't get token from db", err)
		return
	}
	if token == nil {
		h.SendBadRequestError(r.Context(), w, "token not exists", nil)
		return
	}

	counter, err := h.db.TokenCounters().Add(params.Token, params.Project)
	if err != nil {
		h.SendInternalServerError(r.Context(), w, "can't add counter to db", err)
		return
	}

	h.SendJSONResponse(r.Context(), w, counter)
}
