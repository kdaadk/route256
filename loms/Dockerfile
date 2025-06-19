FROM golang:1.23.2-alpine AS builder
LABEL authors="dmiklopov"

WORKDIR /app
COPY . .

RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/lomsbuild ./cmd/

#ADD https://github.com/grpc-ecosystem/grpc-health-probe/releases/latest/download/grpc_health_probe /usr/bin/grpc_health_probe
#RUN chmod +x /usr/bin/grpc_health_probe

FROM alpine:latest
WORKDIR /app

COPY --from=builder /go/bin/goose /app/bin/goose
COPY migrations /app/migrations
COPY --from=builder /app/lomsbuild .
CMD ["app/lomsbuild"]
