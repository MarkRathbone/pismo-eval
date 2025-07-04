#!/bin/bash
set -e

host="localstack"
port=4566

echo "Waiting for LocalStack at $host:$port..."

while ! nc -z $host $port; do
  sleep 1
done

echo "LocalStack is up. Starting app..."
exec ./main
