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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var routesDBClient *dynamodb.Client

func init() {
	client, err := utils.NewDynamoDBClient(context.TODO())
	if err != nil {
		log.Fatalf("failed to create DynamoDB client: %v", err)
	}
	routesDBClient = client
}

func getTargetURL(clientID string) (string, error) {
	if routesDBClient == nil {
		return "", fmt.Errorf("dynamodb client not initialized")
	}

	result, err := routesDBClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("routes"),
		Key: map[string]types.AttributeValue{
			"client_id": &types.AttributeValueMemberS{Value: clientID},
		},
	})
	if err != nil {
		return "", err
	}
	if result.Item == nil {
		return "", nil
	}

	if val, ok := result.Item["target_url"].(*types.AttributeValueMemberS); ok {
		return val.Value, nil
	}
	return "", nil
}

func DispatchEvent(event model.Event) error {
	target, err := getTargetURL(event.ClientID)
	if err != nil {
		return fmt.Errorf("fetch route for client_id %s: %w", event.ClientID, err)
	}
	if target == "" {
		return fmt.Errorf("no route configured for client_id %s", event.ClientID)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	resp, err := http.Post(target, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("http post to %s: %w", target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dispatch returned status %d for client_id %s", resp.StatusCode, event.ClientID)
	}

	return nil
}
