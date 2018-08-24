package http

import (
	"net/http"

	"github.com/lillilli/http_file_server/src/http/handler"
	"github.com/lillilli/http_file_server/src/http/handler/common"
	"github.com/lillilli/http_file_server/src/http/handler/files"
	"github.com/lillilli/http_file_server/src/http/modifiers"
)

func (s *Server) declareRoutes() error {
	baseHandler := handler.New(s.config)
	commonHandler := common.New()

	filesHander, err := files.New(baseHandler, s.config)
	if err != nil {
		return err
	}

	s.mux.HandleFunc("/health", commonHandler.Health).Methods(http.MethodGet)

	s.mux.HandleFunc("/download/{filehash}", filesHander.GetFile).Methods(http.MethodGet)
	s.mux.HandleFunc("/delete/{filehash}", filesHander.Remove).Methods(http.MethodGet)

	s.mux.HandleFunc("/upload", filesHander.Upload()).Methods(http.MethodPost)

	// примеры обработчиков загрузки файлов с pre и post обработкой
	s.mux.HandleFunc(
		"/dublicate_data_upload",
		filesHander.Upload(files.PreUploadModifier(modifiers.DublicateFileData)),
	).Methods(http.MethodPost)

	s.mux.HandleFunc(
		"/dublicate_hash_upload",
		filesHander.Upload(files.PostUploadModifier(modifiers.DublicateFileName)),
	).Methods(http.MethodPost)

	s.mux.HandleFunc(
		"/dublicate_data_and_hash_upload",
		filesHander.Upload(files.PreUploadModifier(modifiers.DublicateFileData), files.PostUploadModifier(modifiers.DublicateFileName)),
	).Methods(http.MethodPost)

	return nil
}
