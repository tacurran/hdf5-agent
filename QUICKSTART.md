# Quick start

```bash
make setup
make testdata
make test
HDF5_DATA_DIR=./data STATIC_DIR= LOG_FORMAT=text go run ./cmd/hdf5-agent
```

Frontend (separate terminal): `cd frontend && npm run dev` → http://localhost:3000

Docker: `docker compose up --build` → http://localhost:8080

See [README.md](README.md) for the API, environment variables, and the catalog example.
