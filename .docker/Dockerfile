FROM golang:1.24-bookworm

ENV CGO_ENABLED=1
ENV PATH="/go/bin:${PATH}"

RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

RUN go install golang.org/x/tools/cmd/goimports@v0.38.0 && \
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.6.2 && \
    ln -sf /go/bin/golangci-lint /usr/local/bin/golangci-lint && \
    ln -sf /go/bin/goimports /usr/local/bin/goimports

WORKDIR /workspace
