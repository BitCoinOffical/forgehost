FROM golang:1.26-alpine AS builder

WORKDIR /notification

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o notification-app ./cmd

FROM alpine

WORKDIR /notification

COPY --from=builder /notification/notification-app .
COPY --from=builder /notification/internal/adapters/email/template ./template

RUN chmod +x notification-app
