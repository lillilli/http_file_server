package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"os/signal"

	"github.com/lillilli/http_file_server/src/config"
	"github.com/lillilli/http_file_server/src/service"
	"github.com/lillilli/vconf"
)

var (
	configFile = flag.String("config", "", "set service config file")
)

func main() {
	flag.Parse()

	cfg := &config.Config{}

	if err := vconf.InitFromFile(*configFile, cfg); err != nil {
		fmt.Printf("unable to load config: %s\n", err)
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	service, err := service.New(cfg)
	if err != nil {
		fmt.Printf("unable to create service: %s\n", err)
		os.Exit(1)
	}

	if err = service.Start(); err != nil {
		fmt.Printf("unable to start service: %s\n", err)
		os.Exit(1)
	}

	<-signals
	close(signals)

	service.Stop()
}
