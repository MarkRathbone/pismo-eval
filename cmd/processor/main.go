package main

import (
	"context"
	"event-processor/internal/consumer"
	"event-processor/internal/processor"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	log.Println("Starting Event Processor...")

	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	queueName := "events"

	if endpoint == "" || region == "" {
		if endpoint == "" {
			log.Println("Missing required environment variable: AWS_ENDPOINT")
		}
		if region == "" {
			log.Println("Missing required environment variable: AWS_REGION")
		}
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	queueOutput, err := sqsClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		log.Fatalf("Unable to get queue URL for %s: %v", queueName, err)
	}
	queueURL := aws.ToString(queueOutput.QueueUrl)

	consumer.StartConsumer(sqsClient, queueURL, processor.HandleMessage)
}
