VERSION ?= dev
LDFLAGS = -s -w -X github.com/ankit-lilly/nqcli/cmd.version=$(VERSION)
BUILD_FLAGS = -trimpath -ldflags="$(LDFLAGS)"
GO_ENV = GOCACHE=/tmp/nqcli-go-build
BUN_ENV = BUN_TMPDIR=/tmp/bun-tmp
DESKTOP_CGO_LDFLAGS =
ifeq ($(shell uname -s),Darwin)
DESKTOP_CGO_LDFLAGS = -mmacosx-version-min=26.0
endif

.PHONY: build fmt fmt-go fmt-ui clean test test-go test-ui testdox tools desktop-deps desktop-dev desktop-build desktop-test desktop-bindings webui-deps webui-build webui-clean explorer-build explorer-clean build-with-explorer

build: webui-build
	@$(GO_ENV) CGO_ENABLED=0 go build $(BUILD_FLAGS) -o nq

fmt: fmt-go fmt-ui

fmt-go:
	@$(GO_ENV) go fmt ./...

fmt-ui:
	@$(BUN_ENV) bunx @biomejs/biome@2.4.15 format --write webui/src webui/scripts webui/package.json webui/tsconfig.json webui/vite.config.ts desktop/frontend/src desktop/frontend/package.json desktop/frontend/tsconfig.json desktop/frontend/vite.config.ts ui/graph-surface

clean:
	@rm -f nq nq-desktop

test: test-go test-ui desktop-test

test-go:
	@$(GO_ENV) CGO_ENABLED=0 go test . ./cmd/... ./internal/...

test-ui: webui-deps
	@cd $(WEBUI_SRC) && $(BUN_ENV) bun run typecheck
	@cd $(WEBUI_SRC) && $(BUN_ENV) bun run test

testdox:
	@go tool github.com/bitfield/gotestdox/cmd/gotestdox ./...

tools:
	@go install github.com/bitfield/gotestdox/cmd/gotestdox

# ---- Desktop app (Wails v3) ----

desktop-deps:
	@cd desktop/frontend && $(BUN_ENV) bun install

desktop-dev: desktop-deps desktop-bindings
	@cd desktop/frontend && $(BUN_ENV) bun run dev &
	@FRONTEND_DEVSERVER_URL=http://localhost:5173 CGO_ENABLED=1 go run ./desktop/

desktop-build: desktop-deps desktop-bindings
	@cd desktop/frontend && $(BUN_ENV) bun run build
	@CGO_ENABLED=1 CGO_LDFLAGS="$(DESKTOP_CGO_LDFLAGS)" go build $(BUILD_FLAGS) -o nq-desktop ./desktop/

desktop-test: desktop-deps
	@$(GO_ENV) go test ./internal/desktop/...
	@cd desktop/frontend && $(BUN_ENV) bun run typecheck
	@cd desktop/frontend && $(BUN_ENV) bun run test

# Generation is pinned to the same version as go.mod. Bundled runtime stays external in Vite.
desktop-bindings:
	@CGO_ENABLED=0 go run github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.20 generate bindings -f '-tags=server' -ts -i -b -names -d desktop/frontend/bindings ./desktop

# ---- Embedded web UI ----

WEBUI_SRC ?= ./webui
WEBUI_DIST = internal/server/webui_dist
WEBUI_OUT_DIR = ../$(WEBUI_DIST)

webui-deps:
	@cd $(WEBUI_SRC) && $(BUN_ENV) bun install

webui-build: webui-deps
	@echo "Building embedded web UI..."
	@rm -rf $(WEBUI_DIST)
	@cd $(WEBUI_SRC) && NQ_WEBUI_BASE=/explorer/ NQ_VERSION=$(VERSION) $(BUN_ENV) bun run build -- --outDir $(WEBUI_OUT_DIR) --emptyOutDir
	@touch $(WEBUI_DIST)/.gitkeep
	@echo "Web UI built into $(WEBUI_DIST)"

webui-clean:
	@rm -rf $(WEBUI_DIST)
	@mkdir -p $(WEBUI_DIST)
	@touch $(WEBUI_DIST)/.gitkeep

explorer-build: webui-build
explorer-clean: webui-clean
build-with-explorer: build
