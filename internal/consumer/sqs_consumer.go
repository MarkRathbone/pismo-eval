package consumer

import (
	"context"
	"event-processor/internal/utils"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	queueURL  = "http://localstack:4566/000000000000/events"
	sqsClient *sqs.Client
)

func StartConsumer(handler func(string)) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithEndpointResolverWithOptions(utils.LocalResolver()),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_ACCESS_KEY_ID"),
				os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"",
			),
		)),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	sqsClient = sqs.NewFromConfig(cfg)

	for {
		out, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     5,
		})
		if err != nil {
			log.Println("Receive error:", err)
			continue
		}

		for _, msg := range out.Messages {
			handler(*msg.Body)
		}
		time.Sleep(1 * time.Second)
	}
}
