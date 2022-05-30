package main

import (
	"context"
	"errors"
	"github.com/luraproject/lura/v2/logging"
	"net/http"
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

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, _ http.Handler) (http.Handler, error) {
	productName, err := getProductName(extra)
	if err != nil {
		return nil, err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Info(logPrefix, "Request to", productName, ":", req.URL)
	}), nil
}

func getProductName(extra map[string]interface{}) (string, error) {
	if _, ok := extra["access-log"]; !ok {
		return "", errors.New("access-log config map missing from krakend.json")
	}

	if _, isMap := extra["access-log"].(map[string]interface{}); !isMap {
		return "", errors.New("access-log config in krakend.json must be a map")
	}

	var productName string
	var ok bool
	if productName, ok = (extra["access-log"].(map[string]interface{})["product_name"]).(string); !ok {
		return "", errors.New("product_name in access-log config map in krakend.json must be a string")
	}

	return productName, nil
}
