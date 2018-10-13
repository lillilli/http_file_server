package http

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lillilli/http_file_server/config"
	"github.com/lillilli/http_file_server/fs"
	"github.com/lillilli/http_file_server/http/handler"
	"github.com/pkg/errors"
)

const (
	readTimeout  = time.Duration(5 * time.Second)
	writeTimeout = readTimeout
)

// Server - http server interface
type Server interface {
	Start() error
	Stop() error
	Address() string
}

// server - http sever structure
type server struct {
	cfg    *config.Config
	server *http.Server
	mux    *mux.Router
}

// NewServer - return new instance of http server
func NewServer(cfg *config.Config) Server {
	api := &server{
		cfg: cfg,
		mux: mux.NewRouter(),
	}

	api.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      api.mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	return api
}

// Start - start http server
func (s server) Start() error {
	log.Printf("[INFO] http server: Starting...")

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

	log.Printf("[INFO] http server: Listening on %s", s.server.Addr)
	return nil
}

func (s server) declareRoutes() error {
	filesHander, err := handler.NewFileHandler(s.cfg.StaticDir, fs.NewStorage())
	if err != nil {
		return errors.Wrap(err, "initializing file handler failed")
	}

	s.mux.HandleFunc("/health", LogRequest(handler.HealthHandler)).Methods(http.MethodGet)
	s.mux.HandleFunc("/download/{filehash}", LogRequest(filesHander.GetFile)).Methods(http.MethodGet)
	s.mux.HandleFunc("/delete/{filehash}", LogRequest(filesHander.Remove)).Methods(http.MethodGet)
	s.mux.HandleFunc("/upload", LogRequest(filesHander.Upload)).Methods(http.MethodPost)

	return nil
}

// Stop - shutdown http server
func (s server) Stop() error {
	log.Printf("[INFO] http server: Stopping...")

	err := s.server.Shutdown(nil)
	return err
}

// Address - return server address
func (s server) Address() string {
	return s.server.Addr
}
