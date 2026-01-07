# Stage 1: The Build Environment
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the final static binary
# - CGO_ENABLED=0 ensures a static binary with no C dependencies.
# - -ldflags="-s -w" removes symbol and debugging info for size reduction.
# - /out/nq is the output path for the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/nq .

# Stage 2: The Final Production Image
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
