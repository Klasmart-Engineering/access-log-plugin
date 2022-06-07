package helper

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	firehoseTypes "github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/cenkalti/backoff/v4"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

var firehoseClient *firehose.Client
var s3Client *s3.Client

func setupAWS(t *testing.T) {
	if firehoseClient != nil && s3Client != nil {
		return
	}

	creds := credentials.NewStaticCredentialsProvider(
		"test",
		"test",
		"")

	cfg := aws.Config{
		Credentials: creds,
		Region:      "eu-west-1",
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           "http://localhost:4566",
				SigningRegion: region,
			}, nil
		}),
		EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			//Despite being deprecated, it seems this is actually still used sometimes
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           "http://localhost:4566",
				SigningRegion: region,
			}, nil
		}),
	}

	firehoseClient = firehose.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func WaitForHealthcheck(t *testing.T) {
	var check = func() error {
		_, err := http.Get("http://localhost:8080/__health")
		if err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         1 * time.Second,
		MaxElapsedTime:      30 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatal(err)
	}
}

func ResetLocalstack(t *testing.T) {
	setupAWS(t)

	streamName := "factory-access-logs"

	_, err := firehoseClient.DeleteDeliveryStream(context.Background(), &firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: &streamName,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "not found.") {
			t.Fatalf("Could not delete delivery stream %s", err)
		}
	}

	bucketName := "factory-access-log-bucket"

	objects := listObjectsOrEmpty(t, bucketName)

	for _, object := range objects {
		_, err = s3Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: &bucketName,
			Key:    &object,
		})
		if err != nil {
			t.Fatalf("Could not delete object %s: %s", object, err)
		}
	}

	_, err = s3Client.DeleteBucket(context.Background(), &s3.DeleteBucketInput{
		Bucket: &bucketName,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "does not exist") {
			t.Fatalf("Could not delete bucket %s", err)
		}
	}

	_, err = s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: &bucketName,
		ACL:    "public-read",
		CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
			LocationConstraint: "eu-west-1",
		},
	})
	if err != nil {
		t.Fatalf("Could not create bucket %s", err)
	}

	deliveryStreamName := "factory-access-logs"
	bucketArn := "arn:aws:s3:::factory-access-log-bucket"
	roleArn := "roleArn := arn:aws:iam::000000000000:role/super-role"
	intervalSeconds := int32(60)
	sizeInMBs := int32(1)
	accessLogsErrorPrefix := "access-logs-error"

	_, err = firehoseClient.CreateDeliveryStream(context.Background(), &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: &deliveryStreamName,
		DeliveryStreamType: "DirectPut",
		ExtendedS3DestinationConfiguration: &firehoseTypes.ExtendedS3DestinationConfiguration{
			BucketARN: &bucketArn,
			RoleARN:   &roleArn,
			BufferingHints: &firehoseTypes.BufferingHints{
				IntervalInSeconds: &intervalSeconds,
				SizeInMBs:         &sizeInMBs,
			},
			ErrorOutputPrefix: &accessLogsErrorPrefix,
		},
	})
	if err != nil {
		t.Fatalf("Could not create delivery stream %s", err)
	}
}

func WaitForLocalstack(t *testing.T) {
	waitForFirehose(t)
	waitForS3(t)
}

func waitForFirehose(t *testing.T) {
	setupAWS(t)

	var check = func() (returnedError error) {
		defer func() {
			err := recover()
			if err != nil {
				returnedError = err.(error)
			}
		}()

		_, err := firehoseClient.ListDeliveryStreams(context.Background(), &firehose.ListDeliveryStreamsInput{})
		if err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         20 * time.Second,
		MaxElapsedTime:      120 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatalf("Timed out waiting for Firehose: %s", err)
	}
}

func waitForS3(t *testing.T) {
	var check = func() (returnedError error) {
		defer func() {
			err := recover()
			if err != nil {
				returnedError = err.(error)
			}
		}()

		_, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
		if err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(check, &backoff.ExponentialBackOff{
		InitialInterval:     100 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          1.2,
		MaxInterval:         20 * time.Second,
		MaxElapsedTime:      120 * time.Second,
		Stop:                -1,
		Clock:               backoff.SystemClock,
	})

	if err != nil {
		t.Fatalf("Timed out waiting for S3: %s", err)
	}
}

func ListObjects(t *testing.T, bucketName string) []string {
	output, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		t.Fatalf("Could not list objects in bucket: %s", err)
	}

	objectKeys := make([]string, output.KeyCount)
	for i, item := range output.Contents {
		objectKeys[i] = *item.Key
	}

	return objectKeys
}

func listObjectsOrEmpty(t *testing.T, bucketName string) []string {
	output, err := s3Client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		return make([]string, 0)
	}

	objectKeys := make([]string, output.KeyCount)
	for i, item := range output.Contents {
		objectKeys[i] = *item.Key
	}

	return objectKeys
}

func GetObjectContent(t *testing.T, bucketName string, objectKey string) []byte {
	object, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		t.Fatalf("Could not get object content: %s", err)
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, object.Body)
	if err != nil {
		t.Fatalf("Could not read object content: %s", err)
	}

	return buf.Bytes()
}
