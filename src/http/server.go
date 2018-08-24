package http

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lillilli/http_file_server/src/config"
	"github.com/lillilli/http_file_server/src/types"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
)

// Server - структура http сервера
type Server struct {
	config *config.Config
	server *http.Server
	log    *logging.Logger
	mux    *mux.Router
}

// NewServer - создание нового http сервера
func NewServer(cfg *config.Config) types.Lifecycle {
	api := &Server{
		config: cfg,
		mux:    mux.NewRouter(),
		log:    logging.MustGetLogger("http server"),
	}

	api.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:        api.mux,
		ReadTimeout:    100 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return api
}

// Start - запускает http сервер
func (s *Server) Start() error {
	s.log.Info("Starting...")

	if err := s.declareRoutes(); err != nil {
		return errors.Wrap(err, "unable to declare routes")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", s.server.Addr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to get address %s", s.server.Addr))
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to start listening %s", tcpAddr))
	}

	go s.server.Serve(listener)

	s.log.Infof("Listening on %s", s.server.Addr)
	return nil
}

// Stop - выключение http серера
func (s *Server) Stop() error {
	s.log.Info("Stopping...")
	return s.server.Shutdown(nil)
}
