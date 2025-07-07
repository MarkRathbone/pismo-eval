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

Although beyond the scope of this task, I also developed a simple deliverer service, delivering data to a http sink. 

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
- Moves events to a DLQ for further inspection on failed delivery

---

## Getting Started

### Prerequisites

- Docker + Docker Compose
- Go
- aws-cli

### Setup

1. You may wish to adjust your end point in the ```scripts/localstack-init.sh```. I used a locally hosted http mock sink.

2. Start the environment
```docker compose up --build```

## Setup Script Summary

This localstack-init script performs the following:

- Creates the `events-dlq` dead-letter queue  
- Creates the `deliver-dlq` dead-letter queue
- Creates the `events` main SQS queue (with redrive policy to `events-dlq` after 5 receives)  
- Creates the `Events` DynamoDB table, keyed by `client_id` (HASH) and `event_time` (RANGE), with a stream on new images  
- Creates the `routes` DynamoDB table, keyed by `client_id`  
- Seeds the `routes` table with a test mapping for `client-123 to http://mock-sink:8080`.

### Usage

Once the deployment is ready, you can start using the deployment. The send-*.sh scripts in the scripts folder will perform a number of sample event sends to the SQS. 

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
├── localstack-scripts
│   ├── localstack-init.sh                # Bootstrap script
│   ├── wait-for-queue.sh                 # Waits for the queue to be ready
│   ├── wait-for-table-and-queue.sh       # Waits for the table and queue to be ready
├── scripts
│   ├── check-*-dlq.sh                    # Check the contents of the dlqs
│   ├── check-dynamodb.sh                 # Check the contents of DynamoDB
│   ├── check-sqs.sh                      # Check the contents of events queue
│   ├── redrive-dlq-to-sqs.sh             # Starts a redrive from the dlq to the events queue
│   ├── send-bad-client-sqs-event.sh      # Sends an event with a client-id that has no route.
│   ├── send-bad-data-sqs-event.sh        # Sends an event where data is a string not an object
│   ├── send-good-sqs-event.sh            # Sends a good event
│   ├── *.json                            # Event json data to be used with send-*.sh scripts
├── Dockerfile.*                          # Separate Dockerfiles per service
├── docker-compose.yml                    # Compose services (localstack, processor, deliver)
└── README.md
```

## Notes

- Only trusted producers should write to the SQS queue.
- Real deployments should include:
  - Authentication and authorization
  - Retry strategies and exponential backoff
  - Monitoring, tracing, and metrics for observability
  - An automated testing suite.
  - An IAC deployment for deploying actual infrastructure.
