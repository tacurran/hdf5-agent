# Personas and Jobs-to-be-Done

**Last updated**: September 30, 2026

This document describes the target user personas for hdf5-agent and the jobs they are trying to get done. Understanding these personas guides feature prioritization, UX decisions, and messaging.

---

## Persona 1: Marketing Data Analyst (Alex)

### Background

- **Role**: Data analyst in the marketing department at a mid-sized B2B or B2C company
- **Team size**: Part of a 3-5 person analytics team supporting marketing campaigns, attribution, and experimentation
- **Skills**: Strong in Excel, SQL, Tableau/Looker. Can call REST APIs with curl or Postman. Not a programmer; doesn't write Python or R regularly.
- **Tools**: Spreadsheets, BI platforms, marketing automation APIs (Google Analytics, Salesforce Marketing Cloud, Facebook Ads, etc.), data export scripts maintained by ops

### Responsibilities

- Pull campaign performance data (impressions, clicks, conversions, spend) from marketing platforms
- Clean, aggregate, and package data for analysis or handoff to ML/BI teams
- Validate data quality (schema correctness, no missing/duplicate records) before sharing
- Make small corrections (fix typo, adjust outlier values) without waiting for engineering

### Pain Points

- **Format chaos**: Data from APIs is JSON or CSV; ML team needs HDF5 or NumPy. Alex exports CSV, someone else converts it, loses type information, iterates.
- **No validation tool**: After packaging data, Alex can't easily confirm schema or spot-check values without writing Python (which Alex isn't comfortable doing).
- **Slow feedback loop**: Sends data to ML team, waits days to hear "wrong schema" or "missing dataset", has to re-export and resend.
- **Dependency on engineering**: For small corrections (fixing a typo, adjusting a date offset), Alex has to file a ticket and wait.

### Goals

- Package marketing data into a format AI applications can consume without schema issues
- Validate dataset structure and values before handing off to ML team
- Make minor corrections quickly without writing code
- Understand what's in an HDF5 file without asking a developer

### Ideal Experience with hdf5-agent

1. Alex exports campaign data to HDF5 using a script maintained by ops (or eventually via agent ingestion)
2. Alex opens the agent UI, selects `campaign_q3.h5`, and sees the tree structure of datasets
3. Alex clicks `/impressions/daily` and sees shape, type, and a preview of values
4. Alex spots a typo in campaign ID at index 42, clicks Edit, changes the value, saves
5. Alex confirms the file structure matches ML team's expectations and sends the file confidently
6. ML team reads the file without issues; no back-and-forth iteration needed

---

## Persona 2: Marketing Operations Engineer (Jamie)

### Background

- **Role**: Data engineer or marketing ops engineer responsible for data pipelines, automation, and internal tools
- **Team size**: Part of a 2-4 person ops team supporting marketing analytics and campaign operations
- **Skills**: Proficient in Python, Go, SQL. Comfortable with Docker, CI/CD, cloud infrastructure. Familiar with data catalogs, ETL workflows, and API design.
- **Tools**: Airflow or similar scheduler, Docker Compose, internal data catalog, REST APIs, SQL databases, S3/GCS for storage

### Responsibilities

- Build and maintain data pipelines that collect, transform, and route marketing data
- Operate a data catalog that indexes available datasets and makes them searchable for analysts and ML teams
- Integrate HDF5 files into the catalog and broader marketing data platform
- Ensure data infrastructure is reliable, scalable, and easy for analysts to use

### Pain Points

- **Library complexity**: Linking libhdf5 in multiple services is painful (CGO flags, distro dependencies, version mismatches). Wants a single HTTP service to handle all HDF5 access.
- **No standard API**: Ad-hoc Python scripts to list/read HDF5 files don't compose well with Go microservices or catalog automation.
- **Operational burden**: Python runtime + h5py dependencies in Docker images increase image size and deployment complexity.
- **Metadata extraction**: Needs to programmatically extract dataset names, shapes, and types to populate the catalog search index.

