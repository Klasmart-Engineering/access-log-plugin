package aws

import (
	"access-log/src/config"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
)

var FirehoseClient *firehose.Client

func SetupAWS(config *config.Config) {
	if !config.UseAwsDefaultCredentials {
		setupWithManuallyProvidedConfig(config)
		return
	}

	setupWithDefaults(config)
}

func setupWithManuallyProvidedConfig(config *config.Config) {
	creds := credentials.NewStaticCredentialsProvider(
		*config.AwsSecretKeyId,
		*config.AwsSecretKey,
		"")

	var cfg aws.Config
	if config.AwsEndpoint != nil {
		cfg = aws.Config{
			Credentials: creds,
			Region:      config.AwsRegion,
			EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           *config.AwsEndpoint,
					SigningRegion: region,
				}, nil
			}),
			EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
				//Despite being deprecated, it seems this is actually still used sometimes
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           *config.AwsEndpoint,
					SigningRegion: region,
				}, nil
			}),
		}
	} else {
		cfg = aws.Config{
			Credentials: creds,
			Region:      config.AwsRegion,
		}
	}

	FirehoseClient = firehose.NewFromConfig(cfg)
}

func setupWithDefaults(config *config.Config) {
	configuration, err := cfg.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	FirehoseClient = firehose.NewFromConfig(configuration, func(o *firehose.Options) {
		o.Region = config.AwsRegion
	})
}
