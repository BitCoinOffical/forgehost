FROM golang:1.27-alpine AS builder

WORKDIR /hosting

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hosting-main ./cmd/server

FROM alpine

WORKDIR /hosting

COPY --from=builder /social/social-main .

RUN chmod +x hosting-main
