package delivery

import (
	"context"
	"log"
	"time"

	"event-processor/internal/model"
	"event-processor/internal/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	stream "github.com/aws/aws-sdk-go-v2/service/dynamodbstreams"
	streamTypes "github.com/aws/aws-sdk-go-v2/service/dynamodbstreams/types"
	"github.com/cenkalti/backoff/v4"
	"golang.org/x/sync/errgroup"
)

func StartStreamProcessor(ctx context.Context) error {
	ddbClient, err := utils.NewDynamoDBClient(ctx)
	if err != nil {
		return err
	}
	descOut, err := ddbClient.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String("Events"),
	})
	if err != nil {
		return err
	}
	streamArn := aws.ToString(descOut.Table.LatestStreamArn)

	streamClient, err := utils.NewDynamoDBStreamClient(ctx)
	if err != nil {
		return err
	}
	sDesc, err := streamClient.DescribeStream(ctx, &stream.DescribeStreamInput{
		StreamArn: aws.String(streamArn),
	})
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	for _, shard := range sDesc.StreamDescription.Shards {
		shard := shard
		g.Go(func() error {
			return processShard(ctx, streamClient, streamArn, shard, g)
		})
	}

	if err := g.Wait(); err != nil {
		if err == context.Canceled {
			log.Println("Delivery: shutdown complete")
			return nil
		}
		return err
	}
	log.Println("Delivery: all shards and dispatches completed")
	return nil
}

func processShard(
	ctx context.Context,
	client *stream.Client,
	streamArn string,
	shard streamTypes.Shard,
	g *errgroup.Group,
) error {
	shardID := aws.ToString(shard.ShardId)

	iterOut, err := client.GetShardIterator(ctx, &stream.GetShardIteratorInput{
		StreamArn:         aws.String(streamArn),
		ShardId:           shard.ShardId,
		ShardIteratorType: streamTypes.ShardIteratorTypeTrimHorizon,
	})
	if err != nil {
		return err
	}
	sharder := iterOut.ShardIterator

	for sharder != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var recordsOut *stream.GetRecordsOutput
		bo := backoff.NewExponentialBackOff()
		bo.MaxElapsedTime = time.Minute
		if err := backoff.Retry(func() error {
			var err error
			recordsOut, err = client.GetRecords(ctx, &stream.GetRecordsInput{
				ShardIterator: sharder,
			})
			return err
		}, bo); err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			log.Printf("Shard %s: persistent GetRecords error: %v", shardID, err)
			time.Sleep(10 * time.Second)
			continue
		}

		for _, rec := range recordsOut.Records {
			if rec.Dynamodb.NewImage == nil {
				continue
			}
			img := rec.Dynamodb.NewImage
			rawClient, ok1 := img["client_id"].(*streamTypes.AttributeValueMemberS)
			rawType, ok2 := img["event_type"].(*streamTypes.AttributeValueMemberS)
			rawData, ok3 := img["data"].(*streamTypes.AttributeValueMemberS)
			if !ok1 || !ok2 || !ok3 {
				continue
			}

			evt := model.Event{
				ClientID:  rawClient.Value,
				EventType: rawType.Value,
				Data:      []byte(rawData.Value),
			}

			g.Go(func() error {
				if err := DispatchEvent(ctx, evt); err != nil {
					log.Printf("Shard %s: DispatchEvent error: %v", shardID, err)
				}
				return nil
			})
		}

		sharder = recordsOut.NextShardIterator

		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	log.Printf("Shard %s: drained, exiting", shardID)
	return nil
}
