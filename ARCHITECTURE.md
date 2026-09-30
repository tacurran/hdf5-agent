# Architecture

## Current State

HDF5 Agent is a single HTTP service. File access stays inside this process.
Every other service in a data platform uses `/api/v1` (JSON) or `pkg/hdf5client`.

The service currently serves both the React frontend (static assets) and the JSON API
from the same process on `:8080`. There is no authentication or authorization layer.

```
browser / catalog / batch job
        │  HTTP JSON  (X-Request-ID)
        ▼
   hdf5-agent  (:8080)
        │  internal/hdf5store (cgo → libhdf5)
        ▼
   HDF5_DATA_DIR/*.h5
```

## Target Architecture

**Status: Planned / Not Yet Implemented**

The target architecture introduces a **Go Backend for Frontend (BFF)** layer that sits
between the React UI and the HDF5 domain service. The BFF provides:

1. **Session management** and **authentication** workflows for browser clients
2. **Authorization** checks before proxying requests to the HDF5 agent
3. **API composition** — a single browser-facing endpoint that may aggregate multiple
   backend calls
4. **Security boundary** — the browser never calls identity admin APIs directly

The existing `hdf5-agent` service remains the **domain/data plane** for HDF5 operations.
It continues to expose `/api/v1` for programmatic access by batch jobs, catalog services,
and the BFF itself. The BFF becomes the **control plane** for browser-based workflows.

```
┌─────────────────────────────────────────────────────────────────┐
│                         Browser / UI                            │
└────────────┬────────────────────────────────────────────────────┘
             │  HTTPS (session cookie or token)
             ▼
┌────────────────────────────────────────────────────────────────┐
│                      Go BFF (planned)                          │
│  • Login/logout flows                                          │
│  • Session validation                                          │
│  • Authorization (ReBAC via Topaz)                             │
│  • Proxy to hdf5-agent with internal auth                      │
└────────────┬───────────────────────────────────────────────────┘
             │  HTTP JSON (X-Request-ID, internal credentials)
             ▼
┌────────────────────────────────────────────────────────────────┐
│                    hdf5-agent  (:8080)                         │
│                 internal/hdf5store (cgo → libhdf5)             │
└────────────┬───────────────────────────────────────────────────┘
             │
             ▼
        HDF5_DATA_DIR/*.h5

┌────────────────────────────────────────────────────────────────┐
│              catalog / batch job (service accounts)            │
└────────────┬───────────────────────────────────────────────────┘
             │  HTTP JSON (X-Request-ID, service credentials)
             └──────────────────► hdf5-agent  (:8080)
```

Programmatic clients (batch jobs, catalog services) continue to call `hdf5-agent` directly
using service credentials. The BFF is **only** for browser/UI workflows.

### Planned Identity Integration

**Integration with [Twothink Alembic Identity Plane](https://github.com/twothinkinc/alembic-identity-plane-go)**

The BFF will integrate with the Twothink Alembic Identity Plane, a Go service that
provides a complete identity and authorization stack:

- **[Ory Kratos](https://www.ory.sh/kratos/)** — identity provider, user registration,
  login/logout, session management, account recovery
- **[Ory Hydra](https://www.ory.sh/hydra/)** — OAuth2 and OpenID Connect provider for
  federated authentication
- **[Topaz](https://www.topaz.sh/)** — authorization engine implementing relationship-based
  access control (ReBAC) with support for complex permission policies
- **Security workflows** — password reset, email verification, multi-factor authentication

The identity plane already implements a BFF pattern: it exposes browser-safe endpoints
for login, session validation, and authorization checks while keeping Kratos/Hydra admin
APIs internal. The HDF5 BFF will:

1. **Delegate authentication** to the identity plane's login/session endpoints
2. **Validate sessions** on each request by calling the identity plane's session-check API
3. **Query permissions** via Topaz before allowing HDF5 operations (e.g., "can this user
   read file X?" or "does this user have role Y?")
4. **Pass validated identity context** to `hdf5-agent` (user ID, roles, permissions) so
   HDF5 operations can be audited and scoped

The identity plane will run as a separate service. The HDF5 BFF becomes a **client** of
that service, not a reimplementation of its logic.

**Reference**: [twothinkinc/alembic-identity-plane-go](https://github.com/twothinkinc/alembic-identity-plane-go)

### Migration Path

1. **Phase 1** (current): `hdf5-agent` is a standalone service with no auth
2. **Phase 2** (planned): Introduce the BFF for UI workflows; `hdf5-agent` remains
   accessible for service-to-service calls
3. **Phase 3** (planned): Integrate identity plane for login/auth/permissions in the BFF
4. **Phase 4** (planned): `hdf5-agent` enforces authorization via validated tokens/headers
   from the BFF or service accounts

The OpenAPI contract at `/api/v1` remains stable. Existing clients (batch jobs, catalog)
continue to work during and after the migration.

## Packages (Current Implementation)

| Package | Role |
|---------|------|
| `cmd/hdf5-agent` | Process: env config, slog, signals, graceful shutdown |
| `internal/config` | Environment-only configuration |
| `internal/hdf5store` | HDF5 operations (replaces the old `backend.py` / h5py path) |
| `internal/httpserver` | Versioned routes, structured errors, CORS, request IDs |
| `internal/apierror` | Error envelope `{error:{code,message,request_id}}` |
| `pkg/hdf5client` | Typed HTTP client for sibling services |
| `api` | Embedded OpenAPI document |

HDF5 C calls are serialized with a mutex. Distro libhdf5 builds are often not
thread-safe.

**Note**: The planned BFF layer will be a separate Go service (not yet implemented) that
calls `hdf5-agent` using `pkg/hdf5client` or direct HTTP. The BFF will handle browser
sessions, authentication, and authorization before proxying to the HDF5 domain service.

## Test strategy

1. **Pure unit tests** (no CGO): config, path sanitization, error JSON, HTTP
   handlers against a fake `Store`, HTTP client against `httptest`.
2. **HDF5 integration tests** (`internal/hdf5store`): create a temp file with
   `WriteSample`, then list, walk, read, update, create, delete.
3. **Frontend**: Vitest for API URL/error helpers. The UI is still thin; Go
   owns the behavior that used to live in Python.
4. **CI**: `gofmt`, `go vet`, `go test ./...` (must fail the job), frontend
   lint/test/build, Docker image + `/healthz`. Metrics are generated every run
   and compared to `metrics/baseline.json`.

There is no Python remaining in the repository.

## Production behavior

- Timeouts on the `http.Server` (`ReadTimeout`, `WriteTimeout`, `IdleTimeout`)
- `Shutdown` on SIGINT/SIGTERM
- Structured JSON logs (`log/slog`)
- `/healthz` vs `/readyz`
- Non-root container user 65532
- Path checks reject `..` and names that are not `*.h5` / `*.hdf5` basenames
