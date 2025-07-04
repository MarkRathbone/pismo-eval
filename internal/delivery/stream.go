package delivery

import (
	"context"
	"encoding/json"
	"event-processor/internal/model"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	"github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
)

func StartStreamProcessor() error {
	streamArn := os.Getenv("DYNAMO_STREAM_ARN")
	if streamArn == "" {
		return ErrMissingStreamARN
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
		config.WithEndpointResolverWithOptions(LocalResolver()),
		config.WithCredentialsProvider(aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("AWS_ACCESS_KEY_ID"),
				os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"",
			),
		)),
	)
	if err != nil {
		return err
	}

	client := dynamodbstreams.NewFromConfig(cfg)

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

func processShard(client *dynamodbstreams.Client, streamArn string, shard types.Shard) {
	iteratorOut, err := client.GetShardIterator(context.TODO(), &dynamodbstreams.GetShardIteratorInput{
		StreamArn:         &streamArn,
		ShardId:           shard.ShardId,
		ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
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
			var event model.Event
			err := attributevalue.UnmarshalMap(record.Dynamodb.NewImage, &event)
			if err != nil {
				log.Println("Unmarshal error:", err)
				continue
			}
			go DispatchEvent(event)
		}

		shardIterator = out.NextShardIterator
		time.Sleep(2 * time.Second)
	}
}

var ErrMissingStreamARN = fmt.Errorf("DYNAMO_STREAM_ARN is not set")