### Goals

- Provide a stable, versioned API that other services can call to interact with HDF5 files
- Automate metadata extraction for the data catalog
- Simplify deployment (single Docker container, no Python or C library management)
- Enable analysts to self-serve (browse, validate, edit) without filing ops tickets

### Ideal Experience with hdf5-agent

1. Jamie deploys hdf5-agent in a Docker Compose stack alongside the data catalog service
2. The catalog calls `GET /api/v1/files` to list available HDF5 files
3. For each file, the catalog calls `GET /api/v1/files/{name}` to get the tree structure and dataset metadata
4. Jamie indexes this in Elasticsearch; analysts can now search for "campaign datasets with >30 days of history"
5. The agent's health endpoints integrate with the load balancer; uptime is tracked in ops dashboards
6. Jamie uses the Go client library (`pkg/hdf5client`) to call the API from the catalog service, avoiding raw HTTP

---

## Persona 3: ML/AI Engineer (Morgan)

### Background

- **Role**: Machine learning engineer building predictive models for marketing personalization, attribution, or forecasting
- **Team size**: Part of a 2-3 person ML team serving marketing and product
- **Skills**: Expert in Python (pandas, NumPy, scikit-learn, TensorFlow/PyTorch). Comfortable with HDF5 (h5py), Jupyter notebooks, and model deployment pipelines.
- **Tools**: Jupyter, h5py, pandas, scikit-learn, TensorFlow, Docker, model serving infrastructure

### Responsibilities

- Ingest data from marketing analysts
- Build, train, and validate models
- Deploy models to production; monitor performance
- Provide feedback to marketing on data quality and schema requirements

### Pain Points

- **Schema mismatches**: Receives HDF5 files with unexpected dataset names, wrong shapes, or incorrect data types. Wastes days diagnosing and asking marketing to fix and resend.
- **Type errors**: Marketing exports numeric IDs as floats instead of ints; models break.
- **Missing datasets**: Expected dataset isn't present; has to email back and forth.
- **Manual validation**: Has to write Python scripts to inspect files and validate schemas before trusting them in training pipelines.

### Goals

- Receive HDF5 files with predictable schemas and correct types
- Reduce iteration cycles with marketing analysts over schema issues
- Trust that files have been validated before handoff
- Focus on model development, not data debugging

### Benefit from hdf5-agent

Morgan doesn't directly use hdf5-agent, but benefits indirectly:
- Alex (marketing analyst) validates files before sending them to Morgan
- Morgan gets files with reliable schemas, correct types, and expected datasets
- Fewer back-and-forth cycles over data quality issues
- Morgan can trust the data and focus on modeling

---

## Jobs-to-be-Done (JTBD)

The JTBD framework describes the functional goals users are trying to achieve, independent of specific features. These guide product decisions and feature prioritization.

### Job 1: Validate Data Before Handoff

**When** an analyst packages data for the ML team  
**I want to** confirm the file structure, dataset names, shapes, types, and sample values  
**So that** the ML team can ingest it without schema errors, and I avoid time-wasting back-and-forth

**Current workarounds**:
- Ask an ML engineer to write a Python script to inspect the file (slow, dependency on eng)
- Send the file and hope for the best; fix issues when ML reports them (wasteful iteration)

**How hdf5-agent solves it**:
- UI shows tree structure, dataset metadata, and value previews
- Analyst confirms schema against ML team's spec document
- Confident handoff without writing code

---

### Job 2: Quickly Correct Data Errors

**When** I discover a typo or wrong value in a dataset  
**I want to** fix it in place without re-running the entire export pipeline  
**So that** I can deliver corrected data in minutes instead of hours or days

**Current workarounds**:
- Re-run the export script (slow, may take hours if source system is slow)
- Ask a developer to write a Python script to patch the value (dependency on eng)
- Edit the CSV before converting to HDF5, then reconvert (error-prone)

