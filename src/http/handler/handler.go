package handler

import (
	"encoding/json"
	"net/http"

	"github.com/lillilli/http_file_server/src/config"
	"github.com/lillilli/http_file_server/src/types"
	logging "github.com/op/go-logging"
)

// BaseHandler - структура базового обработчика запросов
type BaseHandler struct {
	log *logging.Logger
	cfg *config.Config
}

// New - создание базового нового обработчика
func New(cfg *config.Config) *BaseHandler {
	return &BaseHandler{logging.MustGetLogger("http handler"), cfg}
}

// ParseReqParams - преобразование тела запроса в структуру
func (h BaseHandler) ParseReqParams(req *http.Request, v interface{}) error {
	decoder := json.NewDecoder(req.Body)
	return decoder.Decode(v)
}

// SendInternalServerError - отправка на клиент сообщения о внутренней ошибке
func (h BaseHandler) SendInternalServerError(w http.ResponseWriter, msg string, err error) {
	h.log.Errorf("internal server error: %s : %v", msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

// SendBadRequestError - отправка на клиент сообщения о ошибке запроса
func (h BaseHandler) SendBadRequestError(w http.ResponseWriter, msg string, err error) {
	h.log.Infof("bad request error: %s, error: %v", msg, err)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(msg))
}

// SendJSONResponse - отправка на клиент json сообщения
func (h BaseHandler) SendJSONResponse(w http.ResponseWriter, result interface{}) {
	b, err := json.Marshal(result)
	if err != nil {
		h.SendInternalServerError(w, "can`t marshal json result", err)
		return
	}

	h.log.Infof("send json result: %s", b)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

// SendJSONError - отправка на клиент json сообщения об ошибке
func (h BaseHandler) SendJSONError(w http.ResponseWriter, code int, message string) {
	b, err := json.Marshal(types.Error{Message: message})
	if err != nil {
		h.SendInternalServerError(w, "can`t marshal json result", err)
		return
	}

	h.log.Infof("send json error result: %s", b)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(b)
}

// Send404 - отправка 404 статуса
func (h BaseHandler) Send404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}

// SendOK - отправка ok сообщения клиенту
func (h BaseHandler) SendOK(w http.ResponseWriter) {
	w.Write([]byte("ok"))
}
