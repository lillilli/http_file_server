package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"gitlab2.sqtools.ru/go-pkg/logging"
	"gopkg.in/alexcesaro/statsd.v2"

	"gitlab2.sqtools.ru/go-pkg/metric"
	"github.com/lillilli/dr_web_exercise/src/config"
	"github.com/lillilli/dr_web_exercise/src/types"
)

const (
	RequestTimeContextKey = "requestTime"
	RequestTypeContextKey = "requestType"

	GetFileRequestType     = "getFile"
	UploadFileRequestType  = "uploadFile"
	GetManifestRequestType = "getManifest"
)

// BaseHandler - структура базового обработчика запросов
type BaseHandler struct {
	log       logging.Logger
	collector metric.StatsdCollector
	cfg       *config.Config
}

// New - создание базового нового обработчика
func New(collector metric.StatsdCollector, cfg *config.Config) *BaseHandler {
	return &BaseHandler{logging.NewLogger("http handler"), collector, cfg}
}

// ParseReqParams - преобразование тела запроса в структуру
func (h BaseHandler) ParseReqParams(req *http.Request, v interface{}) error {
	decoder := json.NewDecoder(req.Body)
	return decoder.Decode(v)
}

// SendInternalServerError - отправка на клиент сообщения о внутренней ошибке
func (h BaseHandler) SendInternalServerError(ctx context.Context, w http.ResponseWriter, msg string, err error) {
	h.log.Errorf("internal server error: %s : %v", msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))

	reqType, _ := ctx.Value(RequestTypeContextKey).(string)
	h.sendCountByRequestType(reqType, 1, false)
}

// SendBadRequestError - отправка на клиент сообщения о ошибке запроса
func (h BaseHandler) SendBadRequestError(ctx context.Context, w http.ResponseWriter, msg string, err error) {
	h.log.Infof("bad request error: %s, error: %v", msg, err)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(msg))

	reqType, _ := ctx.Value(RequestTypeContextKey).(string)
	h.sendCountByRequestType(reqType, 1, false)
}

// SendJSONResponse - отправка на клиент json сообщения
func (h BaseHandler) SendJSONResponse(ctx context.Context, w http.ResponseWriter, result interface{}) {
	b, err := json.Marshal(result)
	if err != nil {
		h.SendInternalServerError(ctx, w, "can`t marshal json result", err)
		return
	}

	h.log.Infof("send json result: %s", b)
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)

	h.finishSuccessMetric(ctx)
}

// SendJSONError - отправка на клиент json сообщения об ошибке
func (h BaseHandler) SendJSONError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	b, err := json.Marshal(types.Error{Message: message})
	if err != nil {
		h.SendInternalServerError(ctx, w, "can`t marshal json result", err)
		return
	}

	h.log.Infof("send json error result: %s", b)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(b)

	reqType, _ := ctx.Value(RequestTypeContextKey).(string)
	h.sendCountByRequestType(reqType, 1, false)
}

// Send404 - отправка 404 статуса
func (h BaseHandler) Send404(ctx context.Context, w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))

	reqType, _ := ctx.Value(RequestTypeContextKey).(string)
	h.sendCountByRequestType(reqType, 1, false)
}

// SendOK - отправка ok сообщения клиенту
func (h BaseHandler) SendOK(ctx context.Context, w http.ResponseWriter) {
	w.Write([]byte("ok"))

	h.finishSuccessMetric(ctx)
}

func (h BaseHandler) finishSuccessMetric(ctx context.Context) {
	reqType, reqTypeOk := ctx.Value(RequestTypeContextKey).(string)
	timing, timingOk := ctx.Value(RequestTimeContextKey).(statsd.Timing)
	if !reqTypeOk || !timingOk {
		return
	}
	h.sendCountByRequestType(reqType, 1, true)
	h.sendTimingByRequestType(reqType, timing)
}

// SetRequestType - утсанавливает тип запроса в контекст запроса
func (h BaseHandler) SetRequestType(r *http.Request, reqType string) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, RequestTypeContextKey, reqType)
	return r.WithContext(ctx)
}

func (h BaseHandler) sendCountByRequestType(reqType string, value interface{}, success bool) {
	switch reqType {
	case GetFileRequestType:
		if success {
			h.collector.Send(h.cfg.Stats.Metrics.Requests.GetFile.Success.RequestCount, value)
		} else {
			h.collector.Send(h.cfg.Stats.Metrics.Requests.GetFile.Error.Count, value)
		}
	case UploadFileRequestType:
		if success {
			h.collector.Send(h.cfg.Stats.Metrics.Requests.UploadFile.Success.RequestCount, value)
		} else {
			h.collector.Send(h.cfg.Stats.Metrics.Requests.UploadFile.Error.Count, value)
		}
	case GetManifestRequestType:
		if success {
			h.collector.Send(h.cfg.Stats.Metrics.Requests.GetManifest.Success.RequestCount, value)
		} else {
			h.collector.Send(h.cfg.Stats.Metrics.Requests.GetManifest.Error.Count, value)
		}
	}
}

func (h BaseHandler) sendTimingByRequestType(reqType string, timing statsd.Timing) {
	switch reqType {
	case GetFileRequestType:
		h.collector.Send(h.cfg.Stats.Metrics.Requests.GetFile.Success.RequestTime, timing)
	case UploadFileRequestType:
		h.collector.Send(h.cfg.Stats.Metrics.Requests.UploadFile.Success.RequestTime, timing)
	case GetManifestRequestType:
		h.collector.Send(h.cfg.Stats.Metrics.Requests.GetManifest.Success.RequestTime, timing)
	}
}
