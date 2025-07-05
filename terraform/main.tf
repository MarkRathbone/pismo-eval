terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.0"
    }
    null = {
      source  = "hashicorp/null"
      version = ">= 3.0"
    }
  }
}

variable "aws_region" {
  type    = string
  default = "eu-west-1"
}

variable "localstack_endpoint" {
  type    = string
  default = "http://localhost:4566"
}

variable "dlq_name" {
  type    = string
  default = "events-dlq"
}

variable "queue_name" {
  type    = string
  default = "events"
}

variable "events_table" {
  type    = string
  default = "Events"
}

variable "routes_table" {
  type    = string
  default = "routes"
}

provider "aws" {
  region     = var.aws_region
  access_key = "test"
  secret_key = "test"

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    sqs      = var.localstack_endpoint
    dynamodb = var.localstack_endpoint
  }
}


resource "aws_sqs_queue" "dlq" {
  name = var.dlq_name
}

resource "aws_sqs_queue" "main" {
  name = var.queue_name

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 5
  })
}

resource "aws_dynamodb_table" "events" {
  name         = var.events_table
  billing_mode = "PROVISIONED"

  hash_key       = "client_id"
  range_key      = "event_type"
  read_capacity  = 5
  write_capacity = 5

  attribute {
    name = "client_id"
    type = "S"
  }
  attribute {
    name = "event_type"
    type = "S"
  }

  stream_enabled   = true
  stream_view_type = "NEW_IMAGE"
}

resource "aws_dynamodb_table" "routes" {
  name         = var.routes_table
  billing_mode = "PROVISIONED"

  hash_key       = "client_id"
  read_capacity  = 5
  write_capacity = 5

  attribute {
    name = "client_id"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "route_seed" {
  table_name = aws_dynamodb_table.routes.name
  hash_key   = "client_id"
  item = jsonencode({
    client_id  = { S = "client-123" }
    target_url = { S = "http://mock-sink:8080" }
  })
}

resource "null_resource" "send_test_event" {
  triggers = {
    queue_url = aws_sqs_queue.main.id
  }

  provisioner "local-exec" {
    command = <<EOT
aws --endpoint-url="${var.localstack_endpoint}" \
    --region="${var.aws_region}" \
    sqs send-message \
      --queue-url="${aws_sqs_queue.main.id}" \
      --message-body '{
        "client_id": "client-123",
        "event_type": "signup",
        "data": {
          "email": "test@example.com",
          "ip": "192.168.1.1"
        }
      }'
EOT
  }
}
