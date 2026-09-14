# Stage 1: Build the embedded Web UI.
FROM oven/bun:1.4.0-alpine AS webui-builder

WORKDIR /src

COPY ui ./ui
COPY webui/package.json webui/bun.lock ./webui/
RUN cd webui && bun install --frozen-lockfile

COPY webui ./webui
RUN cd webui && NQ_WEBUI_BASE=/explorer/ NQ_VERSION=container \
    bun run build -- --outDir /out/webui_dist --emptyOutDir

# Stage 2: Build the Go binary with the generated Web UI.
FROM golang:1.27.1-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=webui-builder /out/webui_dist ./internal/server/webui_dist

# Build the final static binary
# - CGO_ENABLED=0 ensures a static binary with no C dependencies.
# - -ldflags="-s -w" removes symbol and debugging info for size reduction.
# - /out/nq is the output path for the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/nq .

# Stage 3: The final production image.
# Use gcr.io/distroless/static:nonroot, the smallest image that
# supports non-CGo static binaries and contains basic environment
# requirements without a shell or package manager.

FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /out/nq /app/nq

ENV NEPTUNE_URL=""

EXPOSE 8080

ENTRYPOINT ["/app/nq"]
CMD ["server", "--addr", "0.0.0.0:8080"]
