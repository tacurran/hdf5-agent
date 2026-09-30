# Product Requirements Document (PRD)

**Product**: hdf5-agent  
**Version**: 1.0 (MVP baseline)  
**Date**: September 30, 2026  
**Status**: Living document  
**Owner**: Product Management  

---

## Overview

hdf5-agent is an HTTP service for browsing and editing HDF5 files, designed to enable marketing data analysts to package, validate, and correct datasets for AI application workflows. The agent provides a JSON API and a React UI for interacting with HDF5 files without requiring analysts to write Python code or install HDF5 C libraries.

## Goals

1. **Enable self-service validation**: Marketing analysts can browse HDF5 file structure and dataset contents via a web UI or REST API before handing data to AI teams.
2. **Provide a stable API for catalog integration**: Marketing ops can build data catalog services that inventory HDF5 datasets by calling a versioned JSON API.
3. **Simplify deployment**: The agent runs as a single Docker container with no external dependencies (no Python runtime, no separate database), suitable for marketing infrastructure environments.
4. **Support basic editing**: Analysts can make small corrections (fix typos, update individual values) without re-running data export pipelines.
5. **Ensure interoperability**: Files written by the agent are readable by standard HDF5 tools (h5py, MATLAB, h5dump). Files from those tools are browsable by the agent.

## Non-Goals (Explicit Scope Limits)

- **Not a data warehouse or query engine**: No SQL, no aggregations, no complex filtering. Users read and write whole datasets or individual values by index.
- **Not a high-performance scientific platform**: Optimized for marketing datasets (thousands to tens of thousands of rows), not terabyte-scale research data.
- **No built-in authentication/authorization in v1.0**: Deployment environments are expected to handle authn/authz via network policies or reverse proxy. Multi-tenancy is deferred.
- **No schema migration or transformation**: The agent reads and writes HDF5; it does not convert to/from CSV, JSON, or SQL in v1.0.
- **No advanced HDF5 features in v1.0**: Compression, chunking, attributes, compound types, variable-length strings are out of scope unless needed for read compatibility with existing files.

## Target Personas

### Primary: Marketing Data Analyst (Alex)

- **Background**: Works in marketing analytics at a mid-sized company. Uses Excel, BI tools, and occasionally calls REST APIs. Not a programmer.
- **Pain**: Needs to package campaign performance data (impressions, clicks, conversions by date and segment) for the ML team. CSV loses types, JSON is too verbose. HDF5 works for ML but Alex has no tool to validate the file before sending it.
- **Needs**: A web UI to open an HDF5 file, see the datasets, confirm shapes and types, and spot-check a few values. Ability to fix typos without asking a developer.

### Secondary: Marketing Operations Engineer (Jamie)

- **Background**: Builds data pipelines and automation. Comfortable with Python, Go, Docker. Responsible for the marketing data catalog.
- **Pain**: Needs to programmatically inventory HDF5 files, extract metadata (dataset names, shapes, types), and surface that in a searchable catalog UI. Doesn't want to link libhdf5 in multiple services.
- **Needs**: A REST API to list files, get file structure, read dataset metadata. Structured errors, health endpoints, and OpenAPI docs. Go client library as a convenience.

### Tertiary Consumer: ML Engineer (Morgan)

- **Background**: Builds AI models in Python (h5py, pandas, scikit-learn, TensorFlow). Receives HDF5 files from marketing.
- **Pain**: Files sometimes have unexpected schemas, missing datasets, or wrong data types, causing days of back-and-forth with marketing.
- **Benefit**: If Alex validates files with hdf5-agent before sending, Morgan gets reliable schemas and types, reducing iteration cycles.

## Use Cases

### UC-1: Browse HDF5 File Structure (Must Have)

**Actor**: Marketing analyst (Alex)  
**Precondition**: Alex has an HDF5 file `campaign_data.h5` on disk  
**Flow**:
1. Alex opens the hdf5-agent UI in a browser (e.g., `http://localhost:8080`)
2. The UI lists available HDF5 files in the configured data directory
3. Alex clicks `campaign_data.h5`
4. The UI displays the hierarchical structure: groups and datasets
5. Alex clicks a dataset `/impressions/daily`
6. The UI shows shape (e.g., `[365, 5]`), data type (e.g., `int64`), and a sample of values

**Success criteria**:
- File list loads in <1 second for directories with <100 files
- Tree structure renders correctly with nested groups
- Dataset preview shows shape, dtype, and at least the first 100 values

