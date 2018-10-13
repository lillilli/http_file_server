package handler

import (
	"net/http"
)

type baseHandler struct{}

func (h baseHandler) sendBadRequestError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func (h baseHandler) sendInternalError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(message))
}
