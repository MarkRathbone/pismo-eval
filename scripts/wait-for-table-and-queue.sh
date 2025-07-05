#!/usr/bin/env sh
set -e

: "${AWS_ENDPOINT:?AWS_ENDPOINT must be set}"
: "${AWS_REGION:?AWS_REGION must be set}"
TABLE_NAME="${ROUTES_TABLE:-routes}"
QUEUE_NAME="${QUEUE_NAME:-events}"

echo "⏳ waiting for DynamoDB table \"$TABLE_NAME\"…"
until aws \
  --endpoint-url="$AWS_ENDPOINT" \
  --region="$AWS_REGION" \
  dynamodb describe-table \
    --table-name "$TABLE_NAME" > /dev/null 2>&1; do
  sleep 1
done
echo "table \"$TABLE_NAME\" is ready"

echo "⏳ waiting for SQS queue \"$QUEUE_NAME\"…"
until aws \
  --endpoint-url="$AWS_ENDPOINT" \
  --region="$AWS_REGION" \
  sqs get-queue-url \
    --queue-name "$QUEUE_NAME" > /dev/null 2>&1; do
  sleep 1
done

echo "queue \"$QUEUE_NAME\" is ready, launching app"
exec "$@"
