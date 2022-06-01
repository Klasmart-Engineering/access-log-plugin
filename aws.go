package main

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
)

var FirehoseClient *firehose.Client

func setupAWS(config *config) {
	creds := credentials.NewStaticCredentialsProvider(
		config.awsSecretKeyId,
		config.awsSecretKey,
		"")

	var cfg aws.Config
	if config.awsEndpoint != nil {
		cfg = aws.Config{
			Credentials: creds,
			Region:      config.awsRegion,
			EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				logger.Debug(logPrefix, "Returning AWS Endpoint (w options)", config.awsEndpoint, "for region", region)
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           *config.awsEndpoint,
					SigningRegion: region,
				}, nil
			}),
			EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				//Despite being deprecated, it seems this is actually still used sometimes
				logger.Debug(logPrefix, "Returning AWS Endpoint", config.awsEndpoint, "for region", region)
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           *config.awsEndpoint,
					SigningRegion: region,
				}, nil
			}),
		}
	} else {
		cfg = aws.Config{
			Credentials: creds,
			Region:      config.awsRegion,
		}
	}

	FirehoseClient = firehose.NewFromConfig(cfg)
}
