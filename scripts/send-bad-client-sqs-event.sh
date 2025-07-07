#!/bin/bash
set -e

echo "Sending bad client event to SQS..."

AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
aws --endpoint-url=http://localhost:4566 --region eu-west-1 sqs send-message \
  --queue-url http://localhost:4566/000000000000/events \
  --message-body file://sqs-bad-client-event.json

