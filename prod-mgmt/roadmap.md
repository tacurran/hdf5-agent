# Product Roadmap

**Last updated**: September 30, 2026  
**Owner**: Product Management  
**Status**: Living document (review monthly)

This roadmap organizes features and initiatives into **Now / Next / Later** horizons, oriented around marketing analyst workflows and AI data transfer use cases. It is outcome-focused and avoids committing to specific dates, recognizing the autonomous nature of the development process.

---

## Roadmap Horizons

### Now (Current Focus)

**Goal**: Solidify the MVP for marketing analyst validation workflows and catalog integration.

| Initiative | Outcome | Status | Notes |
|------------|---------|--------|-------|
| **API stability and observability** | `/api/v1` is fully documented in OpenAPI 3; structured errors; request ID correlation; health/readiness endpoints | ✅ Complete | OpenAPI served at `/api/v1/openapi.yaml`; `pkg/hdf5client` for Go services |
| **Basic UI for browsing and editing** | Marketing analysts can list files, view tree structure, preview dataset values, and edit individual indices via web UI | ✅ Complete | React 18 + Vite 8; Vitest tests; served from same origin in Docker |
| **Dataset read and update** | `GET /api/v1/files/{name}/datasets` reads full datasets (up to `MAX_DATASET_POINTS`); `PUT` updates flattened indices in place | ✅ Complete | Integration tests validate read-modify-write correctness |
| **Catalog integration example** | `examples/data-catalog` demonstrates API usage from a Go sibling service | ✅ Complete | Shows `pkg/hdf5client` usage; Docker Compose setup |
| **Operational documentation** | README, QUICKSTART, ARCHITECTURE, Docker deployment, CI/CD setup | ✅ Complete | Includes Apple Silicon CGO workaround; `.envrc` for direnv |
| **Quality metrics baseline** | `make metrics` generates `metrics/current.json` and compares to `metrics/baseline.json` in CI | ✅ Complete | Tracks LOC, test coverage, dependencies |

**Next steps in Now**:
- Gather feedback from 2-3 internal marketing teams using the agent in staging
- Document any edge cases or usability friction in the UI
- Monitor API usage patterns and error rates

---

### Next (Near-Term Opportunities)

**Goal**: Address adoption blockers and expand workflows for self-service data preparation.

| Initiative | Outcome | Rationale / User Need |
|------------|---------|----------------------|
| **Pure-Go HDF5 migration** | Remove CGO dependency; static binary; trivial cross-compilation; no libhdf5 runtime requirement | Eliminates build friction (CGO flags, Apple Silicon workarounds, pkg-config). See `docs/pure-go-hdf5-analysis.md`. Decision gated on round-trip corpus tests with h5py/h5dump validation. **High impact** for ops simplicity. |
| **Improved dataset editing UX** | UI supports bulk edits (e.g., edit a range of indices, copy-paste values from CSV); better table navigation for large datasets | User feedback: editing one index at a time is tedious. Analysts want to correct a batch of values or copy-paste from a spreadsheet. |
| **CSV/JSON ingestion** | `POST /api/v1/files/{name}/ingest` accepts CSV or JSON payload and writes an HDF5 dataset | Analysts currently export CSV from BI tools, then manually convert to HDF5. Agent should streamline this. Reduces dependency on external scripts. |
| **HDF5 → CSV export** | `GET /api/v1/files/{name}/datasets?path=...&format=csv` returns dataset as CSV | ML engineers or analysts sometimes need CSV for quick inspection in Excel or BI tools. Complements ingestion. |
| **Authentication / multi-tenancy** | Configurable authentication (API key, OAuth, or authn proxy integration); tenant isolation for files | Needed when multiple marketing teams share one agent instance. Initially deferred to deployment environment (network policies); will become a blocker for broader internal rollout. |
| **Dataset validation rules** | API endpoint to validate dataset schema against a JSON schema or user-defined rules | Analysts want to automate schema checks (e.g., "confirm dataset /impressions/daily has shape [365, 5] and dtype int64"). Reduces manual inspection; integrates with CI/CD for data pipelines. |

