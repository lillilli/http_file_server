package main

import (
	"flag"
	"fmt"
	"log"
	netHTTP "net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/logutils"
	"github.com/lillilli/http_file_server/config"
	"github.com/lillilli/http_file_server/fs"
	"github.com/lillilli/http_file_server/http/handler"
	"github.com/lillilli/vconf"
)

var (
	configFile = flag.String("config", "", "set service config file")
)

const (
	readTimeout  = time.Duration(5 * time.Second)
	writeTimeout = readTimeout
)

func main() {
	flag.Parse()

	cfg := &config.Config{}

	if err := vconf.InitFromFile(*configFile, cfg); err != nil {
		fmt.Printf("unable to load config: %s\n", err)
		os.Exit(1)
	}

	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "INFO", "ERROR"},
		MinLevel: logutils.LogLevel(cfg.Log.MinLevel),
		Writer:   os.Stdout,
	})

	startHTTPServer(cfg)
}

func startHTTPServer(cfg *config.Config) {
	router := mux.NewRouter()

	filesHander, err := handler.NewFileHandler(cfg.StaticDir, fs.NewStorage())
	if err != nil {
		log.Printf("[ERROR] Initializing file handler failed: %v", err)
		return
	}

	router.HandleFunc("/health", LogRequest(handler.HealthHandler)).Methods(netHTTP.MethodGet)
	router.HandleFunc("/download/{filehash}", LogRequest(filesHander.GetFile)).Methods(netHTTP.MethodGet)
	router.HandleFunc("/delete/{filehash}", LogRequest(filesHander.Remove)).Methods(netHTTP.MethodGet)
	router.HandleFunc("/upload", LogRequest(filesHander.Upload)).Methods(netHTTP.MethodPost)

	srv := netHTTP.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,

		Handler: router,
	}

	log.Printf("[INFO] Starting server on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

// LogRequest - log incoming requests
func LogRequest(handle func(w netHTTP.ResponseWriter, r *netHTTP.Request)) netHTTP.HandlerFunc {
	return func(w netHTTP.ResponseWriter, r *netHTTP.Request) {
		start := time.Now()
		log.Printf("[DENUG] Handling %s request %s", r.Method, r.URL.Path)
		handle(w, r)
		log.Printf("[DEBUG] Handle %s complete, handle time (nanoseconds): %d", r.URL.Path, time.Since(start).Nanoseconds())
	}
}
