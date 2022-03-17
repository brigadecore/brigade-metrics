FROM --platform=$BUILDPLATFORM brigadecore/go-tools:v0.7.0 as builder

ARG VERSION
ARG COMMIT
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0

WORKDIR /src
COPY exporter/ exporter/
COPY go.mod go.mod
COPY go.sum go.sum

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
  -o bin/exporter \
  -ldflags "-w -X github.com/brigadecore/brigade-foundations/version.version=$VERSION -X github.com/brigadecore/brigade-foundations/version.commit=$COMMIT" \
  ./exporter

EXPOSE 8080

FROM gcr.io/distroless/static:nonroot as final

COPY --from=builder /src/bin/ /brigade-metrics/bin/

ENTRYPOINT ["/brigade-metrics/bin/exporter"]
