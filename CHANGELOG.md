# Changelog

## [Unreleased]

### Added
- Versioned `/api/v1` JSON API with OpenAPI, `/healthz`, `/readyz`, request IDs
- Go packages for config, HDF5 store, HTTP server, and `pkg/hdf5client`
- Go tests for handlers, store, serialization, errors, and health
- Frontend Vitest coverage of the API client helpers
- Quality metrics (`metrics/baseline.json`, `cmd/metrics`, CI artifact)
- Sibling catalog example under `examples/data-catalog`
- GitHub Actions workflow at `.github/workflows/ci.yml`

### Changed
- Module path is `github.com/tacurran/hdf5-agent`
- Docker image runs as uid 65532 and serves the built UI from the same process
- Lint and tests fail CI when they fail

### Removed
- Python `backend.py` (replaced by `internal/hdf5store`)
- Root `ci.yml` that GitHub Actions never ran
- Scaffolding docs (`START_HERE.md`, `GITHUB_SETUP.md`, and related)

## [1.0.0] - 2026-04-30

### Added
- Initial Go + React MVP for browsing HDF5 files
