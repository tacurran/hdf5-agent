# Architecture draft: Go‑native, GPU‑capable causal‑analysis platform

Status: **draft for discussion.** This sketches where the HDF5 agent fits in the
larger effort and proposes a concrete design for the next MVP step (exporting
JSON Schema and CSV). It is deliberately opinionated so we have something to
argue with; every "decision" below is negotiable. Assumptions and open
questions are called out explicitly — please correct them.

## 1. Vision & scope

Goal: a **Go‑native** data‑science environment for **causal analysis** that is
simple to build/deploy (single static binaries, no cgo), scales horizontally,
and can offload heavy numeric kernels to the **GPU without a C toolchain**.

Today the reading/decomposition algorithms are mostly **Python + C++**. The
long‑term direction is to move that logic into Go for operational simplicity and
parallelism, keeping Python/C++ only where it still earns its place (mature
GPU/DS libraries) and behind a stable service boundary.

This document covers three horizons:

- **Now:** the HDF5 agent as the storage/IO service, plus the next MVP feature —
  format export (JSON Schema, CSV).
- **Next:** a Go‑native compute service for causal analysis with CPU parallelism
  and optional GPU offload.
- **Later:** a full platform (pipelines, columnar interchange, multiple compute
  backends), migrating off Python/C++ incrementally.

Non‑goals here: committing to a specific causal‑inference algorithm set, or a
final GPU stack. Those need their own design once the primitives below are real.

## 2. Current state (as‑is)

The HDF5 agent is a single Go HTTP service; all HDF5 access stays in‑process and
everything else talks to it over `/api/v1` (JSON) or `pkg/hdf5client`. It uses
`gonum.org/v1/hdf5` (cgo → `libhdf5`) today. See
[`ARCHITECTURE.md`](../ARCHITECTURE.md) and, for the planned move off cgo,
[`docs/pure-go-hdf5-analysis.md`](pure-go-hdf5-analysis.md).

```
browser / catalog / batch job
        │  HTTP JSON (X-Request-ID)
        ▼
   hdf5-agent (:8080)
        │  internal/hdf5store  (cgo → libhdf5, to become pure Go)
        ▼
   HDF5_DATA_DIR/*.h5
```

Key existing seams we will reuse:
- The `Store` interface in `internal/httpserver` already abstracts persistence
  behind method calls — new capabilities (export, schema) can be added as
  handlers without touching the HTTP core.
- `hdf5store.Node` (file/group/dataset tree with `Shape`/`Dtype`/`NPoints`) and
  `hdf5store.Dataset` (typed `Data any`) are the natural inputs to exporters.
- The structured error envelope, request IDs, CORS, timeouts, and OpenAPI
  contract are already in place and should extend to new endpoints unchanged.

## 3. Target architecture (to‑be)

Service‑oriented, each service a single cgo‑free static binary, composed over
versioned JSON (and, later, Arrow for bulk/columnar transfer):

```
                    ┌─────────────────────────────────────────────┐
                    │                 clients / UI                 │
                    └───────────────┬─────────────────────────────┘
                                    │ HTTP JSON (+ Arrow stream for bulk)
        ┌───────────────────────────┼───────────────────────────┐
        ▼                           ▼                           ▼
  hdf5-agent                 causal-engine                pipeline / DAG
  (storage + IO)             (compute)                    (orchestration)
   • read/walk HDF5           • CPU: goroutine pools       • job graph
   • write HDF5               • GPU: WebGPU compute        • retries, caching
   • EXPORT: JSON Schema,       (gogpu/wgpu) offload       • lineage
     CSV, (Arrow/Parquet)     • causal algos (ported       • schedules
   • pure-Go HDF5 (target)      from py/cpp incrementally)
        │
        ▼
   HDF5_DATA_DIR / object store
```

Guiding principles:
1. **cgo‑free by default.** Pure‑Go HDF5 (see the analysis doc) + pure‑Go GPU
   (`gogpu/wgpu`) keep the whole stack buildable with `CGO_ENABLED=0` and
   cross‑compilable. cgo/C++/CUDA is allowed only behind an explicit service
   boundary during migration (§8).
2. **Stable contracts over shared code.** Services integrate through versioned
   HTTP/Arrow contracts, never by importing each other's internals — exactly the
   rule the HDF5 agent already enforces (`pkg/hdf5client`, not `internal/`).
3. **Storage vs compute split.** The HDF5 agent owns bytes and format
   conversion; the causal engine owns math. This keeps GPU concerns out of the
   IO service and lets each scale independently.

## 4. MVP now: format export (JSON Schema + CSV)

This is the immediate task and fits the current codebase with no architectural
change — new package + new read‑only endpoints.

