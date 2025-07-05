package consumer

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func StartConsumer(sqsClient *sqs.Client, queueURL string, handler func(string) error) {
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
			if err := handler(*msg.Body); err != nil {
				log.Printf("Handler error for message %s: %v", *msg.MessageId, err)
				continue
			}

			_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Printf("Delete error for message %s: %v", *msg.MessageId, err)
			}
		}

		time.Sleep(1 * time.Second)
	}
}
