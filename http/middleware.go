package http

import (
	"log"
	"net/http"
	"time"
)

// LogRequest - log incoming requests
func LogRequest(handle func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[DEBUG] log middleware: Handling %s request %s", r.Method, r.URL.Path)
		handle(w, r)
		log.Printf("[DEBUG] log middleware: Handle %s complete, handle time (nanoseconds): %d", r.URL.Path, time.Since(start).Nanoseconds())
	}
}
