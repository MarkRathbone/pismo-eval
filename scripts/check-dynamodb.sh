#!/usr/bin/env bash
set -e

echo "Scanning DynamoDB table: Events"
AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
  aws --endpoint-url=http://localhost:4566 --region eu-west-1 \
    dynamodb scan --table-name Events

echo
echo "Scanning DynamoDB table: routes"
AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
  aws --endpoint-url=http://localhost:4566 --region eu-west-1 \
    dynamodb scan --table-name routes
