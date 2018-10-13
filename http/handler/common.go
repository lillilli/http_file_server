package handler

import "net/http"

// HealthHandler - handle health request, return ok
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}
