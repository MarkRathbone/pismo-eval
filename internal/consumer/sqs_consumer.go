package consumer

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type handlerFunc func(ctx context.Context, payload string) error

func StartConsumer(ctx context.Context, sqsClient *sqs.Client, queueURL string, handler handlerFunc) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer: context canceled, exiting StartConsumer")
			return nil
		default:
		}

		out, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     5,
		})
		if err != nil {
			log.Println("Receive error:", err)
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return nil
			}
			continue
		}

		for _, msg := range out.Messages {
			if err := handler(ctx, *msg.Body); err != nil {
				log.Printf("handler error for message %q: %v", *msg.MessageId, err)
				continue
			}

			select {
			case <-ctx.Done():
				log.Println("Consumer: context canceled before delete, exiting")
				return nil
			default:
			}
			if _, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				log.Println("Delete error:", err)
			}
		}

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			log.Println("Consumer: context canceled during throttle, exiting")
			return nil
		}
	}
}
