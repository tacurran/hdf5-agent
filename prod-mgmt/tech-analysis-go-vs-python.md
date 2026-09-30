# Technical Analysis: Go vs Python Backend

**Purpose**: Compare the current Go HDF5 backend to the previous Python implementation (`backend.py` with h5py) that it replaced. This analysis supports architecture decisions and provides context for technical trade-offs.

**Date**: September 30, 2026  
**Author**: Product Management (with Engineering input)  
**Status**: Historical analysis; for reference and future re-evaluation

---

## Executive Summary

hdf5-agent currently uses a **Go backend** with `gonum.org/v1/hdf5` (CGO bindings to libhdf5) in `internal/hdf5store`. This replaced a previous **Python backend** (`backend.py` using h5py). The migration delivered:

- **Single-binary deployment**: No Python runtime or pip dependencies in production
- **Simpler Docker images**: Multi-stage Go build vs Python + virtualenv
- **Explicit concurrency**: Go goroutines with mutex serialization vs Python GIL
- **Stable API contract**: `/api/v1` OpenAPI spec in `api/openapi.yaml`
- **Typed client library**: `pkg/hdf5client` for Go services calling the API

**Trade-offs**:
- **CGO complexity**: Go backend still uses CGO (libhdf5 C library), adding build friction (pkg-config, Apple Silicon path workarounds, CGO_ENABLED=1 everywhere). This is being evaluated for replacement with a pure-Go HDF5 library (see `docs/pure-go-hdf5-analysis.md`).
- **Ecosystem maturity**: Python has a richer HDF5 ecosystem (h5py, pandas integration, Jupyter). Go ecosystem is leaner but sufficient for this service's narrow API surface.

**Recommendation**: The Go migration was correct for operational simplicity and API-first architecture. The remaining CGO dependency is a target for future removal via pure-Go HDF5 (spike in progress).

---

## Context: Why the Migration Happened

### The Previous Python Backend

The original implementation was a Python HTTP service:
- **`backend.py`** with Flask or similar framework
- **h5py** for HDF5 file access
- JSON API endpoints similar to current `/api/v1` routes
- Deployed in a Docker container with Python 3.x, pip, and system libhdf5

### Motivations for Migration

Based on `ARCHITECTURE.md`, `PROJECT_SUMMARY.md`, and repository history, the migration was driven by:

1. **Deployment complexity**: Python runtime + pip dependencies + system libhdf5 in Docker. Dependency conflicts between Python packages, wheel build issues, large image sizes.
2. **API-first architecture**: Go's `net/http` and structured logging (`log/slog`) align with a service-oriented architecture where the API is the contract, not the library.
3. **Client ecosystem**: Many downstream services (e.g., `examples/data-catalog`) are written in Go. A typed Go client library (`pkg/hdf5client`) was easier to maintain than a Python SDK.
4. **Operational simplicity**: Single Go binary, environment-only config, fast startup, no runtime dependencies beyond the OS-level libhdf5.
5. **Concurrency model**: Explicit goroutines and mutex-based HDF5 serialization vs implicit GIL in Python. More control over threading behavior.

---

## Comparison Matrix

| Dimension | Python (backend.py + h5py) | Go (current: gonum/hdf5 CGO) | Go (future: pure-Go HDF5) |
|-----------|----------------------------|------------------------------|---------------------------|
| **HDF5 Library** | h5py (Cython → libhdf5) | gonum.org/v1/hdf5 (CGO → libhdf5) | github.com/scigolib/hdf5 (pure Go) |
| **Runtime Dependencies** | Python 3.x, pip, libhdf5 | libhdf5 (runtime .so/.dylib) | None (static binary) |
| **Deployment** | Docker: Python base image + apt install + pip | Docker: multi-stage Go build + libhdf5 runtime | Docker: scratch/distroless + binary + static assets |
| **Binary Size** | ~500MB+ (Python + deps) | ~40MB (Go binary + libs) | ~15MB (static Go binary) |
| **Build Complexity** | pip install h5py (may need compiler) | CGO_ENABLED=1, pkg-config, Apple Silicon workarounds | CGO_ENABLED=0, standard Go build |
| **Cross-Compilation** | Not practical (Python is interpreted) | Difficult (CGO requires C toolchain per platform) | Trivial (GOOS/GOARCH, no C toolchain) |
| **Concurrency** | GIL limits parallelism; h5py releases GIL for C calls but still serialized | Explicit mutex in `Store`; libhdf5 not thread-safe | Per-file locking possible; pure-Go is thread-safe |
| **API Client** | Python SDK (requests) or curl | `pkg/hdf5client` (typed Go) | Same as Go CGO |
| **Ecosystem** | Rich: pandas, Jupyter, vast Python HDF5 tooling | Lean: gonum, limited Go HDF5 tooling | Same as Go CGO |
| **Error Handling** | Python exceptions → HTTP 500 handling | Go errors → structured `apierror` JSON | Same as Go CGO |
| **Startup Time** | ~1-2s (Python import overhead) | ~100ms (Go binary) | ~50ms (static Go binary) |
| **Maturity** | h5py: battle-tested, widely used | gonum/hdf5: stable but CGO friction | scigolib/hdf5: v0.x, less proven |
| **Interoperability** | h5py writes files readable by C tools, MATLAB | gonum writes HDF5; read by standard tools | Needs validation (spike in progress) |

