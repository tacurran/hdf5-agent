# Quick start

```bash
make setup
make testdata
make test
```

Start the agent — either the helper script or the raw command:

```bash
./run.sh dev                    # backend + frontend, sets HDF5 CGO flags for you
# or run each process yourself:
HDF5_DATA_DIR=./data STATIC_DIR= LOG_FORMAT=text go run ./cmd/hdf5-agent
```

Frontend (separate terminal, only needed with the raw command):
`cd frontend && npm run dev` → http://localhost:3000

> **macOS / Apple Silicon:** bare `go run` / `go build` need HDF5 CGO paths.
> `make` and `./run.sh` set these automatically. For raw `go` commands, use
> [direnv](https://direnv.net) (an `.envrc` is included — run `direnv allow`) or
> `export CGO_CFLAGS="$(pkg-config --cflags-only-I hdf5)" CGO_LDFLAGS="$(pkg-config --libs-only-L hdf5)"`.

Docker: `docker compose up --build` → http://localhost:8080

See [README.md](README.md) for the API, environment variables, and the catalog example.
