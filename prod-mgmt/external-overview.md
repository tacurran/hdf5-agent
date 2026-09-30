# hdf5-agent: External Overview

**Audience**: Partners, external analysts, prospects  
**Date**: September 30, 2026  
**Version**: 1.0

---

## What is hdf5-agent?

**hdf5-agent** is an HTTP service for browsing and editing HDF5 files. It provides a JSON API and a web UI, designed for data analysts who need to package, validate, and correct datasets for AI application workflows.

HDF5 is a widely used file format for storing structured numeric data in scientific computing, machine learning, and data science. hdf5-agent makes HDF5 accessible to non-programmers by providing:

- A **web-based UI** for browsing file structure, viewing dataset contents, and making quick edits
- A **REST API** (`/api/v1`) for programmatic access from data catalogs, pipelines, and automation tools
- A **lightweight Docker container** that runs alongside your existing data infrastructure

---

## Why HDF5 for AI Data Transfer?

HDF5 is a natural choice for transferring tabular and numeric data into AI pipelines:

- **Preserves types**: Integers, floats, multi-dimensional arrays — no ambiguity like CSV
- **Efficient**: Compact binary format, faster than JSON for large numeric datasets
- **Widely supported**: Python (h5py, pandas), MATLAB, R, C/C++, and many ML frameworks can read HDF5
- **Schema-rich**: Hierarchical structure (groups and datasets) keeps related data organized

However, HDF5 has historically lacked user-friendly tools for non-programmers. hdf5-agent fills that gap.

---

## Use Cases

### Data Validation Before Handoff
A marketing analyst exports campaign performance data (impressions, clicks, conversions) from a BI tool, packages it as HDF5, and uses hdf5-agent to confirm the schema before sending it to the ML team. The ML team receives reliable data and avoids days of back-and-forth over schema mismatches.

### Data Catalog Integration
A data catalog service calls the hdf5-agent API to inventory available HDF5 files, extract metadata (dataset names, shapes, types), and present them in a searchable interface for data discovery.

### Quick Corrections
An analyst discovers a typo in a dataset (e.g., campaign ID entered incorrectly). Instead of re-running an hour-long export, the analyst edits the value directly in the hdf5-agent UI, saves, and the corrected file is ready immediately.

---

## Key Features

- **File operations**: List, create, delete HDF5 files via API or UI
- **Dataset browsing**: View hierarchical structure (groups and datasets), shapes, data types
- **Dataset reading**: Read full datasets or previews (up to 100k points by default)
- **Dataset editing**: Update individual values by index; changes persist and are readable by standard HDF5 tools
- **REST API**: Versioned JSON API (`/api/v1`) with OpenAPI 3 specification
- **Health endpoints**: `/healthz` and `/readyz` for load balancer and orchestrator integration
- **Structured errors**: Stable error codes (`not_found`, `invalid_path`, etc.) with request ID correlation
- **Docker-first**: Single container, environment-only config, no secrets required
- **Go client library**: `pkg/hdf5client` for Go services calling the API

---

## Getting Started

### Try Locally with Docker

```bash
# Clone the repository
git clone https://github.com/tacurran/hdf5-agent.git
cd hdf5-agent

# Create a data directory and sample file
mkdir -p data
make testdata

# Start the service
docker compose up --build

# Open the UI
open http://localhost:8080
```

The UI displays available HDF5 files. Click a file to browse its structure and datasets.

### API Example

```bash
# List files
curl -s http://localhost:8080/api/v1/files | jq

# Get file structure
curl -s http://localhost:8080/api/v1/files/sample.h5 | jq

# Read a dataset
curl -s "http://localhost:8080/api/v1/files/sample.h5/datasets?path=/temperatures" | jq
```

