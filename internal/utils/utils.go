package utils

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

const TableName = "Events"

func NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
	)
	if err != nil {
		return nil, err
	}

	customEndpoint := os.Getenv("AWS_ENDPOINT")
	if customEndpoint != "" {
		return dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(customEndpoint)
		}), nil
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func NewSQSClient(ctx context.Context) (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
	)
	if err != nil {
		return nil, err
	}

	customEndpoint := os.Getenv("AWS_ENDPOINT")
	if customEndpoint != "" {
		return sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(customEndpoint)
		}), nil
	}

	return sqs.NewFromConfig(cfg), nil
}

func NewDynamoDBStreamClient(ctx context.Context) (*dynamodbstreams.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
	)
	if err != nil {
		return nil, err
	}

	customEndpoint := os.Getenv("AWS_ENDPOINT")
	if customEndpoint != "" {
		return dynamodbstreams.NewFromConfig(cfg, func(o *dynamodbstreams.Options) {
			o.BaseEndpoint = aws.String(customEndpoint)
		}), nil
	}

	return dynamodbstreams.NewFromConfig(cfg), nil
}
