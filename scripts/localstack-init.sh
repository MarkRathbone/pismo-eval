#!/usr/bin/env bash
set -euo pipefail

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test

REGION="eu-west-1"
ENDPOINT="http://localstack:4566"
AWS_CLI="aws --endpoint-url=${ENDPOINT} --region=${REGION}"

# Queue names
MAIN_QUEUE_NAME="events"
DLQ_NAME="${MAIN_QUEUE_NAME}-dlq"

# DynamoDB tables
EVENTS_TABLE="Events"
ROUTES_TABLE="routes"

echo "Waiting for LocalStack…"
until curl -s ${ENDPOINT}/_localstack/health | grep -q '"sqs":.*"running"' \
   && grep -q '"dynamodb":.*"running"'; do
  sleep 1
done

echo "Creating Dead-Letter Queue: ${DLQ_NAME}"
DLQ_URL=$(${AWS_CLI} sqs create-queue \
  --queue-name "${DLQ_NAME}" \
  --query 'QueueUrl' --output text)
echo "  • DLQ URL: ${DLQ_URL}"

echo "Fetching DLQ ARN"
DLQ_ARN=$(${AWS_CLI} sqs get-queue-attributes \
  --queue-url "${DLQ_URL}" \
  --attribute-names QueueArn \
  --query 'Attributes.QueueArn' --output text)
echo "  • DLQ ARN: ${DLQ_ARN}"

echo "Creating main SQS queue: ${MAIN_QUEUE_NAME} with RedrivePolicy → ${DLQ_NAME} after 5 receives"
REDRIVE_POLICY=$(
  jq -c -n \
    --arg target "${DLQ_ARN}" \
    --argjson max 5 \
    '{deadLetterTargetArn: $target, maxReceiveCount: $max}'
)
MAIN_QUEUE_URL=$(${AWS_CLI} sqs create-queue \
  --queue-name "${MAIN_QUEUE_NAME}" \
  --attributes RedrivePolicy="${REDRIVE_POLICY}" \
  --query 'QueueUrl' --output text)
echo "  • Main queue URL: ${MAIN_QUEUE_URL}"
echo

echo "Setting QUEUE_URL environment variable for downstream steps"
export QUEUE_URL="${MAIN_QUEUE_URL}"
echo "  • QUEUE_URL=${QUEUE_URL}"
echo

echo "Creating DynamoDB table: ${EVENTS_TABLE}"
${AWS_CLI} dynamodb create-table \
  --table-name "${EVENTS_TABLE}" \
  --attribute-definitions AttributeName=client_id,AttributeType=S \
  --key-schema AttributeName=client_id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --stream-specification StreamEnabled=true,StreamViewType=NEW_IMAGE || true
echo

echo "Creating DynamoDB table: ${ROUTES_TABLE}"
${AWS_CLI} dynamodb create-table \
  --table-name "${ROUTES_TABLE}" \
  --attribute-definitions AttributeName=client_id,AttributeType=S \
  --key-schema AttributeName=client_id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 || true
echo

echo "Seeding routes table with client_id → target_url"
${AWS_CLI} dynamodb put-item \
  --table-name "${ROUTES_TABLE}" \
  --item '{
    "client_id": {"S": "client-123"},
    "target_url": {"S": "https://webhook.site/c59a9948-c50f-4f07-8451-3a38c6d81276"}
  }'
echo

echo "Sending test event to SQS (${MAIN_QUEUE_NAME})"
${AWS_CLI} sqs send-message \
  --queue-url "${QUEUE_URL}" \
  --message-body '{
    "client_id": "client-123",
    "event_type": "signup",
    "data": {
      "email": "test@example.com",
      "ip": "192.168.1.1"
    }
  }'
echo

echo "Setup complete. Watch processor logs for output."
