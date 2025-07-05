#!/bin/bash

set -e

echo "Waiting for LocalStack to be ready..."

until curl -s http://localstack:4566/health | grep '"dynamodb": "running"' > /dev/null; do
  sleep 1
done

# Wait until Kinesis stream becomes ACTIVE
echo "Waiting for Kinesis stream to become ACTIVE..."
while [[ "$(aws --endpoint-url=http://localstack:4566 kinesis describe-stream --stream-name __ddb_stream_Events | jq -r .StreamDescription.StreamStatus)" != "ACTIVE" ]]; do
  sleep 1
done

echo "LocalStack and stream are ready."
