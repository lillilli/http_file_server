package service

import (
	"github.com/lillilli/dr_web_exercise/src/config"
	"github.com/lillilli/dr_web_exercise/src/http"
	"github.com/lillilli/dr_web_exercise/src/types"
	logging "github.com/op/go-logging"
	"github.com/pkg/errors"
)

// Service - описание сервиса
type Service struct {
	cfg        *config.Config
	log        *logging.Logger
	httpServer types.Lifecycle
}

// New - создание нового сервиса
func New(cfg *config.Config) (service *Service, err error) {
	service = &Service{
		httpServer: http.NewServer(cfg),
		log:        logging.MustGetLogger("service"),
		cfg:        cfg,
	}

	return service, nil
}

// Start - запуск нового сервиса
func (s *Service) Start() error {
	s.log.Info("Starting...")

	if err := s.httpServer.Start(); err != nil {
		return errors.Wrap(err, "Http server starting failed")
	}

	s.log.Info("Starting... done")
	return nil
}

// Stop - остановка сервиса
// TODO: подумать насчёт выброса ошибки
func (s *Service) Stop() error {
	s.log.Info("Stopping...")

	if err := s.httpServer.Stop(); err != nil {
		return errors.Wrap(err, "Http server stoping failed")
	}

	s.log.Info("Stopping... done")
	return nil
}
