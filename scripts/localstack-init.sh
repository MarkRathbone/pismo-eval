#!/bin/bash

set -e

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

# Config
REGION="eu-west-1"
ENDPOINT="http://localstack:4566"
QUEUE_NAME="events"
EVENTS_TABLE="Events"
ROUTES_TABLE="routes"
QUEUE_URL="$ENDPOINT/000000000000/$QUEUE_NAME"

# Create SQS queue
echo "Creating SQS queue: $QUEUE_NAME"
aws --endpoint-url=$ENDPOINT --region=$REGION sqs create-queue --queue-name $QUEUE_NAME

# Create DynamoDB table: Events
echo "Creating DynamoDB table: $EVENTS_TABLE"
aws --endpoint-url=$ENDPOINT --region=$REGION dynamodb create-table \
  --table-name $EVENTS_TABLE \
  --attribute-definitions AttributeName=client_id,AttributeType=S \
  --key-schema AttributeName=client_id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_IMAGE || true

# Create DynamoDB table: routes
echo "Creating DynamoDB table: $ROUTES_TABLE"
aws --endpoint-url=$ENDPOINT --region=$REGION dynamodb create-table \
  --table-name $ROUTES_TABLE \
  --attribute-definitions AttributeName=client_id,AttributeType=S \
  --key-schema AttributeName=client_id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 || true

# Seed routes table
echo "Seeding routes table with client_id => target_url"
aws --endpoint-url=$ENDPOINT --region=$REGION dynamodb put-item \
  --table-name $ROUTES_TABLE \
  --item '{
    "client_id": {"S": "client-123"},
    "target_url": {"S": "https://webhook.site/c59a9948-c50f-4f07-8451-3a38c6d81276"}
  }'

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
