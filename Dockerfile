# Build the manager binary
# Force building go binaries on amd64 with cross compilation, avoid emulation
# Target distroless remains unaffected
FROM --platform=linux/amd64 quay.io/konveyor/builder:v1.23.6 AS builder
ARG TARGETOS
ARG TARGETARCH
RUN mkdir -p /gopath
ENV GOPATH=/gopath

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/
COPY pkg/ pkg/
COPY tools/ tools/
COPY vendor/ vendor/
# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-${GOARCH}} go build -a -o manager cmd/main.go
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-${GOARCH}} go build -a -o csv-generator ./tools/csv-generator/

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/csv-generator /usr/bin/csv-generator
USER 65532:65532

ENTRYPOINT ["/manager"]
