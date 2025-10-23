VERSION ?= dev
LDFLAGS = -s -w -X github.com/ankit-lilly/nqcli/cmd.version=$(VERSION)
BUILD_FLAGS = -ldflags="$(LDFLAGS)"

.PHONY: build
build:
	@go build $(BUILD_FLAGS) -o nq

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: clean
clean:
	@rm nq

.PHONY: test
test:
	@go test ./...
