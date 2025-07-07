#!/usr/bin/env bash
set -e

ENDPOINT="http://localhost:4566"
DLQ_ARN="arn:aws:sqs:eu-west-1:000000000000:events-dlq"

aws --endpoint-url="$ENDPOINT" --region="$REGION" \
    sqs start-message-move-task \
      --source-arn "$DLQ_ARN" \
      --max-number-of-messages-per-second 100

echo "DLQ redrive task started (events-dlq to events)."
