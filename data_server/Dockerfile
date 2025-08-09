FROM golang:1.23.12 AS builder
WORKDIR /app

# protoc & plugins (already there if you kept it)
RUN apt-get update && apt-get install -y unzip \
    && curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v29.2/protoc-29.2-linux-x86_64.zip \
    && unzip protoc-29.2-linux-x86_64.zip -d /usr/local \
    && rm protoc-29.2-linux-x86_64.zip \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# **Install goose**
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# generate stubs if needed
RUN protoc -I protos/api-protos \
    --go_out=gen/projectv1 --go_opt=paths=source_relative \
    --go-grpc_out=gen/projectv1 --go-grpc_opt=paths=source_relative \
    protos/api-protos/project.proto

RUN go build -o /data_server ./cmd/main.go

FROM debian:12-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# goose binary
COPY --from=builder /go/bin/goose /usr/local/bin/goose
# server binary
COPY --from=builder /data_server /app/data_server
# migrations
COPY sql/migrations /app/sql/migrations

# tiny entrypoint: migrate then run server
COPY --from=builder /bin/sh /bin/sh
CMD sh -c 'goose -dir /app/sql/migrations postgres "$DATABASE_URL" up && exec /app/data_server'
EXPOSE 9090