Full API documentation: [OpenAPI spec](https://github.com/tacurran/hdf5-agent/blob/main/api/openapi.yaml) or `GET /api/v1/openapi.yaml` on a running instance.

---

## Architecture

hdf5-agent is a single HTTP process built in Go. All HDF5 file access happens inside this service; other services call the JSON API rather than linking the HDF5 C library directly.

```
Browser / Data Catalog / Automation
        │  HTTP JSON (X-Request-ID)
        ▼
   hdf5-agent (:8080)
        │  Go backend + React UI
        ▼
   HDF5_DATA_DIR/*.h5
```

- **Backend**: Go 1.26+, `gonum.org/v1/hdf5` (CGO bindings to libhdf5)
- **Frontend**: React 18, Vite 8, Vitest
- **API**: OpenAPI 3.0.3, structured JSON errors
- **Deployment**: Docker multi-stage build, non-root runtime user (uid 65532)

See the [README](https://github.com/tacurran/hdf5-agent/blob/main/README.md) and [ARCHITECTURE](https://github.com/tacurran/hdf5-agent/blob/main/ARCHITECTURE.md) for details.

---

## Interoperability

Files written by hdf5-agent are **standards-compliant HDF5** and are readable by:
- Python `h5py` (the most common scientific Python library)
- MATLAB `h5read`, `h5write`
- HDF5 command-line tools (`h5dump`, `h5diff`, `h5repack`)
- R, C/C++, and other HDF5-compatible tools

Files written by those tools are browsable and editable by hdf5-agent.

---

## Deployment

hdf5-agent runs as a single Docker container with environment-only configuration. No secrets or complex setup required.

**Requirements**:
- Docker or Kubernetes
- Persistent storage for HDF5 files (mounted to `/data`)
- (Optional) Reverse proxy for authentication/HTTPS

**Configuration**: All via environment variables. No config files.

| Variable | Default | Meaning |
|----------|---------|---------|
| `HTTP_ADDR` | `:8080` | Listen address |
| `HDF5_DATA_DIR` | `./data` | Directory of HDF5 files |
| `CORS_ORIGINS` | `*` | Comma-separated origins |
| `MAX_DATASET_POINTS` | `100000` | Read/update size cap |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

See [README](https://github.com/tacurran/hdf5-agent/blob/main/README.md) for full configuration table.

---

## Project Status and Roadmap

hdf5-agent is **production-ready** for alpha deployments. Current version: 1.1.0 (as of Sept 2026).

**Recent milestones**:
- ✅ Stable `/api/v1` API with OpenAPI spec
- ✅ React UI for browsing and editing datasets
- ✅ Docker deployment with health endpoints
- ✅ Example data catalog integration (`examples/data-catalog`)

**Near-term roadmap**:
- CSV/JSON ingestion and export
- Authentication and multi-tenancy support
- Pure-Go HDF5 implementation (removing C library dependency for simpler builds)
- Dataset validation rules

See the [roadmap](roadmap.md) (internal) for details. For product questions or partnership inquiries, contact the maintainers via GitHub issues or email.

---

## License and Source Code

hdf5-agent is **open source** under the [Apache License 2.0](https://github.com/tacurran/hdf5-agent/blob/main/LICENSE).

- **Repository**: [github.com/tacurran/hdf5-agent](https://github.com/tacurran/hdf5-agent)
- **Issues and feature requests**: [GitHub Issues](https://github.com/tacurran/hdf5-agent/issues)
- **Contributing**: See [CONTRIBUTING.md](https://github.com/tacurran/hdf5-agent/blob/main/CONTRIBUTING.md)

---

## Contact

For questions, feedback, or partnership discussions:
- Open an issue: [github.com/tacurran/hdf5-agent/issues](https://github.com/tacurran/hdf5-agent/issues)
- Email: (add product team contact email if applicable)

We welcome contributions, bug reports, and feature requests from the community.

---

## Learn More

- **Quick start**: [QUICKSTART.md](https://github.com/tacurran/hdf5-agent/blob/main/QUICKSTART.md)
- **API documentation**: [api/openapi.yaml](https://github.com/tacurran/hdf5-agent/blob/main/api/openapi.yaml)
- **Architecture overview**: [ARCHITECTURE.md](https://github.com/tacurran/hdf5-agent/blob/main/ARCHITECTURE.md)
- **Example integration**: [examples/data-catalog](https://github.com/tacurran/hdf5-agent/tree/main/examples/data-catalog)
