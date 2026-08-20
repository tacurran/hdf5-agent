# Contributing

## Setup

Go 1.26+, Node 20+, `libhdf5-dev` (or `brew install hdf5`).

```bash
make setup
make testdata
make test
```

## Code layout

- Go service code lives under `cmd/`, `internal/`, and `pkg/`.
- Do not add Python. HDF5 access belongs in `internal/hdf5store`.
- Other services must use `/api/v1` or `pkg/hdf5client`.
- Exported Go functions need doc comments.
- HTTP changes need a matching edit to `api/openapi.yaml`.

## Tests

`go test ./...` and `npm test` (in `frontend/`) must pass. CI does not swallow
failures with `|| true`.

```bash
make lint
make metrics
```

## License

Contributions are licensed under Apache License 2.0.
