package main

import (
	"context"
	"log"
	"os"

	"event-processor/internal/consumer"
	"event-processor/internal/processor"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	log.Println("Starting Event Processor...")

	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	if endpoint == "" || region == "" {
		log.Fatal("Both AWS_ENDPOINT and AWS_REGION must be set")
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

	var queueURL string
	if envURL := os.Getenv("QUEUE_URL"); envURL != "" {
		queueURL = envURL
		log.Printf("Using QUEUE_URL from env: %s", queueURL)
	} else {
		queueName := os.Getenv("QUEUE_NAME")
		if queueName == "" {
			queueName = "events"
		}
		qOut, err := sqsClient.GetQueueUrl(context.TODO(), &sqs.GetQueueUrlInput{
			QueueName: aws.String(queueName),
		})
		if err != nil {
			log.Fatalf("Unable to get queue URL for %s: %v", queueName, err)
		}
		queueURL = aws.ToString(qOut.QueueUrl)
		log.Printf("Resolved queue URL for %s: %s", queueName, queueURL)
	}

	consumer.StartConsumer(sqsClient, queueURL, processor.HandleMessage)
}
