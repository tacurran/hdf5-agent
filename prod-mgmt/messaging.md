# Messaging and Positioning Guide

**Purpose**: Consistent messaging for internal and external communication about hdf5-agent.  
**Audience**: PM, marketing, sales, support, engineering (when speaking externally)  
**Last updated**: September 30, 2026

---

## Positioning Statement

**hdf5-agent** is the **HTTP service that makes HDF5 accessible to marketing data analysts**, bridging the gap between analyst workflows (spreadsheets, BI tools, REST APIs) and AI application data expectations. It provides a web UI and REST API for browsing, validating, and editing HDF5 files without requiring analysts to write code.

### Key Elements

- **What it is**: HTTP service (Go backend + React UI + JSON API) for HDF5 file interaction
- **Who it's for**: Marketing data analysts (primary); marketing ops engineers (secondary); ML engineers (indirect beneficiary)
- **What problem it solves**: Data transfer friction between marketing and AI teams; schema validation before handoff; lack of HDF5 tooling for non-programmers
- **Why HDF5**: Efficient, typed, widely supported in AI/ML stacks; better than CSV/JSON for numeric data
- **Key differentiation**: API-first (other services call it, don't link libhdf5); operations-friendly (single Docker container, no Python runtime); interoperable (files work with h5py, MATLAB, standard tools)

---

## What to Say / What Not to Say

| Context | What to Say | What NOT to Say |
|---------|-------------|-----------------|
| **Problem** | "Marketing analysts struggle to package and validate data for AI pipelines. CSV loses types, JSON is verbose, and HDF5 has no user-friendly tools." | ❌ "HDF5 is better than all other formats" (too broad; depends on use case) |
| **Solution** | "hdf5-agent is an HTTP service with a web UI and REST API that lets analysts browse, validate, and edit HDF5 files without writing Python code." | ❌ "hdf5-agent is an HDF5 GUI" (too narrow; misses API-first architecture) |
| **Target user** | "Marketing data analysts who need to prepare datasets for AI applications but aren't programmers." | ❌ "Anyone who uses HDF5" (too broad; scientific HPC users have different tools) |
| **Why Go** | "Go backend delivers a single binary, fast startup, and operations-friendly deployment. No Python runtime or C library management in production." | ❌ "Go is faster than Python" (true but not the main point; focus on ops simplicity) |
| **HDF5 library** | "Currently uses CGO bindings to libhdf5; evaluating pure-Go implementation to remove build dependencies." | ❌ "We're migrating because CGO is bad" (sounds like a problem; frame as improvement opportunity) |
| **API** | "Versioned JSON API (`/api/v1`) with OpenAPI spec; structured errors; health endpoints; Go client library (`pkg/hdf5client`)." | ❌ "RESTful API" (buzzword; be specific about versioning and OpenAPI contract) |
| **Interoperability** | "Files written by hdf5-agent are standards-compliant HDF5, readable by h5py, MATLAB, and HDF5 command-line tools." | ❌ "Proprietary HDF5 format" (false; we write standard HDF5) |
| **Roadmap** | "Near-term: CSV ingestion/export, authentication for multi-tenant use, pure-Go HDF5. Future: workflow automation, data lineage." | ❌ "We'll support every HDF5 feature eventually" (scope creep; stay opinionated) |

---

## Elevator Pitch (30 seconds)

**For analysts**:
> "hdf5-agent is a web tool that lets you open, browse, and validate HDF5 files before sending them to the ML team — no Python required. Click a file, see the datasets, spot-check values, fix typos, and hand off data confidently."

**For ops/engineers**:
> "hdf5-agent is an HTTP service with a JSON API for HDF5 file access. Instead of linking libhdf5 in every service, call our API. It's a single Docker container, environment-only config, health endpoints, OpenAPI spec, and a Go client library."

**For leadership**:
> "hdf5-agent reduces schema iteration cycles between marketing and AI teams by giving analysts a self-service tool to validate and correct datasets before handoff. It's production-ready, open source, and integrates with our data catalog."

---

## Messaging Pillars

### 1. Self-Service Data Validation
**Message**: Analysts can validate dataset schemas and values before handing off to ML teams, reducing back-and-forth cycles and time-to-insights.

**Support points**:
- Web UI for browsing file structure, dataset shapes, and value previews
- Edit individual values in place without re-running export pipelines
- Confirm schema matches ML team expectations before handoff

**Use in**: Analyst onboarding, marketing materials, demo scripts

---

### 2. API-First Architecture
**Message**: Other services call hdf5-agent's REST API rather than linking the HDF5 C library, simplifying dependency management and enabling data catalog integration.

**Support points**:
- Versioned JSON API (`/api/v1`) with OpenAPI 3 spec
- Structured errors with stable error codes (`not_found`, `invalid_path`, etc.)
- Go client library (`pkg/hdf5client`) for downstream services
- Health and readiness endpoints for load balancer integration

**Use in**: Ops/eng documentation, partner integration guides, architecture reviews

---

### 3. Operations-Friendly Deployment
**Message**: Single Docker container, environment-only config, no secrets, no Python runtime, fast startup — easy to deploy and maintain in marketing infrastructure.

**Support points**:
- Multi-stage Docker build: single Go binary + libhdf5 runtime libs (~40MB image; ~15MB if pure-Go adopted)
- Non-root runtime user (uid 65532)
- No config files; all environment variables
- Health endpoints (`/healthz`, `/readyz`) for orchestrator monitoring

**Use in**: Deployment docs, ops reviews, infrastructure planning

---

### 4. Standards-Compliant Interoperability
**Message**: Files written by hdf5-agent are guaranteed to be readable by h5py, MATLAB, and standard HDF5 tools. Files from those tools are browsable by the agent.

**Support points**:
- Integration tests validate round-trip correctness with h5dump, h5diff
- Writes standards-compliant HDF5 (no proprietary extensions)
- Reads typical HDF5 files from Python, MATLAB, C tools

**Use in**: ML engineer onboarding, external partner communications, quality assurance messaging

---

## Audience-Specific Messaging

### For Marketing Analysts
**Tone**: Practical, non-technical, problem-focused

**Key messages**:
- "Confirm your data is correct before sending it to the ML team"
- "Fix typos without re-running your export script"
- "No Python required — just open the file in your browser"

**Avoid**: Technical jargon (CGO, libhdf5, mutex, OpenAPI); focus on workflows

---

### For Marketing Ops / Data Engineers
**Tone**: Technical, pragmatic, integration-focused

**Key messages**:
- "Call one API instead of linking libhdf5 in each service"
- "Single Docker container, health endpoints, OpenAPI spec"
- "Go client library for downstream services; structured errors for automation"

**Avoid**: Overselling features not yet built (e.g., don't promise auth if it's not shipped)

---

### For ML Engineers
**Tone**: Indirect beneficiary; focus on data quality

**Key messages**:
- "Marketing analysts validate schemas before handoff, so you get reliable data"
- "Fewer back-and-forth cycles over schema mismatches"
- "Files are standards-compliant HDF5, readable by h5py"

**Avoid**: Positioning the agent as an ML tool (it's not; it's upstream data prep)

---

### For Leadership
**Tone**: Strategic, outcome-focused, metrics-driven

**Key messages**:
- "Reduces time-to-insights by cutting schema iteration cycles"
- "Enables self-service data workflows for marketing analysts"
- "Production-ready, open source, integrates with data catalog"

**Avoid**: Deep technical details; focus on adoption, metrics, ROI

---

## Competitive Positioning

### vs. Manual Python Scripts (h5py)
**They say**: "We can write a Python script to inspect HDF5 files."  
**We say**: "Scripts require Python expertise and are hard for non-programmers. hdf5-agent provides a web UI and REST API that analysts and ops can use without writing code. It's also easier to integrate into data catalogs and pipelines."

### vs. HDFView (Java desktop app)
**They say**: "HDFView is the official HDF5 GUI."  
**We say**: "HDFView is a desktop app, not a web service. hdf5-agent provides a REST API for integration, runs in Docker, and is designed for analysts in marketing workflows, not scientific HPC users."

### vs. Jupyter + h5py
**They say**: "We can use Jupyter notebooks to explore HDF5 files."  
**We say**: "Jupyter is great for ML engineers, but marketing analysts aren't comfortable with notebooks. hdf5-agent provides a no-code UI for analysts, plus an API for ops automation."

### vs. Building HDF5 Access Into Each Service
**They say**: "We can link libhdf5 in each service that needs it."  
**We say**: "That's painful: CGO flags, distro dependencies, version mismatches, build complexity. hdf5-agent centralizes HDF5 access; other services call the API. It's simpler to deploy and maintain."

---

## Messaging Do's and Don'ts

### Do's
- ✅ **Be specific**: "REST API with OpenAPI 3 spec" not "has an API"
- ✅ **Lead with user value**: "Analysts can validate data before handoff" not "We built a Go backend"
- ✅ **Ground claims in the product**: If you say "standards-compliant," point to h5dump/h5diff tests
- ✅ **Use personas**: Speak to Alex (analyst), Jamie (ops), Morgan (ML) by name in internal discussions
- ✅ **Link to docs**: README, QUICKSTART, OpenAPI spec, examples/data-catalog

### Don'ts
- ❌ **Don't hype unbuilt features**: If CSV ingestion isn't shipped, don't claim it in external materials
- ❌ **Don't oversell HDF5**: It's not the best format for everything; it's good for numeric/tabular data going to AI
- ❌ **Don't trash competitors**: "HDFView is old" → "hdf5-agent is web-based and API-first"
- ❌ **Don't use jargon externally**: CGO, mutex, libhdf5, gonum → "Go backend with HDF5 library"
- ❌ **Don't promise dates**: "Available Q2" → "Near-term roadmap" or "In development"

---

## Sample Messaging Snippets

### Social Media / Blog Post Intro
> "Introducing hdf5-agent: an open-source HTTP service that makes HDF5 files accessible to data analysts. Browse, validate, and edit datasets via a web UI or REST API — no Python required. Perfect for marketing teams preparing data for AI pipelines. Apache 2.0 licensed. 🚀"

### Internal Announcement (Slack/Email)
> "📢 New tool alert: hdf5-agent is now available for alpha testing! If your team packages data for AI applications, this HTTP service (web UI + REST API) lets you validate schemas, preview datasets, and fix errors before handoff. Reduces back-and-forth with ML teams. Docker deployment, OpenAPI spec, Go client library. Check out the README and let us know if you'd like to try it!"

### Partner Email
> "Hi [Partner],
> 
> We've built hdf5-agent, an HTTP service for HDF5 file management, and think it could complement your data catalog. It provides a REST API (`/api/v1`) to list files, extract metadata (dataset names, shapes, types), and read/update datasets. We'd love to explore integration opportunities.
> 
> Quick links:
> - OpenAPI spec: [link]
> - Example catalog integration: [examples/data-catalog]
> - Try it with Docker: [README]
> 
> Let's schedule a call if you're interested!"

### Support / FAQ Response
> "Q: Can I use hdf5-agent with files from MATLAB or Python?
> A: Yes! hdf5-agent reads and writes standards-compliant HDF5. Files from MATLAB, Python (h5py), or HDF5 command-line tools are fully supported. Files written by the agent are readable by those tools."

---

## Talking Points for Demos

### Live Demo Script (5 minutes)

1. **Open the UI** (http://localhost:8080): "This is the hdf5-agent web UI. No installation, just open your browser."
2. **List files**: "Here's the data directory. Let's click `campaign_data.h5`."
3. **Browse structure**: "You see groups and datasets hierarchically. Click `/impressions/daily`."
4. **Preview dataset**: "Shape: [365, 5]. Type: int64. Here are the first 100 values. I can scroll or search."
5. **Edit a value**: "Oops, campaign ID at index 42 is wrong. Click Edit, change `1000` to `1001`, Save. Done in 5 seconds."
6. **API call** (curl): "Now let's call the API. `GET /api/v1/files` returns JSON. Here's the file list. And here's the dataset read endpoint. This is how our data catalog queries it."
7. **Interoperability**: "Let's open this file in Python: `h5py.File('campaign_data.h5')`. Yep, the edit we made is there. Standard HDF5."

**Key takeaway**: "Self-service validation for analysts, REST API for ops automation, interoperable with standard tools."

---

## Revision Notes

Update this document when:
- Major features ship (update "What to Say" with new capabilities)
- Messaging tests reveal confusion (adjust language)
- Competitive landscape changes (update "vs." section)
- User feedback suggests better framing (iterate on personas/workflows)

Keep this as a living guide. Review quarterly with PM, marketing, and eng leads.

---

## Cross-References

- **Vision**: [vision.md](./vision.md) — problem statement and strategic alignment
- **PRD**: [prd.md](./prd.md) — use cases, acceptance criteria
- **Personas**: [personas-and-jtbd.md](./personas-and-jtbd.md) — user profiles and jobs-to-be-done
- **External overview**: [external-overview.md](./external-overview.md) — partner-facing one-pager
- **Roadmap**: [roadmap.md](./roadmap.md) — feature priorities and horizons
