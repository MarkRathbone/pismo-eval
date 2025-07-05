package delivery

import (
	"context"
	"encoding/json"
	"event-processor/internal/model"
	"event-processor/internal/utils"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	stream "github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
)

func StartStreamProcessor() error {
	dbClient, err := utils.NewDynamoDBClient(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to create DynamoDB client: %w", err)
	}

	output, err := dbClient.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String("Events"),
	})
	if err != nil {
		log.Fatalf("Failed to describe table: %v", err)
	}

	streamArn := aws.ToString(output.Table.LatestStreamArn)
	if streamArn == "" {
		log.Fatal("Stream is not enabled on table or ARN is empty")
	}

	client, err := utils.NewDynamoDBStreamClient(context.TODO())
	if err != nil {
		return err
	}

	streamDesc, err := client.DescribeStream(context.TODO(), &dynamodbstreams.DescribeStreamInput{
		StreamArn: &streamArn,
	})
	if err != nil {
		return err
	}

	for _, shard := range streamDesc.StreamDescription.Shards {
		go processShard(client, streamArn, shard)
	}

	select {}
}

func processShard(client *dynamodbstreams.Client, streamArn string, shard stream.Shard) {
	iteratorOut, err := client.GetShardIterator(context.TODO(), &dynamodbstreams.GetShardIteratorInput{
		StreamArn:         &streamArn,
		ShardId:           shard.ShardId,
		ShardIteratorType: stream.ShardIteratorTypeTrimHorizon,
	})
	if err != nil {
		log.Println("Iterator error:", err)
		return
	}

	shardIterator := iteratorOut.ShardIterator

	for shardIterator != nil {
		out, err := client.GetRecords(context.TODO(), &dynamodbstreams.GetRecordsInput{
			ShardIterator: shardIterator,
		})
		if err != nil {
			log.Println("Record fetch error:", err)
			return
		}

		for _, record := range out.Records {
			if record.Dynamodb.NewImage == nil {
				continue
			}

			jsonBytes, err := json.Marshal(record.Dynamodb.NewImage)
			if err != nil {
				log.Println("Marshal error:", err)
				continue
			}

			var ddbImage map[string]ddb.AttributeValue
			err = json.Unmarshal(jsonBytes, &ddbImage)
			if err != nil {
				log.Println("Unmarshal error:", err)
				continue
			}

			var event model.Event
			err = attributevalue.UnmarshalMap(ddbImage, &event)
			if err != nil {
				log.Println("Attribute unmarshal error:", err)
				continue
			}

			go DispatchEvent(event)
		}

		shardIterator = out.NextShardIterator
		time.Sleep(2 * time.Second)
	}
}
