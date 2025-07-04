FROM golang:1.24

RUN apt-get update && apt-get install -y netcat-openbsd

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/processor

COPY wait-for-localstack.sh .
CMD ["./wait-for-localstack.sh"]
