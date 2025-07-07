package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"event-processor/internal/consumer"
	"event-processor/internal/processor"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func main() {
	log.Println("Starting Event Processor...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutdown signal received, canceling...")
		cancel()
	}()

	endpoint := os.Getenv("AWS_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	if endpoint == "" || region == "" {
		log.Fatal("Both AWS_ENDPOINT and AWS_REGION must be set")
	}
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	var queueURL string
	if envURL := os.Getenv("QUEUE_URL"); envURL != "" {
		queueURL = envURL
	} else {
		name := os.Getenv("QUEUE_NAME")
		if name == "" {
			name = "events"
		}
		qOut, err := sqsClient.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
			QueueName: aws.String(name),
		})
		if err != nil {
			log.Fatalf("Unable to get queue URL for %s: %v", name, err)
		}
		queueURL = aws.ToString(qOut.QueueUrl)
	}
	log.Printf("Consuming from %s", queueURL)

	if err := consumer.StartConsumer(ctx, sqsClient, queueURL, processor.HandleMessage); err != nil {
		log.Printf("Error running consumer: %v", err)
	}

	log.Println("Event Processor shut down cleanly.")
}