### UC-2: Validate Dataset Values (Must Have)

**Actor**: Marketing analyst (Alex)  
**Precondition**: Alex has opened a dataset in the UI  
**Flow**:
1. Alex sees a table or array preview of dataset values
2. Alex scrolls or searches for a specific date range
3. Alex confirms that campaign IDs match expected values

**Success criteria**:
- Datasets up to `MAX_DATASET_POINTS` (default 100k) display fully
- Larger datasets show a truncation notice
- Values are formatted clearly (no scientific notation confusion for small integers)

### UC-3: Correct a Value in a Dataset (Should Have)

**Actor**: Marketing analyst (Alex)  
**Precondition**: Alex has identified a typo in `/conversions/daily[42]`  
**Flow**:
1. Alex clicks "Edit" on the dataset
2. Alex changes the value at flattened index 42 from `10` to `100`
3. Alex clicks "Save"
4. The UI calls `PUT /api/v1/files/campaign_data.h5/datasets` with `{"path": "/conversions/daily", "indices": [42], "values": [100]}`
5. The backend reads the dataset, updates index 42, writes the dataset back
6. The UI confirms success

**Success criteria**:
- Update completes in <2 seconds for datasets <100k points
- Updated value persists; re-opening the file shows the corrected value
- No corruption of other values in the dataset
- Standard HDF5 tools (h5dump, h5py) see the updated value

### UC-4: Programmatic Dataset Inventory (Must Have)

**Actor**: Marketing ops engineer (Jamie)  
**Precondition**: Jamie is building a data catalog service  
**Flow**:
1. Jamie's catalog calls `GET /api/v1/files` → receives `{"files": [{"name": "campaign_data.h5", "size_bytes": 12345}, ...]}`
2. For each file, Jamie calls `GET /api/v1/files/{name}` → receives the tree of groups/datasets with metadata (names, paths, shapes, dtypes)
3. Jamie indexes this metadata in the catalog search backend

**Success criteria**:
- API returns JSON matching OpenAPI 3 spec at `/api/v1/openapi.yaml`
- Structured errors with stable error codes (`not_found`, `invalid_path`, etc.)
- Request ID correlation via `X-Request-ID` header
- Health and readiness endpoints (`/healthz`, `/readyz`) for load balancer checks

### UC-5: Create an Empty File for Gradual Population (Should Have)

**Actor**: Marketing ops engineer (Jamie)  
**Precondition**: Jamie is building a pipeline that writes datasets incrementally  
**Flow**:
1. Jamie calls `POST /api/v1/files` with `{"name": "new_campaign.h5"}`
2. The agent creates an empty HDF5 file
3. Jamie's pipeline later calls write endpoints (future: external tool) to add datasets

**Success criteria**:
- Empty file is valid HDF5, openable by h5py
- Concurrent create attempts return `409 Conflict`

### UC-6: Delete a File (Must Have)

**Actor**: Marketing analyst (Alex) or ops (Jamie)  
**Precondition**: An old test file `test.h5` exists  
**Flow**:
1. Call `DELETE /api/v1/files/test.h5`
2. The agent deletes the file from the data directory
3. Subsequent `GET /api/v1/files` does not list `test.h5`

**Success criteria**:
- File is removed from disk
- Deleting a non-existent file returns `404 Not Found`
- No other files are affected

## Acceptance Criteria (MVP)

### Functional

- [ ] **API**: All endpoints in `api/openapi.yaml` (v1) implemented and return correct status codes, JSON schemas, and error envelopes
- [ ] **UI**: React app can list files, display tree structure, show dataset metadata (shape/dtype), and preview data
- [ ] **Edit**: `PUT /api/v1/files/{name}/datasets` can update individual indices in a dataset; changes persist and are visible to h5py
- [ ] **Health**: `/healthz` returns 200 if process is alive; `/readyz` returns 200 if data directory is readable, 503 otherwise
- [ ] **CORS**: Configurable CORS origins for UI-API separation in dev
- [ ] **Request ID**: `X-Request-ID` is propagated and logged; returned in error responses
- [ ] **Graceful shutdown**: SIGTERM triggers shutdown within `SHUTDOWN_TIMEOUT` (default 20s)
- [ ] **Docker**: Single multi-stage Dockerfile; runtime runs as non-root user (uid 65532); `/healthz` passes in container

### Non-Functional

