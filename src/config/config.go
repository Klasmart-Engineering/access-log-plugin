package config

import (
	"errors"
)

type Config struct {
	ProductName                string
	IgnoredPaths               IgnoredPaths
	ChannelBufferSize          int
	FirehoseBatchSize          int
	FirehoseSendEarlyTimeoutMs int
	AwsSecretKeyId             string
	AwsSecretKey               string
	AwsEndpoint                *string
	AwsRegion                  string
	DeliveryStreamName         string
}

type IgnoredPath string
type IgnoredPaths []IgnoredPath

func GetConfig(extra map[string]interface{}) (*Config, error) {
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

	ignoredPaths := make([]IgnoredPath, len(ignoredPathsRaw))
	for i, path := range ignoredPathsRaw {
		if _, isString := path.(string); !isString {
			return nil, errors.New("ignore_paths in access-log config map in krakend.json must contain only strings")
		}

		ignoredPaths[i] = IgnoredPath(path.(string))
	}

	var channelBufferSize int
	var channelBufferSizeRaw float64
	if channelBufferSizeRaw, ok = (extra["access-log"].(map[string]interface{})["buffer_size"]).(float64); !ok {
		channelBufferSizeRaw = 1000
	}

	channelBufferSize = int(channelBufferSizeRaw)

	var firehoseBatchSize int
	var firehoseBatchSizeRaw float64
	if firehoseBatchSizeRaw, ok = (extra["access-log"].(map[string]interface{})["firehose_batch_size"]).(float64); !ok {
		firehoseBatchSizeRaw = 500
	}

	firehoseBatchSize = int(firehoseBatchSizeRaw)

	var firehoseSendEarlyTimeoutMs int
	var firehoseSendEarlyTimeoutMsRaw float64
	if firehoseSendEarlyTimeoutMsRaw, ok = (extra["access-log"].(map[string]interface{})["firehose_send_early_timeout_ms"]).(float64); !ok {
		firehoseSendEarlyTimeoutMsRaw = 60000
	}

	firehoseSendEarlyTimeoutMs = int(firehoseSendEarlyTimeoutMsRaw)

	var awsSecretKeyId string
	if awsSecretKeyId, ok = (extra["access-log"].(map[string]interface{})["aws_secret_key_id"]).(string); !ok {
		return nil, errors.New("aws_secret_key_id in access-log config map in krakend.json must be a string")
	}

	var awsSecretKey string
	if awsSecretKey, ok = (extra["access-log"].(map[string]interface{})["aws_secret_key"]).(string); !ok {
		return nil, errors.New("aws_secret_key in access-log config map in krakend.json must be a string")
	}

	var awsEndpoint *string
	if _, exists := extra["access-log"].(map[string]interface{})["aws_endpoint"]; exists {
		var awsEndpointRaw string
		if awsEndpointRaw, ok = (extra["access-log"].(map[string]interface{})["aws_endpoint"]).(string); !ok {
			return nil, errors.New("aws_endpoint in access-log config map in krakend.json must be a string")
		}

		awsEndpoint = &awsEndpointRaw
	}

	var awsRegion string
	if awsRegion, ok = (extra["access-log"].(map[string]interface{})["aws_region"]).(string); !ok {
		return nil, errors.New("aws_region in access-log config map in krakend.json must be a string")
	}

	var deliveryStreamName string
	if deliveryStreamName, ok = (extra["access-log"].(map[string]interface{})["delivery_stream_name"]).(string); !ok {
		return nil, errors.New("delivery_stream_name in access-log config map in krakend.json must be a string")
	}

	return &Config{
		ProductName:                productName,
		IgnoredPaths:               ignoredPaths,
		ChannelBufferSize:          channelBufferSize,
		FirehoseBatchSize:          firehoseBatchSize,
		FirehoseSendEarlyTimeoutMs: firehoseSendEarlyTimeoutMs,
		AwsSecretKeyId:             awsSecretKeyId,
		AwsSecretKey:               awsSecretKey,
		AwsEndpoint:                awsEndpoint,
		AwsRegion:                  awsRegion,
		DeliveryStreamName:         deliveryStreamName,
	}, nil
}

func (ignoredPath IgnoredPath) matches(path string) bool {
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

		if (ignoredPathIndex == len(ignoredPath)-1 && ignoredPathChar != '*') || pathIndex == len(path)-1 {
			return false
		}

		if ignoredPathChar != '*' || (ignoredPathChar == '*' && canPeekPathChar && peekPathChar == '/') {
			if ignoredPathIndex == len(ignoredPath)-1 {
				return false
			}

			ignoredPathIndex += 1
		}

		pathIndex += 1
	}

}

func (ignoredPaths IgnoredPaths) AnyMatch(path string) bool {
	for _, ignoredPath := range ignoredPaths {
		if ignoredPath.matches(path) {
			return true
		}
	}

	return false
}
