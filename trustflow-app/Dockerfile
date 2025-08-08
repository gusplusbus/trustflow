FROM golang:1.22-alpine

WORKDIR /app
COPY . .

# Only initialize go.mod if it doesn't already exist
RUN [ ! -f go.mod ] && go mod init trustflow || echo "go.mod already exists"

# Then always tidy
RUN go mod tidy

# Build from the correct entry point (inside /cmd if that's your setup)
RUN go build -o main ./cmd

EXPOSE 8080
CMD ["./main"]
