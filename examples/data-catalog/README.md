# Data catalog example

This directory is a **sibling service** that treats HDF5 Agent as a data-platform
dependency. It does not import `gonum.org/v1/hdf5` or spawn Python. It talks to
the versioned JSON API through `pkg/hdf5client`.

```
┌──────────────┐     HTTP /api/v1      ┌──────────────┐
│ data-catalog │ ─────────────────────▶│ hdf5-agent   │
│ :8090        │     JSON + request id │ :8080        │
└──────────────┘                       └──────────────┘
        │                                      │
        │ GET /inventory                       │ /data/*.h5
        ▼                                      ▼
   dataset catalog                        HDF5 files
```

## Run with Compose

From the repository root:

```bash
mkdir -p data
go run ./cmd/create-testdata -out data/sample.h5
docker compose -f examples/docker-compose.yml up --build
```

Then:

```bash
curl -s http://127.0.0.1:8080/readyz
curl -s http://127.0.0.1:8090/inventory
```

`GET /inventory` walks every file the agent exposes and returns dataset
paths, dtypes, and shapes. That is the same pattern a larger product would use
for indexing, lineage, or a metadata store.

## Run without Docker

Terminal 1:

```bash
export HDF5_DATA_DIR=./data STATIC_DIR="" LOG_FORMAT=text
go run ./cmd/hdf5-agent
```

Terminal 2:

```bash
export HDF5_AGENT_URL=http://127.0.0.1:8080 HTTP_ADDR=:8090
go run ./examples/data-catalog
```

## Environment

| Variable | Default | Purpose |
|----------|---------|---------|
| `HDF5_AGENT_URL` | `http://127.0.0.1:8080` | Base URL of the HDF5 Agent API |
| `HTTP_ADDR` | `:8090` | Catalog listen address |
