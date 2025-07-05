package consumer

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func StartConsumer(sqsClient *sqs.Client, queueURL string, handler func(string)) {
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
