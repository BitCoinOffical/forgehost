FROM golang:1.27-alpine AS builder

WORKDIR /hosting

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hosting-main ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hosting-worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hosting-monitoring ./cmd/monitoring

FROM alpine

WORKDIR /hosting

COPY --from=builder /hosting/hosting-main .
COPY --from=builder /hosting/hosting-worker .
COPY --from=builder /hosting/hosting-monitoring .

RUN chmod +x hosting-main
RUN chmod +x hosting-worker
RUN chmod +x hosting-monitoring