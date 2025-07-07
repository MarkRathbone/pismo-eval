#!/usr/bin/env bash
set -e

DLQ_URL="http://localhost:4566/000000000000/events-dlq"

AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
  aws --endpoint-url=http://localhost:4566 --region eu-west-1 \
    sqs get-queue-attributes \
      --queue-url "$DLQ_URL" \
      --attribute-names \
        ApproximateNumberOfMessages \
        ApproximateNumberOfMessagesNotVisible \
      --output table
