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

func getTargetURL(clientID string) (string, error) {
	dbClient, err := utils.NewDynamoDBClient(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to create DynamoDB client: %w", err)
	}

	result, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
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

func DispatchEvent(event model.Event) {
	target, err := getTargetURL(event.ClientID)
	if err != nil {
		log.Printf("Error fetching route for %s: %v", event.ClientID, err)
		return
	}
	if target == "" {
		log.Printf("No route for client_id: %s", event.ClientID)
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}

	resp, err := http.Post(target, "application/json", bytes.NewReader(payload))
	if err != nil {
		log.Printf("Failed to deliver event: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("Delivery failed with status: %d", resp.StatusCode)
		return
	}

	log.Printf("Event delivered to %s successfully", target)
}
