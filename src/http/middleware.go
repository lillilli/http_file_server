package http

// import (
// 	"context"
// 	"net/http"

// 	"gitlab2.sqtools.ru/go-pkg/logging"
// 	"gitlab2.sqtools.ru/go-pkg/sqauth"

// 	"gitlab2.sqtools.ru/go-pkg/metric"
// 	"github.com/lillilli/dr_web_exercise/src/config"
// 	"github.com/lillilli/dr_web_exercise/src/http/handler"
// )

// func newMiddleware(baseHandler *handler.BaseHandler, cfg *config.Config, collector metric.StatsdCollector) *middleware {
// 	sqauth.SetDebugMode(cfg.Debug)
// 	authMiddleware := sqauth.NewMiddleware(cfg.SQAuthProject)
// 	return &middleware{cfg, logging.NewLogger("middleware"), baseHandler, authMiddleware, collector}
// }

// type middleware struct {
// 	config         *config.Config
// 	log            logging.Logger
// 	baseHandler    *handler.BaseHandler
// 	authMiddleware *sqauth.Middleware
// 	collector      metric.StatsdCollector
// }

// func (m middleware) userAuthMiddleware(f http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
// 	tokenAuthMiddleware := m.tokenAuthMiddleware(f)
// 	h := m.authMiddleware.Populate(tokenAuthMiddleware, nil)

// 	return h
// }

// func (m middleware) tokenAuthMiddleware(f http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
// 	logMiddleware := m.logMiddleware(f)

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		if m.config.Debug || r.Method == http.MethodOptions {
// 			logMiddleware(w, r)
// 			return
// 		}

// 		token, ok := r.Header["Token"]
// 		if !ok {
// 			m.log.Debugf("Request - %s %s | Project auth token header required", r.Method, r.RequestURI)
// 			m.baseHandler.SendJSONError(r.Context(), w, http.StatusUnauthorized, "401 Unauthorized: Project auth token header required")
// 			return
// 		}

// 		for _, val := range m.config.Projects {
// 			if token[0] == val.Token {
// 				logMiddleware(w, r)
// 				return
// 			}
// 		}

// 		m.log.Debugf("Request - %s %s | Bad project auth token", r.Method, r.RequestURI)
// 		m.baseHandler.SendJSONError(r.Context(), w, http.StatusForbidden, "403 Forbidden: Bad project auth token")
// 	}
// }

// func (m middleware) logMiddleware(f http.HandlerFunc) func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		m.log.Debugf("Request - %s %s %s", r.Method, r.RequestURI, r.Proto)

// 		ctx := r.Context()
// 		ctx = context.WithValue(ctx, handler.RequestTimeContextKey, m.collector.StartTiming())
// 		r = r.WithContext(ctx)

// 		origin := r.Header.Get("Origin")
// 		w.Header().Set("Access-Control-Allow-Credentials", "true")
// 		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE")
// 		w.Header().Set("Access-Control-Allow-Origin", origin)
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Token")

// 		if r.Method == http.MethodOptions {
// 			w.Header().Set("Content-Type", "text/html")
// 			w.Write([]byte("ok"))
// 			return
// 		}

// 		f(w, r)
// 	}
// }
