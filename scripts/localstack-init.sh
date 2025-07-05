#!/usr/bin/env sh
set -e

# credentials (fallback to "test" if unset)
export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-test}
export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-test}

# config
REGION="${AWS_REGION:-eu-west-1}"
ENDPOINT="${AWS_ENDPOINT:-http://localstack:4566}"

DLQ_NAME="events-dlq"
QUEUE_NAME="events"
EVENTS_TABLE="Events"
ROUTES_TABLE="routes"

echo "Creating Dead-Letter Queue: $DLQ_NAME"
aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs create-queue --queue-name "$DLQ_NAME" >/dev/null

# fetch DLQ URL & ARN
DLQ_URL=$(aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs get-queue-url --queue-name "$DLQ_NAME" --output text)
DLQ_ARN=$(aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs get-queue-attributes \
      --queue-url "$DLQ_URL" \
      --attribute-names QueueArn \
      --query "Attributes.QueueArn" \
      --output text)

echo "  DLQ URL: $DLQ_URL"
echo "  DLQ ARN: $DLQ_ARN"

echo "Creating main SQS queue: $QUEUE_NAME (redrive to $DLQ_NAME after 5 receives)"
RAW_POLICY=$(printf '{"deadLetterTargetArn":"%s","maxReceiveCount":%d}' "$DLQ_ARN" 5)
ESCAPED_POLICY=$(printf '%s' "$RAW_POLICY" | sed 's/"/\\"/g')

INPUT_JSON=$(printf \
  '{"QueueName":"%s","Attributes":{"RedrivePolicy":"%s"}}' \
  "$QUEUE_NAME" "$ESCAPED_POLICY")

aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs create-queue --cli-input-json "$INPUT_JSON"

echo "Creating DynamoDB table: $EVENTS_TABLE"
aws --endpoint-url="$ENDPOINT" --region="$REGION" dynamodb create-table \
  --table-name "$EVENTS_TABLE" \
  --attribute-definitions AttributeName=client_id,AttributeType=S \
  --key-schema AttributeName=client_id,KeyType=HASH \
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

echo "Sending test event to SQS…"
QUEUE_URL=$(aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs get-queue-url --queue-name "$QUEUE_NAME" --output text)
aws --endpoint-url="$ENDPOINT" --region="$REGION" sqs send-message \
  --queue-url "$QUEUE_URL" \
  --message-body '{
    "client_id": "client-123",
    "event_type": "signup",
    "data": {
      "email": "test@example.com",
      "ip": "192.168.1.1"
    }
  }'

echo "localstack-init complete."
