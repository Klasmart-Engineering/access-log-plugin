package main

import (
	"context"
	uuid2 "github.com/google/uuid"
	"github.com/luraproject/lura/v2/logging"
	"net/http"
	"time"
)

const logPrefix = "[PLUGIN:ACCESS-LOG]"

var logger logging.Logger = nil

func init() {

}

func (r registerer) RegisterLogger(v interface{}) {
	l, ok := v.(logging.Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(logPrefix, HandlerRegisterer, "access-log plugin loaded")
}

var HandlerRegisterer = registerer("access-log")

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, handler http.Handler) (http.Handler, error) {
	config, err := getConfig(extra)
	if err != nil {
		return nil, err
	}

	accesses := make(chan AccessLog, config.channelBufferSize)
	go firehoseSync(config, accesses)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if config.ignoredPaths.anyMatch(req.URL.Path) {
			logger.Debug(logPrefix, "Ignoring request to ", req.URL, "ignored path")
			handler.ServeHTTP(w, req)
			return
		}

		accesses <- AccessLog{
			Id:             uuid2.New(),
			OccurredAt:     time.Now().Unix(),
			Product:        config.productName,
			Method:         req.Method,
			Path:           req.URL.Path,
			AndroidId:      "TODO - GET FROM JWT",
			SubscriptionId: uuid2.UUID{}, //TODO: GET FROM JWT too
		}

		handler.ServeHTTP(w, req)
	}), nil
}