### 4.1 Package: `internal/export`

A small, format‑agnostic encoder registry so new formats (Arrow, Parquet, JSON)
drop in later:

```go
// Encoder streams one dataset (or a file's schema) to w.
type Encoder interface {
    ContentType() string        // e.g. "text/csv", "application/schema+json"
    Extension() string          // "csv", "schema.json"
    EncodeDataset(w io.Writer, ds *hdf5store.Dataset) error
}

// SchemaEncoder describes structure rather than values.
type SchemaEncoder interface {
    ContentType() string
    EncodeSchema(w io.Writer, root *hdf5store.Node) error
}

// Registry maps a format string ("csv", "jsonschema") to an encoder.
```

Design choices:
- **Streaming.** Encoders write directly to the `http.ResponseWriter` (chunked),
  not into a buffer, so large exports don't blow memory and can bypass the
  `MAX_DATASET_POINTS` JSON cap with a separate, higher `MAX_EXPORT_POINTS` (or
  no cap for CSV, which is streamed row‑by‑row).
- **Pure functions of `Node`/`Dataset`.** Exporters depend only on the existing
  store types, so they're trivially unit‑testable with no HDF5/cgo.

### 4.2 CSV export

- Endpoint: `GET /api/v1/files/{name}/datasets/export?path=/g/ds&format=csv`
- `Content-Type: text/csv`, `Content-Disposition: attachment; filename="ds.csv"`.
- Shape handling (proposed, **decision needed** — §9 Q1):
  - 1‑D `[n]` → one value per row, single column `value`.
  - 2‑D `[r,c]` → `r` rows × `c` columns (row‑major).
  - N‑D → flattened row‑major with leading index columns `i0,i1,…,i{k-1},value`
    (lossless and tool‑agnostic; a `reshape` option can come later).
- Header row with column names; numeric formatting stable and locale‑independent.

### 4.3 JSON Schema export

"JSON Schema" is ambiguous; the proposed meaning (**decision needed** — §9 Q2)
is a **JSON Schema (draft 2020‑12) document describing the file's structure and
datatypes** — a machine‑readable contract downstream consumers can validate a
JSON rendering of the data against. Example intent:

- Endpoint: `GET /api/v1/files/{name}/schema?format=jsonschema`
- `Content-Type: application/schema+json`.
- Mapping: file/group → `object` with `properties` per child; dataset → `array`
  (nested to `len(shape)` depth) whose leaf `items` type derives from `Dtype`
  (`int*`/`float*` → `number`/`integer`, `string` → `string`), annotated with
  `x-hdf5: {shape, dtype, npoints}` for round‑trip fidelity.

This gives us a contract artifact for the causal engine and external tools, and
composes with the existing OpenAPI story.

### 4.4 API/versioning notes

- New endpoints are additive under `/api/v1`; update `api/openapi.yaml` (the repo
  already serves it at `/api/v1/openapi.yaml` and enforces it in review).
- Errors reuse the existing envelope and codes (`invalid_request`, `not_found`,
  `unsupported_type`, `too_large`).
- `pkg/hdf5client` gains `ExportDataset(ctx, name, path, format) (io.ReadCloser,…)`
  and `Schema(ctx, name, format)` so sibling Go services consume exports without
  reimplementing HTTP.

## 5. Compute layer (causal‑engine) — next horizon

A separate Go service (or, initially, a library) that consumes datasets (via
`hdf5client` or, for bulk, an Arrow stream) and runs causal‑analysis primitives.

- **CPU parallelism first.** Go's goroutines + a bounded worker pool cover
  embarrassingly parallel work (per‑variable stats, bootstrap resampling,
  pairwise conditional‑independence tests). This alone likely beats the current
  Python for orchestration‑heavy workloads.
- **GPU offload where it pays.** Route dense linear‑algebra / kernel work
  (covariance/correlation matrices, kernel‑based independence tests, large
  matrix multiplies) to **WebGPU compute shaders via `gogpu/wgpu`**
  (zero‑cgo, WGSL kernels, GPU→CPU readback). Keep a CPU fallback (also the
  software backend) so results are reproducible without a GPU.
- **Algorithm surface (illustrative, not committed):** correlation/partial
  correlation, conditional‑independence testing, constraint‑ or score‑based
  structure learning (PC/GES‑style), DAG representation and manipulation.
  Porting these from Python/C++ is the substantive, uncertain work and needs its
  own design + parity tests (§8).

**Honest GPU caveat.** WebGPU compute (`gogpu`) is young and less proven for
data science than CUDA/RAPIDS. Two viable strategies (§9 Q3):
- **(A) All‑in Go+WGSL:** maximal simplicity/portability, higher porting risk and
  possibly lower peak throughput than CUDA.
