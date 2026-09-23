# Analysis: replacing the cgo `gonum.org/v1/hdf5` backend with a pure‑Go HDF5 library

Status: draft / decision support. No code has been migrated; this document
evaluates whether we *should* migrate.

## 1. TL;DR

- The service currently reads/writes HDF5 through `gonum.org/v1/hdf5`, which is a
  **cgo** wrapper around the system `libhdf5` C library. That single dependency
  is the source of nearly all of our build friction: the CGO flag workaround we
  just shipped (Apple Silicon Homebrew paths), `CGO_ENABLED=1` everywhere,
  `GODEBUG=cgocheck=0`, `libhdf5-dev`/`libhdf5-103` in CI and the Docker image,
  a forced global mutex because distro `libhdf5` is not thread‑safe, and no
  cross‑compilation.
- A pure‑Go HDF5 implementation would remove all of that. The API surface we
  actually depend on is **small** and lives entirely in `internal/hdf5store`
  (two files), so a migration is well‑contained.
- **However**, the maturity/provenance of the available pure‑Go libraries is the
  gating risk (all `v0.x`, apparently one codebase mirrored under several
  accounts, self‑reported benchmarks). Recommendation: **prototype behind an
  interface and gate the switch on a format round‑trip test corpus** rather than
  swapping the dependency outright today.
- On the "works natively GPU" premise: dropping cgo does not *by itself* give
  GPU, and GPU compute isn't in this service's scope today. But a genuine
  **fully cgo‑free** "read HDF5 → GPU compute" path does exist via
  [`gogpu/wgpu`](https://github.com/gogpu/wgpu) (zero‑cgo WebGPU with compute
  shaders), so keeping the stack cgo‑free is the consistent choice *if* compute
  is ever added — see §3 / §3.1.

## 2. Where we are today

`internal/hdf5store` owns 100% of HDF5 access; callers use the HTTP API or
`pkg/hdf5client`, never the store directly. Two files touch the library:

- `internal/hdf5store/store.go` — list/walk/read/update/create/delete.
- `internal/hdf5store/sample.go` — writes the test fixture.

Build/runtime cost of the cgo dependency, by file:

