FROM brigadecore/go-tools:v0.1.0

ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0

WORKDIR /src
COPY exporter/ exporter/
COPY go.mod go.mod
COPY go.sum go.sum

RUN go build \
  -o bin/exporter \
  -ldflags "-w -X github.com/brigadecore/brigade-metrics/exporter/internal/version.version=$VERSION -X github.com/brigadecore/brigade-metrics/exporter/internal/version.commit=$COMMIT" \
  ./exporter

EXPOSE 8080

FROM scratch
COPY --from=0 /src/bin/ /brigade-metrics/bin/
ENTRYPOINT ["/brigade-metrics/bin/exporter"]
