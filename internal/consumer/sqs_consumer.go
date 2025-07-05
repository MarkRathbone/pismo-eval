package consumer

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type handlerFunc func(string) error

func StartConsumer(sqsClient *sqs.Client, queueURL string, handler handlerFunc) {
	for {
		out, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     5,
		})
		if err != nil {
			log.Println("Receive error:", err)
			time.Sleep(time.Second)
			continue
		}

		for _, msg := range out.Messages {
			if err := handler(*msg.Body); err != nil {
				log.Printf("handler error for message %q: %v", *msg.MessageId, err)
				continue
			}

			_, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
			if err != nil {
				log.Println("Delete error:", err)
			}
		}

		time.Sleep(1 * time.Second)
	}
}
