FROM node:20-bookworm AS frontend
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-bookworm AS backend
RUN apt-get update && apt-get install -y --no-install-recommends \
    libhdf5-dev pkg-config \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY api ./api
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY doc.go ./
ENV CGO_ENABLED=1
RUN go build -o /out/hdf5-agent ./cmd/hdf5-agent

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates wget \
    libhdf5-103-1 libhdf5-hl-100 \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --uid 65532 --system --home-dir /app --create-home hdf5agent \
    && mkdir -p /data /app/public \
    && chown -R hdf5agent:hdf5agent /data /app
COPY --from=backend /out/hdf5-agent /usr/local/bin/hdf5-agent
COPY --from=frontend /frontend/dist /app/public
USER 65532:65532
WORKDIR /app
EXPOSE 8080
ENV HTTP_ADDR=:8080 \
    HDF5_DATA_DIR=/data \
    STATIC_DIR=/app/public \
    LOG_FORMAT=json \
    GODEBUG=cgocheck=0
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/hdf5-agent"]