---

## Detailed Trade-Off Analysis

### 1. Deployment and Operations

**Python**:
- Multi-layer Docker image: base Python image, apt install libhdf5-dev, pip install h5py + Flask + deps
- Larger images (~500MB+), slower cold starts, more attack surface
- Dependency conflicts (pip package versions, system libhdf5 vs h5py expectations)
- virtualenv or system site-packages management

**Go (CGO)**:
- Multi-stage Docker: build stage with Go + libhdf5-dev + pkg-config; runtime stage with libhdf5 shared libraries (~40MB)
- Single Go binary, but still needs libhdf5 .so files at runtime
- CGO_ENABLED=1, GODEBUG=cgocheck=0, pkg-config in CI and Makefile
- Apple Silicon workaround (see `run.sh`, `Makefile`, `.envrc`) to fix hardcoded Homebrew paths in vendored gonum

**Go (pure-Go, future)**:
- Single-stage Docker: build Go binary, copy to scratch or distroless base (~15MB)
- No libhdf5 runtime dependency, no CGO flags, no pkg-config, no Apple Silicon workaround
- Cross-compile trivially (linux/arm64, darwin/amd64, windows, wasm)
- Static binary: `./hdf5-agent` runs anywhere

**Verdict**: Go improves over Python significantly. Pure-Go would be the cleanest deployment story.

---

### 2. Performance and Concurrency

**Python (backend.py + h5py)**:
- GIL limits Python-level parallelism, but h5py releases the GIL during HDF5 C calls
- In practice, HDF5 operations are I/O or C-compute bound, so GIL impact is moderate
- Flask/Gunicorn with multiple workers provides request parallelism at the process level
- h5py itself is not thread-safe unless using specific modes; typically one file handle per worker

**Go (current CGO)**:
- Explicit goroutines for HTTP requests; concurrency up to developer
- `internal/hdf5store.Store` uses a global mutex to serialize all HDF5 C calls, because distro libhdf5 builds are typically not thread-safe
- Effect: only one HDF5 operation at a time across all files
- This is conservative but safe; Python multi-worker approach achieves similar parallelism at process level

**Go (pure-Go, future)**:
- Pure-Go HDF5 library is thread-safe (no C library with global state)
- Can relax global mutex to per-file locking: independent files accessed concurrently
- Potential for higher throughput when serving multiple clients reading different files

**Verdict**: Python and Go CGO have similar concurrency profiles (serialized HDF5 access). Pure-Go would enable true concurrent file access.

---

### 3. Build and Developer Experience

**Python**:
- `pip install h5py` requires compiler if binary wheels unavailable for the platform
- macOS: `brew install hdf5`, then pip may still need compilation
- CI: `apt-get install python3-dev libhdf5-dev`; pip install dependencies
- Developer setup: virtualenv, pip, possibly pyenv for version management

**Go (CGO)**:
- Requires libhdf5 C headers: `libhdf5-dev` (Debian/Ubuntu) or `brew install hdf5` (macOS)
- CGO_CFLAGS and CGO_LDFLAGS must be set correctly (automated in `Makefile`, `run.sh`, `.envrc`)
- Apple Silicon: vendored gonum hardcodes Intel Homebrew paths; workaround landed in PR #14
- CI: `apt-get install libhdf5-dev pkg-config`; CGO_ENABLED=1
- Developer setup: Go 1.26+, HDF5 headers, direnv or manual export of CGO flags

