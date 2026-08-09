FROM golang:1.26-alpine AS builder

WORKDIR /auth

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-main ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-worker ./cmd/worker

FROM alpine

WORKDIR /auth

COPY --from=builder /auth/auth-main .
COPY --from=builder /auth/auth-worker .

RUN chmod +x auth-main auth-worker