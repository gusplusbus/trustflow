# auth_server/Dockerfile
FROM golang:1.23-alpine

WORKDIR /app

# tools for go install + TLS roots (good practice even if sslmode=disable)
RUN apk add --no-cache git ca-certificates tzdata bash && update-ca-certificates

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
ENV PATH="/go/bin:${PATH}"

# bring in the code
COPY . .

# run migrations then start the server
CMD ["sh", "-c", "goose -dir ./sql/schema postgres \"$DATABASE_URL\" up && go run ./main.go"]
