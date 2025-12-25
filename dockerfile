# syntax=docker/dockerfile:1.7

# ---- Build stage ----
ARG GO_VERSION=1.23.5
FROM golang:${GO_VERSION}-bookworm AS build

WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux

# Copy source and build
COPY . .

# Tidy modules using full source (no separate go mod download)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod tidy

# Build statically (your Makefile uses this exact command)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -v -o ./bytepurr .

# Sanity check
RUN test -x ./bytepurr

# ---- Runtime stage ----
FROM gcr.io/distroless/static:nonroot AS runtime
WORKDIR /app
COPY --from=build /src/bytepurr /app/bytepurr

USER nonroot:nonroot
ENTRYPOINT ["./bytepurr"]