| Location | Cost imposed by cgo/libhdf5 |
|----------|-----------------------------|
| `Makefile`, `run.sh`, `.envrc` | `pkg-config`‑derived `CGO_CFLAGS`/`CGO_LDFLAGS` shim (the fix from PR #14) |
| `Dockerfile` | `libhdf5-dev` + `pkg-config` in the build stage; `libhdf5-103-1`/`libhdf5-hl-100` in the runtime stage; `CGO_ENABLED=1`; `GODEBUG=cgocheck=0` |
| `.github/workflows/ci.yml` | `apt-get install libhdf5-dev pkg-config`; `CGO_ENABLED=1`; `GODEBUG=cgocheck=0` |
| `store.go` | `sync.Mutex` serializing *all* file access, because typical `libhdf5` builds are not thread‑safe (see the comment on `Store`), plus a `recover()` guard around C calls |

## 3. What dropping cgo actually buys us (and what it does not)

Real, defensible advantages of a pure‑Go implementation:

1. **Static, dependency‑free binaries.** No runtime `libhdf5`; the Docker runtime
   stage drops to a scratch/distroless base with just the binary + static assets.
2. **Trivial cross‑compilation.** `GOOS`/`GOARCH` builds (linux/arm64, windows,
   darwin, `wasm`, etc.) with no C toolchain or per‑platform `libhdf5`.
3. **No CGO flag pain.** Deletes the entire Apple‑Silicon workaround, the
   `pkg-config` shims, `CGO_ENABLED`, and `GODEBUG=cgocheck=0`.
4. **Simpler, faster CI/build.** No `apt-get install libhdf5-dev`; faster cold
   builds; smaller attack surface.
5. **Concurrency.** A pure‑Go reader lets independent `*File` handles run in
   parallel, so we could relax the global `Store` mutex to per‑file locking (the
   mutex exists specifically because C `libhdf5` isn't thread‑safe).
6. **Portability of failure modes.** Errors become Go errors instead of C
   aborts; we can drop the `recover()` guard.

What it does **not** buy us — correcting the stated premise:

> "advantage: works natively GPU"

HDF5 library choice is unrelated to GPU execution. Neither `gonum/hdf5` nor any
pure‑Go HDF5 library performs GPU compute; HDF5 is a **file format / storage
layer**. Removing cgo does not, by itself, enable or accelerate GPU work. The
most charitable reading of the phrase is "works natively on any
platform/architecture" — i.e. the **cross‑compilation / static‑binary** benefit
in points 1–2, which is real. Crucially, GPU compute is **not in this service's
scope today**: it reads/writes small numeric datasets and serves JSON. So GPU is
a forward‑looking concern, not a reason to migrate the storage layer now.

### 3.1 The coherent zero‑cgo GPU path: `gogpu`

An earlier draft claimed "most GPU bindings themselves require cgo, so no‑cgo and
GPU pull in opposite directions." That is **wrong**, and worth correcting:
[`gogpu/wgpu`](https://github.com/gogpu/wgpu) is a pure‑Go, **zero‑cgo** WebGPU
implementation (MIT, Go 1.25+) with **full compute‑shader support** and GPU→CPU
readback across Vulkan / Metal / DX12 / GLES plus a software fallback. It builds
with `CGO_ENABLED=0`, cross‑compiles, and uses pure‑Go runtime FFI (`goffi`)
rather than cgo.

This makes the "works natively GPU" premise *coherent*: a **fully cgo‑free
pipeline** is achievable — pure‑Go HDF5 read → pure‑Go WebGPU compute — shipping
as one static, cross‑compilable binary with no C toolchain and no `libhdf5`.
That is a genuine architectural story, and the pure‑Go HDF5 swap is the storage
half of it.

Caveats that keep this honest:

- **Build‑time vs runtime.** `gogpu/wgpu` is zero‑cgo *to build*, but at runtime
  it still dynamically loads the platform GPU driver (libvulkan/Metal/DX12/GLES)
  or falls back to a CPU software backend. So it is *not* dependency‑free at
  runtime the way a pure‑Go HDF5 reader is — a real GPU deployment still needs a
  driver present.
- **Maturity/provenance.** The `gogpu` org is very new (created Dec 2025) and has
  grown extremely fast (~250K LOC across the ecosystem in a few months, with
  heavy self‑promotion); all components are `v0.x`. Backends are advertised as
  "stable," but this is unproven for production and should be validated
  independently before any dependency.
- **Scope.** Nothing in the current service needs a GPU. Adopting `gogpu` would
  be a *new compute feature*, separate from the storage‑library decision this
  document is about. It should be scoped, justified, and prototyped on its own.

Bottom line: `gogpu` removes my original objection and shows a no‑cgo GPU path
exists — but it is an argument for *keeping the whole stack cgo‑free if/when GPU
compute is added*, not an independent justification for swapping the HDF5
backend today.

## 4. API surface we actually use

This is the entire dependency footprint that a replacement must cover. It is
deliberately narrow.

Files & handles:
- `hdf5.CreateFile` (`F_ACC_EXCL`, `F_ACC_TRUNC`), `hdf5.OpenFile`
  (`F_ACC_RDONLY`, `F_ACC_RDWR`), `File.Close`.

Group / hierarchy traversal (`CommonFG`):
- `CreateGroup`, `OpenGroup`, `Group.Close`.
- `NumObjects`, `ObjectNameByIndex`, `ObjectTypeByIndex`
  (`H5G_GROUP`, `H5G_DATASET`).

Datasets:
- `OpenDataset`, `CreateDataset`, `Dataset.{Space,Datatype,Read,Write,Close,Name}`.

Dataspaces / datatypes:
- `CreateSimpleDataspace`, `Dataspace.{SimpleExtentDims,SimpleExtentNPoints,Close}`.
- `Datatype.{Class,Size,Close}`, classes `T_INTEGER`/`T_FLOAT`, native types
  `T_NATIVE_DOUBLE`, `T_NATIVE_INT64`.

Data model constraints already baked into our code (these *reduce* migration
risk because we support far less than full HDF5):
- Only **integers** (1/2/4/8‑byte) and **floats** (4/8‑byte). Everything else
  returns `ErrUnsupportedType`.
- Reads pull the **whole** dataset into a flat slice; partial updates are
  **read‑modify‑write** of the whole array (`writeIndices`). No hyperslabs.
- Writes use **contiguous** layout, no chunking, no compression, small sizes.
- No attributes, links, references, compounds, or variable‑length types are
  read or written by the service.

So a replacement needs: create/open/close files; create/open groups; enumerate
children with type; create simple (contiguous) integer/float datasets; read a
full dataset into a typed slice; write a full typed slice back; report shape,
element count, datatype class + size. That's it.

## 5. Candidate libraries (2026)

### 5.1 Primary candidate: `github.com/scigolib/hdf5`

This appears to be the **canonical** pure‑Go HDF5 project; the near‑identical
`shyrmapp/hdf5` and `cwbudde/go-hdf5` repos look like mirrors/renames of it.

Concrete facts (verified from the repo, not just marketing):

- Module `github.com/scigolib/hdf5`, `go 1.25`, **MIT**. Only non‑test deps are
  `testify` and a YAML lib — genuinely pure Go, no cgo.
- Latest release **v0.14.1** (2026‑09‑02, "Write Compatibility & Security
  Fixes"); repo created 2025‑06‑10 (earliest of the cluster → likely the
  original); actively pushed. ~31 stars, 1 open issue at time of writing.
- Advertised: read contiguous/chunked/compact + GZIP; write datasets/groups/
  attributes, resizable dims, GZIP/LZF/Shuffle/Fletcher32; HDF5 v0/v2/v3
  superblocks; round‑trip through `h5dump`/`h5diff`/`h5repack` in CI.

**API shape differs from gonum**, so this is an adaptation, not a drop‑in:

```go
f, _ := hdf5.Open("data.h5")            // vs hdf5.OpenFile(path, flags)
f.Walk(func(path string, o hdf5.Object) // vs manual CommonFG enumeration
data, _ := ds.Read()                     // returns decoded data; also
                                         // ds.ReadStrings()/ds.ReadCompound()
fw, _ := hdf5.CreateForWrite("d.h5", hdf5.CreateTruncate, opts...) // vs CreateFile
```

Notable for our mapping (§4):
- `ds.Read()` decodes automatically across layouts (contiguous/chunked + GZIP),
  which is *more* than our current reader does — good for ingesting third‑party
  files. But its return typing differs from gonum's "allocate a typed slice and
  `Read(&slice)`" model, so `readValues`/`writeIndices` need reworking.
- Documented supported datatypes: **int32, int64, float32, float64**, string,
  compound. Our store also advertises **int8/int16** (`size` 1/2). That gap must
  be verified — either the library supports them and the docs are incomplete, or
  we narrow our advertised support / add conversion.
- Write path is `CreateForWrite` + options (incl. B‑tree rebalancing strategies),
  not gonum's `CreateFile`/`CreateGroup`/`CreateDataset`/`Write`. `sample.go`
  must be rewritten against it.

### 5.2 Due‑diligence concerns (important for a load‑bearing dependency)

1. **Provenance.** Confirm `scigolib/hdf5` is the upstream we track (the mirrors
   share identical copy). Pin an exact version; consider vendoring.
2. **Maturity.** `v0.x`, ~31 stars, single maintainer signal — no long‑term
   API‑stability guarantee. Compatibility/benchmark tables are **self‑authored**;
   treat as claims to verify.
3. **Correctness surface.** HDF5 is a large, subtle format. Our needs are tiny,
   but files we *write* must be readable by third‑party consumers (Python
   `h5py`, MATLAB, the C tools), and files we *read* may come from those tools.
   That interop is exactly where an immature reader/writer is most likely to
   break — hence the round‑trip corpus gate in §9.

Other options for completeness:
- Keep `gonum/hdf5` (status quo; cgo).
- `go-hdf5`/`sbinet/go-hdf5` — also cgo, so it doesn't address the goal.
- **Don't use HDF5 at all** for our internal fixtures (see §8).

## 6. Migration effort & mapping

Effort is concentrated and low‑to‑moderate, dominated by verification rather
than code volume.

| Work item | Scope |
|-----------|-------|
| Introduce a small internal interface (`fileBackend`) so the store depends on *our* abstraction, not a third‑party type | `internal/hdf5store` |
| Reimplement `store.go` read/walk/update against the new library | ~1 file, mapping in §4 |
| Reimplement `sample.go` writes | ~1 file |
| Delete cgo scaffolding: `.envrc`, the `pkg-config` blocks in `Makefile`/`run.sh`, `CGO_ENABLED`/`GODEBUG` and `libhdf5*` installs in `Dockerfile` and `ci.yml`; slim runtime image to distroless | build/CI |
| Relax `Store` mutex to per‑file (optional follow‑up) | `store.go` |
| README/QUICKSTART/ARCHITECTURE updates (drop the "HDF5 C dev headers" prerequisite and the Apple‑Silicon note) | docs |

What stays unchanged and de‑risks the swap:
- `internal/httpserver`, `api`, `pkg/hdf5client`, and the HTTP contract are
  untouched — the store's public methods keep the same signatures.
- `internal/hdf5store/store_test.go` exercises the public `Store` API against a
  `WriteSample` fixture, so it becomes the first correctness gate for the swap.

## 7. Risks

- **Format interop (highest).** A pure‑Go writer must emit files that `h5py`/C
  tools read byte‑correctly, and its reader must handle files those tools
  produce (incl. chunked + compressed data we don't currently emit but might
  ingest). Mitigation: a round‑trip corpus test (write in Go → verify with
  `h5dump`/`h5diff`; read C‑tool‑produced files → assert values).
- **Dependency risk.** `v0.x`, unclear governance. Mitigation: pin a version,
  vendor if needed, and keep the `fileBackend` interface so we can revert to
  `gonum/hdf5` without touching the HTTP layer.
- **Datatype coverage.** Verify the library reads/writes native int8/16/32/64
  and float32/64 in the exact in‑memory layout our `Read`/`Write` expect.
- **Performance.** Pure Go may be slower for very large datasets; our
  `MAX_DATASET_POINTS` cap (default 100k) makes this largely moot, but should be
  spot‑checked.

## 8. Alternative worth noting

Because the service only *writes* HDF5 for its own sample fixture and only
*reads* small numeric datasets, a lighter option is to keep HDF5 purely as an
*interchange* concern and use a trivial format internally. That's a bigger
product decision and out of scope here, but it would remove the dependency
entirely. Listed for completeness; not recommended without stakeholder input.

## 9. Recommendation

1. **Do not swap the production dependency yet.** The benefit is real
   (kills cgo/libhdf5 and all its build friction) but the replacement's maturity
   is unproven for our interop requirements.
2. **Spike it behind an interface.** Add a `fileBackend` abstraction in
   `internal/hdf5store`, implement it with **`github.com/scigolib/hdf5`** (§5.1,
   pinned to `v0.14.1` or later), and run the existing store tests plus a new
   **round‑trip corpus** (`h5dump`/`h5diff` against Python‑ and C‑produced files).
   Confirm int8/int16 coverage during the spike (§5.1).
3. **Gate the decision on that corpus.** If it passes cleanly and the canonical
   module has credible governance, proceed with the full migration (§6) and
   delete the cgo scaffolding. If not, stay on `gonum/hdf5`; the PR #14 workaround
   already makes the cgo path painless day‑to‑day.
4. **Treat GPU as a separate, forward‑looking initiative — but a coherent one.**
   A no‑cgo GPU path genuinely exists via [`gogpu/wgpu`](https://github.com/gogpu/wgpu)
   (§3.1), so a fully cgo‑free "read HDF5 → GPU compute" stack is achievable. Do
   not justify the *storage‑library* swap on GPU today, but if/when compute is on
   the roadmap, keeping the stack cgo‑free (pure‑Go HDF5 + `gogpu`) is the
   consistent choice. Scope and prototype `gogpu` independently, with the same
   maturity scrutiny applied to the HDF5 candidates.

### Decision matrix

| Option | Kills cgo/libhdf5 | Cross‑compile / static | Interop risk | Maturity | Effort |
|--------|:---:|:---:|:---:|:---:|:---:|
| Status quo (`gonum/hdf5`) | ❌ | ❌ | low (battle‑tested C) | high | none |
| Pure‑Go HDF5 (spike → migrate) | ✅ | ✅ | **medium‑high (verify)** | low (`v0.x`) | low‑moderate |
| Drop HDF5 internally | ✅ | ✅ | n/a (changes format) | n/a | high / product decision |
