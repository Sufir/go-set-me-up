FROM golang:1.24-bookworm

ENV CGO_ENABLED=0
ENV PATH="/go/bin:${PATH}"

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

WORKDIR /workspace
