FROM golang:1.23.2-alpine AS builder
LABEL authors="dmiklopov"

WORKDIR /app

# Copy go.mod files and pkg dependencies
COPY loms/go.mod loms/go.sum ./
COPY pkg/ pkg/

RUN go mod download

# Copy loms source code
COPY loms/cmd/ cmd/
COPY loms/internal/ internal/
# Add any other loms subdirectories you have

RUN CGO_ENABLED=0 GOOS=linux go build -o lomsbuild ./cmd/

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/lomsbuild .
CMD ["./lomsbuild"]
