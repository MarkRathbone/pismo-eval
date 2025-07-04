package utils

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var TableName = "Events"

func LocalResolver() aws.EndpointResolverWithOptions {
	endpoint := os.Getenv("AWS_ENDPOINT")
	signingRegion := os.Getenv("AWS_REGION")

	return aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpoint,
			SigningRegion: signingRegion,
		}, nil
	})
}
