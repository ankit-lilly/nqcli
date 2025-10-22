# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/nq .


FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/nq ./nq

ENV NEPTUNE_URL=""
ENV NEPTUNE_TOKEN=""

EXPOSE 8080

ENTRYPOINT ["/app/nq"]
CMD ["server", "--addr", "0.0.0.0:8080"]
