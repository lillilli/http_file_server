package http

import (
	"net/http"

	"github.com/lillilli/dr_web_exercise/src/http/handler/common"
)

// import (
// 	"fmt"
// 	"net/http"

// 	"github.com/pkg/errors"

// 	"github.com/lillilli/dr_web_exercise/src/api/git"
// 	"github.com/lillilli/dr_web_exercise/src/config"
// 	"github.com/lillilli/dr_web_exercise/src/fs"
// 	"github.com/lillilli/dr_web_exercise/src/http/handler"
// 	"github.com/lillilli/dr_web_exercise/src/http/handler/admin"
// 	"github.com/lillilli/dr_web_exercise/src/http/handler/arbitrary"
// 	"github.com/lillilli/dr_web_exercise/src/http/handler/common"
// 	handlerTypes "github.com/lillilli/dr_web_exercise/src/http/handler/types"
// 	"github.com/lillilli/dr_web_exercise/src/http/handler/versioned"
// )

func (s *Server) declareRoutes() error {
	// baseHandler := handler.New(s.collector, s.config)
	// m := newMiddleware(baseHandler, s.config, s.collector)
	commonHandler := common.New()
	// adminHandler := admin.New(baseHandler, s.config, s.db)

	// fs, err := checkFs(s.config)
	// if err != nil {
	// 	return err
	// }

	// git := git.New()
	// versionedHandler := versioned.New(baseHandler, s.config, fs, git, s.collector)
	// if err != nil {
	// 	return err
	// }

	// arbitraryHandler := arbitrary.New(baseHandler, s.config, fs, git)
	// if err != nil {
	// 	return err
	// }

	s.mux.HandleFunc("/health", commonHandler.Health).Methods(http.MethodGet)

	// s.mux.HandleFunc("/admin/auth", m.logMiddleware(adminHandler.TokenAuth)).Methods(http.MethodGet)
	// s.mux.HandleFunc("/admin/token", m.userAuthMiddleware(adminHandler.GetTokens)).Methods(http.MethodGet, http.MethodOptions)
	// s.mux.HandleFunc("/admin/token", m.userAuthMiddleware(adminHandler.AddToken)).Methods(http.MethodPost, http.MethodOptions)
	// s.mux.HandleFunc("/admin/counter", m.userAuthMiddleware(adminHandler.AddCounter)).Methods(http.MethodPost, http.MethodOptions)

	// s.mux.HandleFunc("/api/upload/{project}/{filename:.*}", m.tokenAuthMiddleware(arbitraryHandler.Upload)).Methods(http.MethodPost, http.MethodOptions)
	// s.mux.HandleFunc("/api/update/{project}/{filename:.*}", m.tokenAuthMiddleware(arbitraryHandler.Update)).Methods(http.MethodPost, http.MethodOptions)
	// s.mux.HandleFunc("/api/update-from-git/{project}", m.tokenAuthMiddleware(arbitraryHandler.UpdateFromGit)).Methods(http.MethodPost, http.MethodOptions)
	// s.mux.HandleFunc("/api/files/{project}", m.tokenAuthMiddleware(arbitraryHandler.List)).Methods(http.MethodGet, http.MethodOptions)
	// s.mux.HandleFunc("/api/list-static/{project}", m.tokenAuthMiddleware(arbitraryHandler.List)).Methods(http.MethodGet, http.MethodOptions)
	// s.mux.HandleFunc("/api/remove/{project}/{filename:.*}", m.tokenAuthMiddleware(arbitraryHandler.Remove)).Methods(http.MethodDelete, http.MethodOptions)

	// s.mux.HandleFunc("/api/manifest/{project}/{tag}", m.tokenAuthMiddleware(versionedHandler.GetManifest)).Methods(http.MethodGet, http.MethodOptions)
	// s.mux.HandleFunc("/api/warmup/{project}", m.tokenAuthMiddleware(versionedHandler.WarmUp)).Methods(http.MethodPost, http.MethodOptions)

	// s.mux.HandleFunc("/{project}/{filepath:.*}", m.tokenAuthMiddleware(versionedHandler.GetUncachedFile)).Methods(http.MethodGet)
	return nil
}

// // FIXME move somewhere else: this is not a router functionality!
// func checkFs(config *config.Config) (handlerTypes.Fs, error) {
// 	fs := fs.New()

// 	if err := checkDir(fs, config.CacheDir); err != nil {
// 		return fs, errors.Wrap(err, "bad CacheDir")
// 	}

// 	if err := checkDir(fs, config.StaticDir); err != nil {
// 		return fs, errors.Wrap(err, "bad StaticDir")
// 	}

// 	return fs, nil
// }

// func checkDir(fs handlerTypes.Fs, dir string) error {
// 	exists, err := fs.DirExists(dir)
// 	if err != nil {
// 		return err
// 	}

// 	if !exists {
// 		return fmt.Errorf(`directory %s does not exist`, dir)
// 	}

// 	return nil
// }
