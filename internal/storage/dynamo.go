package storage

import (
	"context"
	"event-processor/internal/model"
	"event-processor/internal/utils"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var dbClient *dynamodb.Client

func init() {
	client, err := utils.NewDynamoDBClient(context.TODO())
	if err != nil {
		log.Fatalf("unable to create DynamoDB client: %v", err)
	}
	dbClient = client
}

func SaveEvent(event model.Event) error {
	item, err := attributevalue.MarshalMap(map[string]interface{}{
		"client_id":  event.ClientID,
		"event_type": event.EventType,
		"data":       string(event.Data),
	})
	if err != nil {
		return err
	}

	_, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(utils.TableName),
		Item:      item,
	})

	return err
}
