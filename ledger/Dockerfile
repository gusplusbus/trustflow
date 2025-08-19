# syntax=docker/dockerfile:1

# --- build stage ---
FROM golang:1.23.12 AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ledger ./cmd/main.go

# --- runtime stage (distroless, non-root) ---
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /
COPY --from=build /out/ledger /ledger
USER nonroot:nonroot

EXPOSE 9091
ENTRYPOINT ["/ledger"]