**Go (pure-Go, future)**:
- `go build` with no external dependencies
- No CGO flags, no pkg-config, no libhdf5 headers
- CI: just `go test ./...`
- Developer setup: Go 1.26+ (Node for frontend, but that's separate)

**Verdict**: Python and Go CGO both require system libhdf5; Go CGO has extra CGO flag complexity. Pure-Go eliminates all external dependencies.

---

### 4. API and Client Ecosystem

**Python**:
- Flask/FastAPI for HTTP API; Python SDK for clients (requests library)
- If downstream services are Go (e.g., `examples/data-catalog`), they call the HTTP API from Go with hand-written HTTP client or code-generated from OpenAPI
- Python SDK not useful for Go services

**Go (both CGO and pure-Go)**:
- `net/http` for HTTP API; structured JSON errors via `internal/apierror`
- `pkg/hdf5client`: typed Go client library that other Go services can import
- OpenAPI 3 contract in `api/openapi.yaml`; served at `/api/v1/openapi.yaml`
- Consistent logging with `log/slog`; request IDs with `X-Request-ID`

**Verdict**: Go backend aligns better with Go microservice ecosystems. Python backend would require maintaining both Python and Go clients or having Go services call HTTP directly.

---

### 5. Ecosystem and Interoperability

**Python (h5py)**:
- h5py is battle-tested, widely used in scientific Python
- Excellent integration with pandas, NumPy, Jupyter notebooks
- Extensive documentation and community support
- Files written by h5py are standards-compliant HDF5, readable by MATLAB, C tools, etc.

**Go (gonum/hdf5 CGO)**:
- Smaller ecosystem; gonum is reputable but not as widely adopted
- Uses libhdf5 C library under the hood, so interoperability is strong (same library as h5py)
- Files written by gonum are standards-compliant HDF5 (validated with h5dump, h5diff in tests)

**Go (pure-Go, future)**:
- Even smaller ecosystem; `github.com/scigolib/hdf5` is v0.x, single maintainer, ~31 stars (as of Sept 2026)
- Interoperability must be validated: round-trip tests with h5py, MATLAB, h5dump
- Risk: immature library may have bugs in format writing/reading (mitigated by spike + corpus tests)

**Verdict**: Python has the richest HDF5 ecosystem. Go CGO is sufficient for this service's narrow API surface (list, read, update). Pure-Go is riskier but manageable with validation.

---

### 6. Stability and Maintenance

**Python**:
- h5py is stable (v3.x), long-term maintenance, frequent security updates
- Python 3.x ecosystem is mature
- Flask/FastAPI are well-supported

**Go (gonum/hdf5 CGO)**:
- gonum.org/v1/hdf5 is stable but low-activity (vendored in this repo)
- Go standard library (net/http, log/slog) is very stable
- libhdf5 C library is mature (decades old) but distro versions vary

**Go (pure-Go, future)**:
- `github.com/scigolib/hdf5` is v0.x, unclear long-term maintenance
- Risk: if maintainer abandons the project, we may need to fork/vendor or revert to CGO
- Mitigation: vendor the dependency; spike behind an interface so we can swap back

**Verdict**: Python and Go CGO have strong stability. Pure-Go has higher risk, mitigated by interface abstraction and vendoring.

---

## Migration Impact: What Changed

### Removed from Python Version

- Python runtime, pip, virtualenv
- Flask or FastAPI framework
- h5py library
- System Python dependencies in Dockerfile

### Added in Go Version

- Go 1.26+ compiler
- `gonum.org/v1/hdf5` (vendored)
- `internal/hdf5store` package (Go implementation of file operations)
- `internal/httpserver` (HTTP handlers, middleware, versioning)
- `pkg/hdf5client` (typed Go client library)
- `api/openapi.yaml` (formal OpenAPI 3 contract)
- Build automation: `Makefile`, `run.sh`, `.envrc`
- CGO workarounds for macOS/Apple Silicon

### API Contract

The `/api/v1` API endpoints are stable and match the OpenAPI spec. The JSON response format and error envelopes (`{"error": {"code", "message", "request_id"}}`) are identical or improved from the Python version.

### Migration Notes in Codebase

Per `ARCHITECTURE.md` and `PROJECT_SUMMARY.md`:
> "internal/hdf5store: HDF5 operations (replaces old Python backend)"
> "This service replaces a previous Python backend (backend.py with h5py). All HDF5 operations now happen in internal/hdf5store."

There is **no Python remaining in the repository** today. The comparison is historical.

---

## Future Direction: Pure-Go HDF5

The remaining pain point is the CGO dependency on libhdf5. A pure-Go HDF5 implementation (see `docs/pure-go-hdf5-analysis.md`) would:

- Remove all CGO complexity (flags, pkg-config, Apple Silicon workarounds)
- Enable static binary, trivial cross-compilation, smaller Docker images
- Allow per-file concurrency (no global mutex)
- Reduce build and CI dependencies

**Status** (as of Sept 2026): Spike in progress; `github.com/scigolib/hdf5` is candidate. Decision gated on round-trip correctness tests (write with Go, read with h5py; read C-tool-produced files with Go). No code changes to the HTTP API or client library are planned; migration is contained to `internal/hdf5store`.

**Risk mitigation**:
- Spike behind an interface (`fileBackend`) so we can revert to CGO if needed
- Round-trip corpus tests with h5dump, h5diff, h5py validation
- Pin version, vendor dependency if governance is unclear
- Existing `store_test.go` integration tests provide first-line validation

---

## Recommendations for Future Architecture Decisions

1. **Maintain the API-first architecture**: The Go HTTP service with JSON API is the right model. Keep the API stable; internal store implementation can evolve.
2. **Complete the pure-Go migration if validated**: If the spike passes corpus tests, proceed with the swap. This removes the last major operational friction point.
3. **Do not revert to Python**: The operational benefits of Go (single binary, fast startup, typed client library, explicit concurrency) outweigh any ecosystem richness of Python for this service's scope.
4. **Keep the store interface abstracted**: `internal/hdf5store` should continue to own all HDF5 access. If another backend is ever needed (e.g., a different format, a remote HDF5 service), the HTTP layer should not change.
5. **Monitor for Python client demand**: If external partners need a Python SDK, generate it from the OpenAPI spec (e.g., with openapi-generator). Do not bring Python back into the service itself.

---

## Cross-References

- **Pure-Go evaluation**: See `docs/pure-go-hdf5-analysis.md` for detailed technical analysis of the CGO → pure-Go migration.
- **Architecture overview**: See `ARCHITECTURE.md` for package responsibilities and concurrency model.
- **API contract**: See `api/openapi.yaml` for the canonical `/api/v1` specification.

---

## Appendix: Performance Characteristics

*(Approximate; not rigorous benchmarks; illustrative only)*

| Operation | Python (h5py, multi-worker) | Go (CGO, global mutex) | Estimated Pure-Go |
|-----------|------------------------------|------------------------|-------------------|
| List 100 files | ~500ms | ~300ms | ~200ms (no CGO overhead) |
| Read dataset (10k points) | ~100ms | ~80ms | ~80ms (same I/O bound) |
| Update dataset (1k points) | ~150ms (read-modify-write) | ~120ms | ~120ms |
| Concurrent reads (2 files) | ~100ms (if workers > 1) | ~160ms (serialized) | ~80ms (parallel) |

**Key insights**:
- Go is slightly faster than Python for most operations due to lower runtime overhead
- Concurrency is limited in both Python (GIL/process model) and Go CGO (global mutex)
- Pure-Go would unlock parallel file access, improving multi-client throughput
- All are sufficient for the target dataset sizes (marketing analytics: <100k points per dataset)

---

## Conclusion

The migration from Python (`backend.py` + h5py) to Go (`internal/hdf5store` + gonum/hdf5) was the right technical decision for:
- Operational simplicity (single binary, fast startup, smaller Docker images)
- API-first architecture (stable `/api/v1`, typed Go client library)
- Concurrency control (explicit mutex vs implicit GIL)
- Deployment reliability (fewer runtime dependencies, easier CI/CD)

The remaining CGO dependency is a target for removal via pure-Go HDF5, which would complete the journey to a fully static, cross-platform, dependency-free service. This decision is being carefully validated to ensure interoperability with standard HDF5 tools (h5py, MATLAB, h5dump).

For marketing data analysts (our primary persona), the backend technology is invisible. What matters is the stable `/api/v1` contract, the UI for browsing and editing, and the reliability of files written by the agent. The Go migration delivered on those promises and set the foundation for continued improvements.
