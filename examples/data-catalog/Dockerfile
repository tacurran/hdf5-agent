FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY api ./api
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY examples/data-catalog ./examples/data-catalog
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/catalog ./examples/data-catalog

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --uid 65532 --system --no-create-home catalog
COPY --from=build /out/catalog /usr/local/bin/catalog
USER 65532:65532
EXPOSE 8090
ENV HTTP_ADDR=:8090
HEALTHCHECK --interval=15s --timeout=3s --start-period=5s CMD wget -qO- http://127.0.0.1:8090/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/catalog"]
