# Product Vision

**Last updated**: September 30, 2026

## Problem Statement

Marketing data analysts need to collect tabular and numeric data from various sources (campaigns, web analytics, customer surveys, third-party APIs) and transfer it to AI applications for modeling, forecasting, and personalization. These workflows have friction:

1. **Format fragmentation**: Data arrives in CSV, JSON, SQL dumps, and Excel. AI pipelines often expect structured, typed, multi-dimensional arrays.
2. **Schema loss**: CSVs lack type information. JSON is verbose and slow for large numeric datasets. Relational exports lose array structure.
3. **No intermediate validation**: Analysts package data, send it to ML engineers, then discover schema mismatches or missing datasets days later.
4. **Tool gaps**: Analysts use spreadsheets and BI tools. ML engineers use Python/R. There's no shared viewing/editing layer for data in flight.

**HDF5** is a natural interchange format for numeric/tabular data moving into AI applications: it preserves types, supports multi-dimensional arrays, is widely supported in Python/MATLAB/R scientific stacks, and is efficient for moderate-sized datasets. However, HDF5 has no user-friendly tooling for non-programmers. Analysts cannot easily browse, validate, or make quick corrections without writing Python scripts.

## Vision

**hdf5-agent** is the **HTTP service that makes HDF5 accessible to marketing data analysts**, enabling them to package, browse, validate, and correct datasets before handing them to AI applications. It bridges the gap between analyst workflows (web UIs, REST APIs, catalogs) and the AI/ML ecosystem's data format expectations.

### What Success Looks Like

- A marketing analyst collects campaign performance data from an API, packages it as HDF5, and uses hdf5-agent's UI to confirm the schema before sending it to the ML team.
- A data catalog service (owned by marketing ops) calls hdf5-agent's API to inventory available datasets and present them in a searchable interface.
- An AI application reads HDF5 files prepared by analysts, trusting the schema and types without manual validation, because the analyst verified them through the agent.
- The agent runs as a lightweight Docker container alongside other marketing data infrastructure, with no Python runtime or C library installation burden on end-user machines.

### What This Is Not

- **Not a data warehouse**: hdf5-agent is a file-level service, not a query engine. It doesn't replace SQL or data lakes.
- **Not a full scientific data platform**: It targets moderate-sized marketing datasets (thousands to tens of thousands of rows), not terabyte-scale research data or high-performance computing.
- **Not an AI platform**: It provides the data layer for AI workflows but does not run models, training, or inference.

## Target Users

**Primary persona**: Marketing data analyst at a mid-sized company with an emerging AI/ML capability. They understand spreadsheets, BI tools, and REST APIs, but are not programmers. They need to get data into a format their ML team can consume.

**Secondary persona**: Marketing operations engineer building data catalog and workflow automation. They call hdf5-agent via its REST API and integrate it into broader data pipelines.

**Tertiary consumer**: ML/AI engineers or applications that read the HDF5 files prepared by analysts. They benefit from reliable schemas and types, reducing iteration cycles with marketing teams.

## Product Principles

1. **API-first**: The HTTP API is the contract. The UI is a reference client. Other services should call the API, never link libhdf5 directly.
2. **Opinionated simplicity**: Support integers, floats, and simple multi-dimensional arrays. Skip exotic HDF5 features (compression, chunking, compound types, attributes) unless demanded by real user workflows.
3. **Operations-friendly**: Single binary, Docker-first, environment-only config, health endpoints, structured errors. Easy to deploy in marketing infrastructure that may lack deep DevOps expertise.
4. **Trustworthy interchange**: Files written by the agent must be readable by standard tools (h5py, MATLAB, HDF5 command-line tools). Files from those tools must be browsable/editable by the agent.

## Strategic Alignment

This product aligns with broader data platform goals:

- **Reduce data transfer friction** between marketing and AI teams
- **Enable self-service data workflows** for non-engineers
- **Standardize interchange formats** across the organization
- **Simplify deployment and operations** for data infrastructure

## Success Metrics (Long-Term)

- **Adoption**: Number of marketing teams using hdf5-agent in their AI data transfer workflows
- **Data volume**: Datasets prepared via the agent and successfully consumed by AI applications
- **Reduction in schema iteration cycles**: Fewer back-and-forth exchanges between marketing and ML teams due to data format issues
- **Operational stability**: Uptime and health check success rates in production environments

## Open Questions

- Do analysts need to edit datasets in place, or is browse-only sufficient for initial validation?
- Should the agent support ingestion from CSV/JSON/SQL and export to those formats, or only HDF5 read/write/edit?
- What authentication/authorization is required for multi-tenant marketing data infrastructure?
- How does this fit into a broader marketing data catalog or lineage strategy?

These questions inform roadmap prioritization and are revisited quarterly based on user feedback and partnership discussions.
