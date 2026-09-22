FROM golang:1.27-alpine AS builder

WORKDIR /social

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o social-main ./cmd/server

FROM alpine

WORKDIR /social

COPY --from=builder /social/social-main .
COPY --from=builder /social/docs ./docs

RUN chmod +x social-main