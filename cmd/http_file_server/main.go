package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/lillilli/http_file_server/config"
	"github.com/lillilli/http_file_server/http"
	"github.com/lillilli/vconf"
	"github.com/pkg/errors"
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

	if err := startHTTPServer(cfg); err != nil {
		log.Fatalln(err)
	}
}

func startHTTPServer(cfg *config.Config) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	server := http.NewServer(cfg)

	if err := server.Start(); err != nil {
		return errors.Wrap(err, "unable to start http server")
	}

	<-signals
	close(signals)

	return server.Stop()
}
