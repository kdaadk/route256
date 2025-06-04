FROM golang:1.23.2-alpine AS builder
LABEL authors="dmiklopov"

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/lomsbuild ./cmd/

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/lomsbuild .
CMD ["app/lomsbuild"]
