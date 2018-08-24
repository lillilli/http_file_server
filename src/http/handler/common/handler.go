package common

import "net/http"

// Handler - структура обработчика запросов
type Handler struct {
}

// New - создание нового обработчика
func New() *Handler {
	return &Handler{}
}

// Health - health роут сервиса
func (h Handler) Health(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("ok"))
}
