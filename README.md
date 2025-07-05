# Pismo Event Processor

This project demonstrates a resilient, reactive event processing pipeline designed to support multi-tenant platforms with multiple producers and client-specific delivery logic.

## Scenario

Imagine a stream of events from diverse producers such as:

- Monitoring systems
- User applications
- Transaction authorizers
- External integrations

Each event must:

- Be validated against a schema
- Include a `client_id` and `event_type`
- Be routed to a client-specific destination
- Be processed and delivered with low latency and high resilience

This project implements this pipeline using **SQS**, **DynamoDB**, and **DynamoDB Streams**, simulated locally with **LocalStack**.

---

## Architecture Overview

```text
+-------------+        +------------+        +----------------+        +------------------+
|  Producers  | ---->  |   SQS      | ---->  |   Processor     | ----> |  DynamoDB         |
|             |        |  (events)  |        |  (Validation +  |        |  (Persistent +    |
|             |        |            |        |   Storage)      |        |   Stream-enabled) |
+-------------+        +------------+        +----------------+        +--------+---------+
                                                                           |
                                                                           v
                                                                +------------------+
                                                                |   Deliverer       |
                                                                |  (Dynamo Stream   |
                                                                |   → HTTP targets) |
                                                                +------------------+
```

---

## Features and Requirements Coverage

| Requirement                                  | Status | How It’s Covered                           |
| -------------------------------------------- | ------ | ------------------------------------------ |
| **Multi-producer support**                   | Yes    | `event_type` distinguishes sources         |
| **Multi-tenant delivery**                    | Yes    | `client_id` identifies tenants             |
| **Client-specific routing**                  | Yes    | Routes stored in `routes` DynamoDB table   |
| **Validation of event contract**             | Yes    | JSON schema in `/schema/event_schema.json` |
| **Resilient event flow (no event loss)**     | Yes    | SQS as buffer, DynamoDB for persistence    |
| **Reactive stream-based processing**         | Yes    | Uses DynamoDB Streams to trigger delivery  |
| **Low-latency delivery**                     | Yes    | Stream to HTTP delivery without polling     |
| **No HTTP ingestion (no collector pattern)** | Yes    | Events are only consumed via SQS           |
| **Clear separation of concerns**             | Yes    | Separate processor and delivery services   |

---

## Components

### `/cmd/processor`

Consumes messages from SQS:

- Validates against JSON schema
- Stores valid events in DynamoDB (`Events` table)
- Optionally dispatches events immediately when `DIRECT_DISPATCH=true`

### `/cmd/deliver`

- Subscribes to DynamoDB stream
- Looks up target URL in `routes` table
- Forwards event to webhook via HTTP POST

---

## Getting Started

### Prerequisites

- Docker + Docker Compose
- Go >= 1.24
- aws-cli

### Setup

1. You may wish to adjust your end point at line 43 of ```scripts/localstack-init.sh```. I used a locally hosted http mock sink.

2. Start the environment
```docker-compose up --build```


## Setup Script Summary

This script performs the following:

- Creates the `events` queue
- Creates the `events-dlq` DLQ
- Creates the `Events` and `routes` DynamoDB tables
- Adds test route for `client-123`
- Sends a test event to SQS

## Project Structure

```
.
├── cmd
│   ├── deliver      # DynamoDB stream consumer and dispatcher
│   └── processor    # SQS consumer and validator
├── internal
│   ├── consumer     # SQS poller
│   ├── delivery     # HTTP dispatcher, stream logic
│   ├── model        # Event struct
│   ├── processor    # Core handler
│   ├── storage      # DynamoDB integration
│   ├── validator    # JSON schema validation
│   └── utils        # Helper functions
├── schema
│   └── event_schema.json
├── scripts
│   ├── localstack-init.sh      # Bootstrap script
│   ├── send-sqs-event.sh       # Manual test event injector
│   ├── event.json              # Event payload for send-sqs-event.sh
├── Dockerfile.*                # Separate Dockerfiles per service
├── docker-compose.yml          # Compose services (localstack, processor, deliver)
└── README.md
```

## Notes

- Only trusted producers should write to the SQS queue.
- Real deployments should include:
  - Authentication and authorization
  - Retry strategies and exponential backoff
  - Monitoring, tracing, and metrics for observability

