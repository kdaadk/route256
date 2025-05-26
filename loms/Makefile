run:
	go run ./cmd

lint:
	gofumpt -l -w .
	find . -name '*.go' -exec gci write --skip-generated -s standard -s default {} \;
	find . -name '*.go' | xargs goimports -w
	golangci-lint run ./...

test: lint
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

#### Protobuf
vendor-proto/google/protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf vendor-proto/protobuf && \
	cd vendor-proto/protobuf && \
	git sparse-checkout set --no-cone src/google/protobuf && \
	git checkout
	mkdir -p vendor-proto/google
	mv vendor-proto/protobuf/src/google/protobuf vendor-proto/google
	rm -rf vendor-proto/protobuf

vendor-proto/validate:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/bufbuild/protoc-gen-validate vendor-proto/tmp && \
	cd vendor-proto/tmp && \
	git sparse-checkout set --no-cone validate && \
	git checkout
	mkdir -p vendor-proto/validate
	mv vendor-proto/tmp/validate vendor-proto
	rm -rf vendor-proto/tmp

vendor-proto/google/api:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/googleapis/googleapis vendor-proto/googleapis-tmp && \
	cd vendor-proto/googleapis-tmp && \
	git sparse-checkout set --no-cone google/api && \
	git checkout
	mkdir -p vendor-proto/google
	mv vendor-proto/googleapis-tmp/google/api vendor-proto/google
	rm -rf vendor-proto/googleapis-tmp

LOCAL_BIN:=$(CURDIR)/bin
PROTOC = PATH="$(PATH):$(LOCAL_BIN)" protoc
GO_PLUGIN        := google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
GO_GRPC_PLUGIN   := google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
GO_VALIDATE_PLUGIN   := github.com/envoyproxy/protoc-gen-validate@v1.0.4

.PHONY: .bin-deps
.bin-deps:
	@echo "Installing protoc plugins..."
	mkdir -p "$(LOCAL_BIN)"
	GOBIN="$(LOCAL_BIN)" go install $(GO_PLUGIN)
	GOBIN="$(LOCAL_BIN)" go install $(GO_GRPC_PLUGIN)
	GOBIN="$(LOCAL_BIN)" go install $(GO_VALIDATE_PLUGIN)
	GOBIN="$(LOCAL_BIN)" go install $(GO_VALIDATE_PLUGIN)
	@echo "Plugins installed to $(LOCAL_BIN)"
	@ls -la "$(LOCAL_BIN)"

.PHONY: .vendor-rm
.vendor-rm:
	rm -rf vendor-proto

.PHONY: .vendor-proto
.vendor-proto: .vendor-rm vendor-proto/google/protobuf vendor-proto/validate vendor-proto/google/api

LOMS_PROTO_PATH := proto
GO_OUT_DIR := $(LOMS_PROTO_PATH)
PROTO_FILES := $(wildcard $(LOMS_PROTO_PATH)/*.proto)

.PHONY: .protoc-generate
.protoc-generate: .bin-deps .vendor-proto
	@echo "Generating protobuf files..."
	@for proto in $(PROTO_FILES); do \
		echo "Processing $$proto..."; \
		$(PROTOC) \
			-I "$(LOMS_PROTO_PATH)" \
			-I vendor-proto \
			-I vendor-proto/google/api \
			--plugin=protoc-gen-go="$(LOCAL_BIN)/protoc-gen-go" \
			--go_out=$(GO_OUT_DIR) \
			--go_opt=paths=source_relative \
			--plugin=protoc-gen-go-grpc="$(LOCAL_BIN)/protoc-gen-go-grpc" \
			--go-grpc_out=$(GO_OUT_DIR) \
			--go-grpc_opt=paths=source_relative \
			--plugin=protoc-gen-validate="$(LOCAL_BIN)/protoc-gen-validate" \
			--validate_out=lang=go,paths=source_relative:$(GO_OUT_DIR) \
			$$proto || exit 1; \
	done
	go mod tidy

.all: .protoc-generate