- **(B) Hybrid:** Go owns IO/orchestration/CPU‑parallel work; GPU‑heavy kernels
  stay in a C++/CUDA service behind the JSON/Arrow boundary until Go+WGSL parity
  is proven. Lower risk, keeps some cgo/C++ at the edge.

## 6. Data & interchange strategy

- **HDF5** remains the system of record / ingest format (owned by the agent).
- **JSON** for control plane, small results, and schema contracts.
- **CSV** for human/tool‑friendly tabular export (MVP).
- **Apache Arrow** (pure‑Go `apache/arrow-go`, no cgo) is the recommended
  **bulk/columnar interchange** between the agent and the compute engine, and the
  natural staging layout for GPU upload (contiguous typed buffers ≈ GPU storage
  buffers). Recommend adopting Arrow as the internal bulk format once export
  exists (**decision needed** — §9 Q4). **Parquet** for durable columnar output.

```
HDF5 (store) ──read──▶ agent ──Arrow stream──▶ causal-engine ──▶ GPU buffers (wgpu)
                       │                                   └────▶ CPU worker pool
                       └──export──▶ CSV / JSON Schema / Parquet (consumers)
```

## 7. Cross‑cutting concerns

Reuse what the agent already does well, and standardize it across services:
- **Contracts/versioning:** OpenAPI per service; semver on the JSON API; Arrow
  schemas versioned alongside.
- **Observability:** `log/slog` structured logs + `X-Request-ID` propagation
  across service hops (already present); add metrics/traces per service.
- **Config:** environment‑only (as today) — 12‑factor, container‑friendly.
- **Security/ops:** non‑root containers, path sanitization, request size limits,
  timeouts, graceful shutdown — all already established patterns to inherit.
- **Testing:** pure‑Go unit tests for exporters/algorithms (no cgo); golden‑file
  parity tests against the current Python/C++ outputs during migration; format
  round‑trips validated with external tools (`h5dump`, a JSON Schema validator).

## 8. Migration from Python/C++ (strangler pattern)

1. **Wrap, don't rewrite (yet).** Put the existing Python/C++ algorithms behind
   the same JSON/Arrow contract the Go engine will expose. Traffic flows through
   a stable interface regardless of implementation language.
2. **Reimplement per‑algorithm in Go**, gated by **golden‑output parity tests**
   (same inputs → numerically equivalent results within tolerance).
3. **Flip per‑algorithm** once parity holds; keep the reference implementation
   available for regression comparison.
4. **Retire** the Python/C++ path when all algorithms are ported and stable.

This lets simplicity/scalability wins land incrementally without a risky big‑bang
rewrite, and keeps a numeric oracle throughout.

## 9. Open questions / decisions needed

- **Q1 — CSV N‑D semantics:** index‑columns‑plus‑value (lossless) vs a
  `reshape=r,c` option vs erroring on >2‑D? Proposed: index columns by default.
- **Q2 — "JSON Schema" meaning:** structure/type contract (proposed) vs
  data‑as‑JSON vs both (schema + `format=json` data export)?
- **Q3 — GPU strategy:** all‑in Go+WGSL (`gogpu`) vs hybrid Go‑IO + CUDA/C++
  compute behind a boundary? Drives how aggressively we pursue pure‑Go.
- **Q4 — Bulk interchange:** adopt Arrow as the internal bulk format now, or stay
  JSON‑only until the compute engine exists?
- **Q5 — Causal scope:** which algorithms are in the MVP compute milestone? This
  determines the first GPU kernels to build.
- **Q6 — Deployment target:** where does this run (single node w/ GPU, k8s,
  serverless)? Affects service boundaries and the GPU‑driver assumption
  (`gogpu` needs a driver at runtime or falls back to software).

## 10. Suggested next steps

1. Land the MVP **export** feature (`internal/export`, CSV + JSON Schema
   endpoints, OpenAPI + client updates, unit tests). Small, self‑contained,
   unblocks downstream consumers.
2. In parallel, run the **pure‑Go HDF5 spike** from the analysis doc so storage
   becomes cgo‑free.
3. Prototype **one GPU kernel** end‑to‑end with `gogpu/wgpu` (e.g. a
   correlation matrix) to de‑risk strategy Q3 before committing the compute
   engine's design.
4. Resolve Q1–Q6, then write the dedicated **causal‑engine** design doc.

> This is a starting point, not a decision record. Tell me which of Q1–Q6 to
> lock down and I'll turn the relevant section into a concrete spec (and, for the
> export MVP, implement it).