**Prioritization notes**:
- **Pure-Go HDF5** is high-priority if spike validation succeeds; unblocks static binary and cross-platform builds.
- **CSV ingestion and export** address frequent user requests; relatively low complexity.
- **Authentication** is critical for multi-tenant production use; needed before wider internal rollout beyond alpha teams.
- **Dataset validation rules** are nice-to-have; can be deferred if adoption is strong without them.

---

### Later (Future Vision)

**Goal**: Expand the agent into a full data preparation platform for marketing → AI workflows.

| Initiative | Outcome | Rationale / User Need |
|------------|---------|----------------------|
| **Data lineage and provenance** | Track dataset creation history: source system, export timestamp, transformations applied | Marketing ops and ML engineers need to understand where data came from and how it was modified. Supports audit trails and debugging. |
| **Workflow automation** | Scheduled ingestion from marketing APIs (Google Analytics, Salesforce, etc.); automated export to ML pipeline storage (S3, GCS) | Analysts currently run manual scripts. Agent could orchestrate end-to-end workflows: fetch → package → validate → deliver. |
| **Advanced HDF5 features** | Support for attributes, compression (GZIP/LZF), chunked datasets, hierarchical group management | Needed if file sizes grow or if third-party files use these features. Currently not a blocker; defer until user requests. |
| **Collaborative editing** | Multiple analysts can view and edit the same file; real-time conflict resolution or locking | Useful for team workflows (e.g., one analyst prepares data, another QA's it). Complex to implement; requires WebSocket or polling for updates. |
| **Python SDK** | Official Python client library generated from OpenAPI spec | If external partners or ML engineers prefer Python for scripting. Generate with openapi-generator; maintain as separate package. |
| **Visualization and plotting** | UI shows line charts, histograms, scatter plots for dataset previews | Analysts want quick visual sanity checks (e.g., "does this trend look right?"). Complements table view. Medium complexity (frontend charting library). |
| **Integration with data catalogs** | Pre-built connectors for popular catalog platforms (Amundsen, DataHub, Alation) | Marketing ops shouldn't have to build custom catalog integration. Provide reference implementations or plugins. |
| **Dataset diffing** | Compare two datasets or two versions of the same dataset; highlight changes | Useful for QA: "what changed between v1 and v2 of this export?" Medium complexity (diff algorithm + UI). |

**Prioritization notes**:
- **Data lineage** and **workflow automation** are strategic; align with broader data platform vision. Require coordination with other teams (data engineering, ML platform).
- **Advanced HDF5 features** are deferred unless real user files require them (e.g., third-party compressed files can't be read).
- **Collaborative editing** is complex; only pursue if multiple teams report friction from serial editing workflows.
- **Python SDK** is straightforward (generate from OpenAPI); schedule when external partnerships demand it.

---

## Themes by Persona

### For Marketing Analysts (Alex)
- **Now**: Browse and edit datasets via UI; validate schemas before handoff
- **Next**: CSV ingestion/export; bulk editing; visual sanity checks
- **Later**: Workflow automation (fetch from APIs → package → deliver); visualization

### For Marketing Ops (Jamie)
- **Now**: REST API for catalog integration; Docker deployment; health endpoints
- **Next**: Authentication for multi-tenant use; dataset validation rules
- **Later**: Data lineage tracking; pre-built catalog connectors; workflow orchestration

### For ML Engineers (Morgan, indirect beneficiary)
- **Now**: Reliable schemas from analyst-validated files; fewer back-and-forth cycles
- **Next**: CSV export for quick inspection; schema validation automation
- **Later**: Data lineage (understand provenance); Python SDK for scripting

---

## Cross-Cutting Initiatives

These span multiple horizons and support all personas:

| Initiative | Horizon | Impact |
|------------|---------|--------|
| **Pure-Go HDF5 migration** | Next | Simplifies builds, enables static binary, removes CGO friction for all users |
| **Documentation and onboarding** | Ongoing (Now) | Analyst-facing guides; API tutorials; video walkthroughs; reduce time-to-first-validation |
| **Performance optimization** | Later | Support larger datasets (>100k points); concurrent file access (if pure-Go adopted) |
| **Security and compliance** | Next (auth), Later (audit logs) | Authentication, authorization, audit trails for regulated industries |
| **Cloud-native deployment** | Later | Kubernetes Helm chart; Terraform modules; auto-scaling for high-traffic catalog queries |

---

## Decision Framework

When prioritizing features, apply these criteria:

1. **Impact on adoption**: Does this unblock a new team or expand usage to a new use case?
2. **Operational simplicity**: Does this reduce deployment complexity or maintenance burden?
3. **User pain intensity**: How often do users hit this friction? How severe is the workaround?
4. **Technical risk**: How much validation or prototyping is needed? What's the fallback plan?
5. **Strategic alignment**: Does this support broader data platform goals (catalog integration, AI workflow automation)?

**Example**: Pure-Go HDF5 scores high on operational simplicity, technical risk is mitigated by spike, and it unblocks cross-platform builds (strategic). → Prioritize in Next if spike succeeds.

**Example**: Collaborative editing scores high on user pain for specific teams, but low on adoption (most teams have serial workflows) and high on complexity. → Defer to Later.

---

## Open Questions for Roadmap Refinement

- **Ingestion priorities**: Which formats do analysts need most? CSV? JSON? SQL dumps? Parquet?
- **Validation rules**: Should these be user-defined (JSON schema) or built-in (e.g., "no nulls", "shape matches")?
- **Multi-tenancy model**: Per-team data directories? File-level ACLs? Separate agent instances?
- **Catalog integration**: Should agent push metadata to catalog, or should catalog poll agent? Push may require webhook/event system.
- **Workflow automation**: Does this belong in the agent, or in a separate orchestrator (Airflow, Prefect) that calls the agent API?

Answers inform feature scoping and are revisited in quarterly roadmap reviews.

---

## Metrics to Track Progress

| Metric | Current | Target (6 months) | Target (12 months) |
|--------|---------|-------------------|--------------------|
| **Adoption** (teams using agent) | 1-2 alpha teams | 5+ teams in production | 10+ teams; external partners |
| **API calls per week** | ~1k (alpha) | >10k | >50k |
| **Datasets validated per week** | ~10 files | >50 files | >200 files |
| **Uptime** (via health check) | N/A (pre-prod) | >99.5% | >99.9% |
| **Support requests per month** | ~5 (alpha feedback) | <10 (stable product) | <20 (scaled usage) |
| **Time to first validation** (new user) | ~30 min (setup + trial) | <15 min (docs + UI onboarding) | <10 min (self-service) |

---

## How to Use This Roadmap

- **Monthly reviews**: PM updates roadmap based on velocity, feedback, and blockers. Adjust Next/Later boundaries.
- **Quarterly planning**: Align roadmap themes with broader data platform OKRs. Revisit open questions with user research.
- **Stakeholder communication**: Use this document in leadership reviews, partnership discussions, and team planning.
- **Engineering input**: Eng leads provide complexity estimates and technical risk assessments to inform prioritization.

---

## Conclusion

This roadmap balances immediate adoption needs (stable API, usable UI, catalog integration) with longer-term vision (workflow automation, lineage, multi-tenancy). The **Now** horizon is solid; the agent is production-ready for alpha use. The **Next** horizon focuses on removing operational friction (pure-Go HDF5) and expanding self-service workflows (CSV ingestion/export, auth). The **Later** horizon aligns with strategic data platform goals (lineage, orchestration, advanced features).

Roadmap priorities are reviewed monthly and adjusted based on user feedback, adoption metrics, and technical discoveries (e.g., pure-Go spike results). This is a living document, not a fixed commitment.
