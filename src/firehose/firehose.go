package firehose

import (
	"access-log/src/aws"
	"access-log/src/config"
	"access-log/src/logging"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	uuid2 "github.com/google/uuid"
)

type AccessLog struct {
	Id             uuid2.UUID `json:"id"`
	OccurredAt     int64      `json:"occurred_at"`
	Product        string     `json:"product"`
	Method         string     `json:"method"`
	Path           string     `json:"path"`
	AndroidId      string     `json:"android_id"`
	SubscriptionId uuid2.UUID `json:"subscription_id"`
}

func FirehoseSync(config *config.Config, accesses chan AccessLog) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT)

	batch := make([]AccessLog, config.FirehoseBatchSize)
	batchCursor := 0

	sendBatch := func() {
		if batchCursor == 0 {
			return
		}

		sendBatchToFirehose(config.DeliveryStreamName, batch[:batchCursor])
		batchCursor = 0
	}

	processLoop := func() (shuttingDown bool) {
		defer func() {
			err := recover()
			if err != nil {
				logging.Error("Panic during firehose sync loop", err)
			}
		}()

		select {
		case accessLog := <-accesses:
			logging.Debug("Received access log", accessLog)
			batch[batchCursor] = accessLog
			batchCursor += 1

			if batchCursor == config.FirehoseBatchSize {
				sendBatch()
			}
		case <-time.After(time.Duration(config.FirehoseSendEarlyTimeoutMs) * time.Millisecond):
			sendBatch()
		case <-exit:
			logging.Debug("Krakend shutting down, sending batch")
			sendBatch()
			return true
		}

		return false
	}

	for {
		if processLoop() {
			return
		}
	}
}

func sendBatchToFirehose(deliveryStreamName string, batch []AccessLog) {
	logging.Debug(fmt.Sprintf("Sending batch of %d to Firehose", len(batch)))

	records := make([]types.Record, len(batch))
	for i, batchEntry := range batch {
		serialised, err := json.Marshal(batchEntry)
		if err != nil {
			logging.Error("Failed to serialise batch entry", batchEntry, err)
		}

		records[i] = types.Record{
			Data: serialised,
		}
	}

	output, err := aws.FirehoseClient.PutRecordBatch(context.Background(), &firehose.PutRecordBatchInput{
		DeliveryStreamName: &deliveryStreamName,
		Records:            records,
	})

	if err != nil {
		logging.Error("Failed to send batch to Firehose, err.Error()", err)
		return
	}

	if *output.FailedPutCount > int32(0) {
		logging.Error(fmt.Sprintf("Failed to send %d of %d batch entries", output.FailedPutCount, len(batch)))
	}

	return
}
