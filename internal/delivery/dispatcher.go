package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"event-processor/internal/model"
	"event-processor/internal/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	routesDBClient *dynamodb.Client
	sqsClient      *sqs.Client
	httpClient     *http.Client
	deliverDLQURL  string
)

func init() {
	db, err := utils.NewDynamoDBClient(context.Background())
	if err != nil {
		log.Fatalf("failed to create DynamoDB client: %v", err)
	}
	routesDBClient = db

	sqsC, err := utils.NewSQSClient(context.Background())
	if err != nil {
		log.Fatalf("failed to create SQS client: %v", err)
	}
	sqsClient = sqsC

	deliverDLQURL = os.Getenv("DELIVER_DLQ_URL")
	if deliverDLQURL == "" {
		log.Fatal("DELIVER_DLQ_URL must be set")
	}

	httpClient = &http.Client{Timeout: 5 * time.Second}
}

func getTargetURL(clientID string) (string, error) {
	out, err := routesDBClient.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("routes"),
		Key: map[string]types.AttributeValue{
			"client_id": &types.AttributeValueMemberS{Value: clientID},
		},
	})
	if err != nil {
		return "", err
	}
	if out.Item == nil {
		return "", nil
	}
	if v, ok := out.Item["target_url"].(*types.AttributeValueMemberS); ok {
		return v.Value, nil
	}
	return "", nil
}

func DispatchEvent(ctx context.Context, evt model.Event) error {
	target, err := getTargetURL(evt.ClientID)
	if err != nil {
		log.Printf("route lookup failed for %q: %v", evt.ClientID, err)
		return sendToDLQ(evt)
	}
	if target == "" {
		log.Printf("no route for client_id %q", evt.ClientID)
		return sendToDLQ(evt)
	}

	b, err := json.Marshal(evt)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return sendToDLQ(evt)
	}

	for i := 1; i <= 3; i++ {
		req, _ := http.NewRequestWithContext(ctx, "POST", target, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			log.Printf("delivered event for %q to %s", evt.ClientID, target)
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		log.Printf("delivery attempt %d for %q failed: %v", i, evt.ClientID, err)
		time.Sleep(time.Duration(i) * time.Second)
	}

	log.Printf("all retries exhausted for %q, pushing to deliver DLQ", evt.ClientID)
	return sendToDLQBytes(b)
}

func sendToDLQ(evt model.Event) error {
	b, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("marshal for DLQ: %w", err)
	}
	return sendToDLQBytes(b)
}

func sendToDLQBytes(b []byte) error {
	_, err := sqsClient.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(deliverDLQURL),
		MessageBody: aws.String(string(b)),
	})
	if err != nil {
		return fmt.Errorf("enqueue to DLQ: %w", err)
	}
	return fmt.Errorf("event pushed to DLQ")
}
