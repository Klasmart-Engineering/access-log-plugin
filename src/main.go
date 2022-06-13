package main

import (
	"access-log/src/aws"
	"access-log/src/config"
	"access-log/src/firehose"
	"access-log/src/jwt"
	"access-log/src/logging"
	"context"
	uuid2 "github.com/google/uuid"
	lura "github.com/luraproject/lura/v2/logging"
	"net/http"
	"time"
)

func init() {

}

func (r registerer) RegisterLogger(v interface{}) {
	l, ok := v.(lura.Logger)
	if !ok {
		return
	}
	logging.SetLogger(l)
	logging.Debug(HandlerRegisterer, "access-log plugin loaded")
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
	configuration, err := config.GetConfig(extra)
	if err != nil {
		return nil, err
	}

	aws.SetupAWS(configuration)

	accesses := make(chan firehose.AccessLog, configuration.ChannelBufferSize)
	go firehose.FirehoseSync(configuration, accesses)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if configuration.IgnoredPaths.AnyMatch(req.URL.Path) {
			logging.Debug("Ignoring request to ", req.URL, "ignored path")
			handler.ServeHTTP(w, req)
			return
		}

		oAuth2ServiceJwt, err := jwt.GetOAuth2ServiceJwt(req.Header.Values("Authorization"))
		if err != nil {
			logging.Error("Unable to get claims from Authorization header", err)
			w.WriteHeader(401)
			return
		}

		if oAuth2ServiceJwt.SubscriptionId == nil {
			logging.Error("Subscription ID not present in JWT")
			w.WriteHeader(403)
			return
		}

		if oAuth2ServiceJwt.AndroidId == nil {
			logging.Error("Android ID not present in JWT")
			w.WriteHeader(403)
			return
		}

		accesses <- firehose.AccessLog{
			Id:             uuid2.New(),
			OccurredAt:     time.Now().Unix(),
			Product:        configuration.ProductName,
			Method:         req.Method,
			Path:           req.URL.Path,
			AndroidId:      *oAuth2ServiceJwt.AndroidId,
			SubscriptionId: *oAuth2ServiceJwt.SubscriptionId,
		}

		handler.ServeHTTP(w, req)
	}), nil
}
