package main

import (
	"errors"
)

type config struct {
	productName                string
	ignoredPaths               ignoredPaths
	channelBufferSize          int
	firehoseBatchSize          int
	firehoseSendEarlyTimeoutMs int
}

type ignoredPath string
type ignoredPaths []ignoredPath

func getConfig(extra map[string]interface{}) (*config, error) {
	if _, exists := extra["access-log"]; !exists {
		return nil, errors.New("access-log config map missing from krakend.json")
	}

	if _, isMap := extra["access-log"].(map[string]interface{}); !isMap {
		return nil, errors.New("access-log config in krakend.json must be a map")
	}

	var productName string
	var ok bool
	if productName, ok = (extra["access-log"].(map[string]interface{})["product_name"]).(string); !ok {
		return nil, errors.New("product_name in access-log config map in krakend.json must be a string")
	}

	var ignoredPathsRaw []interface{}
	if ignoredPathsRaw, ok = (extra["access-log"].(map[string]interface{})["ignore_paths"]).([]interface{}); !ok {
		return nil, errors.New("ignore_paths in access-log config map in krakend.json must be an array")
	}

	ignoredPaths := make([]ignoredPath, len(ignoredPathsRaw))
	for i, path := range ignoredPathsRaw {
		if _, isString := path.(string); !isString {
			return nil, errors.New("ignore_paths in access-log config map in krakend.json must contain only strings")
		}

		ignoredPaths[i] = ignoredPath(path.(string))
	}

	var channelBufferSize int
	var channelBufferSizeRaw float64
	if channelBufferSizeRaw, ok = (extra["access-log"].(map[string]interface{})["buffer_size"]).(float64); !ok {
		return nil, errors.New("buffer_size in access-log config map in krakend.json must be an int")
	}

	channelBufferSize = int(channelBufferSizeRaw)

	var firehoseBatchSize int
	var firehoseBatchSizeRaw float64
	if firehoseBatchSizeRaw, ok = (extra["access-log"].(map[string]interface{})["firehose_batch_size"]).(float64); !ok {
		return nil, errors.New("firehose_batch_size in access-log config map in krakend.json must be an int")
	}

	firehoseBatchSize = int(firehoseBatchSizeRaw)

	var firehoseSendEarlyTimeoutMs int
	var firehoseSendEarlyTimeoutMsRaw float64
	if firehoseSendEarlyTimeoutMsRaw, ok = (extra["access-log"].(map[string]interface{})["firehose_send_early_timeout_ms"]).(float64); !ok {
		return nil, errors.New("firehose_send_early_timeout_ms in access-log config map in krakend.json must be an int")
	}

	firehoseSendEarlyTimeoutMs = int(firehoseSendEarlyTimeoutMsRaw)

	return &config{
		productName:                productName,
		ignoredPaths:               ignoredPaths,
		channelBufferSize:          channelBufferSize,
		firehoseBatchSize:          firehoseBatchSize,
		firehoseSendEarlyTimeoutMs: firehoseSendEarlyTimeoutMs,
	}, nil
}

func (ignoredPath ignoredPath) matches(path string) bool {
	var ignoredPathIndex, pathIndex int
	for {
		ignoredPathChar := ignoredPath[ignoredPathIndex]
		pathChar := path[pathIndex]
		var peekPathChar uint8
		canPeekPathChar := pathIndex != len(path)-1
		if canPeekPathChar {
			peekPathChar = path[pathIndex+1]
		}

		if ignoredPathChar != pathChar && ignoredPathChar != '*' {
			return false
		}

		if ignoredPathIndex == len(ignoredPath)-1 && pathIndex == len(path)-1 {
			return true
		}

		if (ignoredPathIndex == len(ignoredPath)-1) || (pathIndex == len(path)-1) {
			return false
		}

		if ignoredPathChar != '*' || (ignoredPathChar == '*' && canPeekPathChar && peekPathChar == '/') {
			ignoredPathIndex += 1
		}

		pathIndex += 1
	}

}

func (ignoredPaths ignoredPaths) anyMatch(path string) bool {
	for _, ignoredPath := range ignoredPaths {
		if ignoredPath.matches(path) {
			return true
		}
	}

	return false
}