**How hdf5-agent solves it**:
- UI allows editing individual dataset values by index
- `PUT /api/v1/files/{name}/datasets` updates values in place
- Changes persist; standard tools see the corrected values

---

### Job 3: Automate Metadata Inventory

**When** I need to build a searchable data catalog of available datasets  
**I want to** programmatically extract file names, dataset paths, shapes, and types  
**So that** analysts and ML engineers can discover datasets without manually browsing directories

**Current workarounds**:
- Write Python scripts with h5py to walk files and extract metadata (bespoke, brittle)
- Manually maintain a spreadsheet or wiki of available files (stale, error-prone)

**How hdf5-agent solves it**:
- REST API (`GET /api/v1/files`, `GET /api/v1/files/{name}`) returns structured JSON metadata
- Ops engineer calls the API from the catalog service
- Catalog indexes the metadata in a search backend (e.g., Elasticsearch)

---

### Job 4: Ensure Interoperability with ML Tools

**When** I package data for an ML engineer  
**I want to** guarantee the file is readable by h5py, MATLAB, and other standard HDF5 tools  
**So that** the ML team doesn't encounter "corrupt file" or "unsupported format" errors

**Current workarounds**:
- Test every file with `h5dump` or h5py before sending (manual, time-consuming)
- Hope the export tool is correct (risky; errors discovered late)

**How hdf5-agent solves it**:
- Agent writes standards-compliant HDF5; integration tests validate with h5dump/h5diff
- Files written by the agent are guaranteed to be readable by standard tools
- Reduces risk of format incompatibility

---

### Job 5: Deploy Data Infrastructure Without Complexity

**When** I deploy the HDF5 service in staging or production  
**I want to** run a single Docker container with environment-only config  
**So that** I don't have to manage Python runtimes, C library versions, or complex build toolchains

**Current workarounds**:
- Deploy Python + h5py + Flask; deal with virtualenvs, system libhdf5 versions, wheel build issues
- Maintain bespoke Dockerfiles with multi-stage Python builds (slow, large images)

**How hdf5-agent solves it**:
- Single Go binary in a multi-stage Docker image
- Non-root runtime user (uid 65532)
- Environment-only config (no secrets, no files to mount beyond the data directory)
- Health and readiness endpoints for load balancer integration

---

## Summary Table

| Persona | Primary JTBD | Top Pain Point | Top Benefit from hdf5-agent |
|---------|--------------|----------------|------------------------------|
| **Alex** (Analyst) | Validate data before handoff | No tool to inspect HDF5 files without Python | UI to browse, validate, and edit datasets |
| **Jamie** (Ops) | Automate metadata inventory | No standard API; library complexity in each service | REST API + Go client; single Docker container |
| **Morgan** (ML) | Ingest reliable schemas | Schema mismatches from marketing | Indirect: fewer schema errors because Alex validated |

---

## Using These Personas

- **Feature prioritization**: When choosing between features, ask "Which persona's top JTBD does this serve?"
- **UX decisions**: Design the UI for Alex (non-programmer, Excel mindset), not for Morgan (Python expert)
- **API design**: Design the API for Jamie (ops automation, catalog integration), with structured errors and OpenAPI docs
- **Messaging**: Speak to Alex and Jamie directly; position Morgan as a beneficiary, not a primary user

---

## Open Questions for User Research

- How often do analysts need to edit datasets vs. just browse them?
- What is the typical size of marketing datasets? (Rows, datasets per file, files per project)
- Do analysts need to create entirely new datasets via UI, or only edit existing ones?
- What authentication/authorization is needed for multi-team marketing environments?
- Are there other data formats analysts need to convert to/from HDF5 (CSV, JSON, Parquet)?

Answers to these questions inform roadmap prioritization and are updated quarterly based on user interviews and feedback.
