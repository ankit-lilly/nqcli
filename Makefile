BUILD_FLAGS = -ldflags="-s -w"

.PHONY: build
build:
	@go build $(BUILD_FLAGS) -o nq

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: clean
clean:
	@rm nq
