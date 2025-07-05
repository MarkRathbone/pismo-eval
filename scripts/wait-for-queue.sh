#!/usr/bin/env sh
set -e

: "${AWS_ENDPOINT:?AWS_ENDPOINT must be set}"
: "${AWS_REGION:?AWS_REGION must be set}"
QUEUE_NAME="${QUEUE_NAME:-events}"

echo "waiting for SQS queue \"$QUEUE_NAME\"…"
until aws \
  --endpoint-url="$AWS_ENDPOINT" \
  --region="$AWS_REGION" \
  sqs get-queue-url \
    --queue-name "$QUEUE_NAME" > /dev/null 2>&1; do
  sleep 1
done

echo "queue \"$QUEUE_NAME\" is ready, launching app"
exec "$@"
