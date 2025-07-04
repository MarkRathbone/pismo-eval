#!/bin/bash

set -e

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

# Config
REGION="eu-west-1"
ENDPOINT="http://localhost:4566"
QUEUE_NAME="events"
TABLE_NAME="Events"
QUEUE_URL="$ENDPOINT/000000000000/$QUEUE_NAME"

echo "Waiting for LocalStack to become ready..."

# Wait for LocalStack to be reachable
until curl -s "$ENDPOINT/_localstack/health" | grep '"sqs": "running"' > /dev/null; do
  sleep 2
done

echo "LocalStack is up."

# Create SQS queue
echo "Creating SQS queue: $QUEUE_NAME"
aws --endpoint-url=$ENDPOINT --region=$REGION sqs create-queue --queue-name $QUEUE_NAME

# Create DynamoDB table
echo "Creating DynamoDB table: $TABLE_NAME"
aws --endpoint-url=http://localhost:4566 --region=eu-west-1 dynamodb create-table \
  --table-name Events \
  --attribute-definitions AttributeName=client_id,AttributeType=S \
  --key-schema AttributeName=client_id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_IMAGE

# Inject test event
echo "Sending test event to SQS..."

aws --endpoint-url=$ENDPOINT --region=$REGION sqs send-message \
  --queue-url $QUEUE_URL \
  --message-body '{
    "client_id": "client-123",
    "event_type": "signup",
    "data": {
      "email": "test@example.com",
      "ip": "192.168.1.1"
    }
  }'

echo "Setup complete. Watch the processor logs for output."