- [ ] **Performance**: File list <1s for 100 files; tree structure <2s for files with 100 datasets; dataset read <2s for 100k points
- [ ] **Reliability**: No data corruption; updated datasets pass `h5diff` validation against expected values
- [ ] **Interoperability**: Files written by the agent are readable by h5py 3.x and HDF5 command-line tools (h5dump, h5diff)
- [ ] **Deployment**: Runs with only environment variable config; no secrets required; data directory is the only external mount
- [ ] **Logging**: Structured JSON logs with `log/slog`; configurable log level (`debug`, `info`, `warn`, `error`)
- [ ] **Testing**: Unit and integration tests pass in CI; frontend tests pass; Docker build succeeds

### Documentation

- [ ] `README.md` quick start works on macOS and Linux
- [ ] OpenAPI spec is accurate and served at `/api/v1/openapi.yaml`
- [ ] `ARCHITECTURE.md` describes the HTTP/store layering and CGO threading constraints
- [ ] `examples/data-catalog` demonstrates API usage from a Go service

## Success Metrics

### Launch (First 3 Months)

- **Adoption**: 2-3 marketing teams use the agent in staging/dev environments
- **API calls**: >1000 API requests per week from catalog or manual curl/client usage
- **Data validated**: At least 10 HDF5 files browsed or edited per week
- **Zero data corruption incidents**: No reports of files becoming unreadable after agent writes

### Steady State (6 Months)

- **Adoption**: 5+ marketing teams use the agent in production
- **API calls**: >10k API requests per week
- **Uptime**: >99.5% availability (measured via health check endpoint)
- **Reduction in schema iteration**: Anecdotal feedback from ML engineers that schema mismatches decreased
- **Community PRs or issues**: External contributors open issues or PRs, indicating broader interest

## Open Questions

1. **Authentication**: Do we need built-in authn/authz, or is network-level access control sufficient? (Initial answer: defer to deployment environment for MVP; revisit if multi-tenancy is needed)
2. **Write from UI**: Should analysts be able to write entirely new datasets via UI, or only edit existing ones? (Initial answer: edit-only for MVP; creation via API or external tools)
3. **Ingestion/Export**: Should the agent convert CSV → HDF5 or HDF5 → CSV? (Initial answer: defer to v2.0; focus on HDF5-native workflows first)
4. **Compression**: Do analyst-generated files need GZIP/LZF compression? (Initial answer: no for MVP; revisit if file sizes become a concern)
5. **Attribute support**: Do datasets need HDF5 attributes (metadata tags)? (Initial answer: no for MVP; revisit based on user requests)

## Dependencies and Risks

### Dependencies

- **Go HDF5 library**: Currently `gonum.org/v1/hdf5` (CGO → libhdf5). Evaluating pure-Go alternative (`github.com/scigolib/hdf5`) to remove CGO dependency. Decision gated on round-trip correctness tests (see `docs/pure-go-hdf5-analysis.md`).
- **HDF5 C library**: Must be installed in Docker runtime and CI environments. Complicates builds (CGO flags, pkg-config). Motivates pure-Go migration.
- **React/Vite**: Frontend is thin; UI development is secondary to API stability.

### Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| **Data corruption in writes** | High | Integration tests with h5diff validation; read-modify-write correctness tests; code review for all store methods |
| **CGO build friction on macOS** | Medium | Shipped workaround in `Makefile`, `run.sh`, `.envrc` (PR #14); pure-Go spike in progress |
| **Library maturity (if pure-Go adopted)** | Medium | Spike behind interface; round-trip corpus test; pin version; vendor if needed |
| **Performance for large datasets** | Low | `MAX_DATASET_POINTS` cap; truncation warnings; not targeting >100k point datasets in v1 |
| **Multi-tenancy assumptions** | Medium | Defer authn/authz to deployment; document assumption; re-evaluate if multiple teams need isolated access |

## Rollout Plan

1. **Internal alpha**: Deploy in staging for 1-2 marketing teams; gather feedback on UI usability and API ergonomics
2. **Beta with docs**: Expand to 3-5 teams; ensure `README`, `QUICKSTART`, and OpenAPI docs are clear; iterate based on support requests
3. **General availability**: Announce via internal data platform channels; publish external-facing overview; support production deployments
4. **Post-launch**: Monitor metrics; collect feature requests; prioritize v2.0 roadmap (ingestion, export, auth, pure-Go if validated)

## Revision History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 1.0 | 2026-09-30 | Initial MVP PRD | PM |

