package main

import (
	uuid2 "github.com/google/uuid"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type AccessLog struct {
	Id             uuid2.UUID
	OccurredAt     int64
	Product        string
	Method         string
	Path           string
	AndroidId      string
	SubscriptionId uuid2.UUID
}

func firehoseSync(config *config, accesses chan AccessLog) {
	//TODO: panic handler - should just start loop again

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT)

	batch := make([]AccessLog, config.firehoseBatchSize)
	batchCursor := 0

	sendBatch := func() {
		if batchCursor == 0 {
			logger.Debug(logPrefix, "Nothing in batch to send")
			return
		}

		err := sendBatchToFirehose(batch[:batchCursor])
		if err != nil {
			logger.Error(logPrefix, "Unable to send access log batch to Firehose - dropping", err)
		}
		batchCursor = 0
	}

	for {
		select {
		case accessLog := <-accesses:
			logger.Debug(logPrefix, "Received access log", accessLog)
			batch[batchCursor] = accessLog
			batchCursor += 1

			if batchCursor == config.firehoseBatchSize {
				sendBatch()
			}
		case <-time.After(time.Duration(config.firehoseSendEarlyTimeoutMs) * time.Millisecond):
			logger.Debug(logPrefix, "Sending batch early as send early timeout exceeded")
			sendBatch()
		case <-exit:
			logger.Debug(logPrefix, "Krakend shutting down, sending batch")
			sendBatch()
			return
		}
	}
}

func sendBatchToFirehose(batch []AccessLog) error {
	//TODO
	logger.Debug("Sending batch to Firehose")
	for i, batchItem := range batch {
		logger.Info("Batch element", i, ":", batchItem)
	}

	return nil
}
