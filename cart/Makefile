run:
	go run ./cmd & \
	go run ./mock/product



.PHONY: test coverage

test:
	go test ./...

benchmark:
	go test -bench=. -benchmem ./...

coverage:
	go test -coverprofile=coverage.tmp.out ./...
	grep -v "_mock.go" coverage.tmp.out > coverage.out
	go tool cover -func=coverage.out

linter:
	golangci-lint run --config ./golangci.yml


