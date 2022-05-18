# =============================================================================
#  Multi-stage Dockerfile Example
# =============================================================================
#  This is a simple Dockerfile that will build an image of scratch-base image.
#  Usage:
#    docker build -t simple:local . && docker run --rm simple:local
# =============================================================================

# -----------------------------------------------------------------------------
#  Build Stage
# -----------------------------------------------------------------------------
FROM golang:alpine AS build

# Important:
#   Because this is a CGO enabled package, you are required to set it as 1.
ENV CGO_ENABLED=1

RUN apk add --no-cache \
    # Important: required for go-sqlite3
    gcc \
    # Required for Alpine
    musl-dev

WORKDIR /workspace

COPY . /workspace/

RUN \
    go mod init github.com/mattn/sample && \
    go mod tidy && \
    go install -ldflags='-s -w -extldflags "-static"' ./simple.go

RUN \
    # Smoke test
    set -o pipefail; \
    /go/bin/simple | grep 99\ こんにちわ世界099

# -----------------------------------------------------------------------------
#  Main Stage
# -----------------------------------------------------------------------------
FROM scratch

COPY --from=build /go/bin/simple /usr/local/bin/simple

ENTRYPOINT [ "/usr/local/bin/simple" ]
