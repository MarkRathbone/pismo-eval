#!/usr/bin/env sh
set -e

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

# config
REGION="${AWS_REGION:-eu-west-1}"
ENDPOINT="${AWS_ENDPOINT:-http://localstack:4566}"

# Names
DLQ_NAME="events-dlq"
QUEUE_NAME="events"
DELIVER_DLQ_NAME="deliver-dlq"
EVENTS_TABLE="Events"
ROUTES_TABLE="routes"

# 1) create the events DLQ
echo "Creating Dead-Letter Queue for events: $DLQ_NAME"
aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs create-queue --queue-name "$DLQ_NAME" >/dev/null

DLQ_URL=$(aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs get-queue-url --queue-name "$DLQ_NAME" --output text)
DLQ_ARN=$(aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs get-queue-attributes --queue-url "$DLQ_URL" \
    --attribute-names QueueArn \
    --query "Attributes.QueueArn" --output text)

echo "  events DLQ URL: $DLQ_URL"
echo "  events DLQ ARN: $DLQ_ARN"

# 2) create the main events queue, with that DLQ as its redrive target
echo "Creating main SQS queue: $QUEUE_NAME (redrive to $DLQ_NAME after 5 receives)"
RAW_POLICY=$(printf '{"deadLetterTargetArn":"%s","maxReceiveCount":%d}' "$DLQ_ARN" 5)
ESCAPED_POLICY=$(printf '%s' "$RAW_POLICY" | sed 's/"/\\"/g')

INPUT_JSON=$(printf \
  '{"QueueName":"%s","Attributes":{"RedrivePolicy":"%s"}}' \
  "$QUEUE_NAME" "$ESCAPED_POLICY")

aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs create-queue --cli-input-json "$INPUT_JSON"

# 3) create the “delivery” DLQ
echo "Creating Delivery Dead-Letter Queue: $DELIVER_DLQ_NAME"
aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs create-queue --queue-name "$DELIVER_DLQ_NAME" >/dev/null

DELIVER_DLQ_URL=$(aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs get-queue-url --queue-name "$DELIVER_DLQ_NAME" --output text)

echo "  deliver DLQ URL: $DELIVER_DLQ_URL"

# 4) DynamoDB tables
echo "Creating DynamoDB table: $EVENTS_TABLE"
aws --endpoint-url="$ENDPOINT" --region="$REGION" dynamodb create-table \
  --table-name "$EVENTS_TABLE" \
  --attribute-definitions \
      AttributeName=client_id,AttributeType=S \
      AttributeName=event_time,AttributeType=S \
  --key-schema \
      AttributeName=client_id,KeyType=HASH \
      AttributeName=event_time,KeyType=RANGE \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_IMAGE || true

echo "Creating DynamoDB table: $ROUTES_TABLE"
aws --endpoint-url="$ENDPOINT" --region="$REGION" dynamodb create-table \
  --table-name "$ROUTES_TABLE" \
  --attribute-definitions AttributeName=client_id,AttributeType=S \
  --key-schema AttributeName=client_id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 || true

echo "Seeding routes table with client_id ⇒ target_url"
aws --endpoint-url="$ENDPOINT" --region="$REGION" dynamodb put-item \
  --table-name "$ROUTES_TABLE" \
  --item '{
    "client_id": {"S": "client-123"},
    "target_url": {"S": "http://mock-sink:8080"}
  }'

echo "localstack-init complete."