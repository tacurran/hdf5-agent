# Glossary

**Purpose**: Define HDF5 and AI data transfer terminology for marketing analysts, new team members, and partners.  
**Audience**: Non-technical users, onboarding materials, external communications  
**Last updated**: September 30, 2026

---

## HDF5 Terminology

### HDF5
**Hierarchical Data Format version 5**. A file format and library for storing and managing large amounts of structured data, especially numeric and tabular datasets. Widely used in scientific computing, machine learning, and data science.

**Why it matters**: HDF5 preserves data types (integers, floats, multi-dimensional arrays), is more efficient than CSV or JSON for numeric data, and is supported by Python (h5py), MATLAB, R, and many ML frameworks.

**File extensions**: `.h5`, `.hdf5`

---

### File
An HDF5 file (e.g., `campaign_data.h5`) is a container that holds groups and datasets in a hierarchical structure, similar to a filesystem with directories and files.

---

### Group
A container within an HDF5 file that organizes datasets, similar to a folder in a filesystem. Groups can contain other groups (nested hierarchy) or datasets.

**Example**: `/impressions/` is a group; it might contain datasets like `daily`, `hourly`, `by_segment`.

---

### Dataset
A multi-dimensional array of data within an HDF5 file. Datasets have:
- **Shape**: Dimensions (e.g., `[365, 5]` for 365 rows and 5 columns)
- **Data type (dtype)**: `int8`, `int16`, `int32`, `int64`, `float32`, `float64`, etc.
- **Values**: The actual numeric data

**Example**: `/impressions/daily` might be a dataset with shape `[365, 5]` (365 days, 5 metrics) and dtype `int64`.

---

### Shape
The dimensions of a dataset, expressed as a list of integers.

**Examples**:
- `[100]`: 1-dimensional array (vector) with 100 elements
- `[365, 5]`: 2-dimensional array (matrix) with 365 rows and 5 columns
- `[10, 20, 30]`: 3-dimensional array (tensor)

---

### Data Type (dtype)
The type of data stored in a dataset. Common types in hdf5-agent:
- **Integers**: `int8`, `int16`, `int32`, `int64` (signed integers of 1, 2, 4, or 8 bytes)
- **Floats**: `float32` (single precision), `float64` (double precision)

**Why it matters**: Preserving data types ensures that campaign IDs remain integers and costs remain floats, avoiding ambiguity from CSV (where everything is text).

---

### Flattened Index
A single integer index into a multi-dimensional dataset, treating it as a 1D array.

**Example**: For a dataset with shape `[365, 5]` (365 rows, 5 columns):
- Flattened index `0` → row 0, column 0
- Flattened index `5` → row 1, column 0 (365 * 5 = 1825 total elements)

**Why it matters**: The hdf5-agent API uses flattened indices for updates (`PUT /api/v1/files/{name}/datasets` with `indices` array).

---

### Attribute
Metadata attached to a dataset or group (e.g., units, description, creation timestamp). 

**Note**: hdf5-agent does not currently support reading or writing attributes (v1.0 scope). This may be added in future versions if users request it.

---

### h5py
The most popular Python library for reading and writing HDF5 files. Part of the scientific Python ecosystem (NumPy, pandas, scikit-learn).

**Example usage**:
```python
import h5py
f = h5py.File('campaign_data.h5', 'r')
dataset = f['/impressions/daily']
print(dataset.shape)  # (365, 5)
print(dataset[:])      # Read all values
f.close()
```

---

### h5dump
A command-line tool (part of the HDF5 software suite) that prints the structure and contents of an HDF5 file in human-readable text.

**Example**:
```bash
h5dump campaign_data.h5
```

**Why it matters**: Used in testing to verify that files written by hdf5-agent are standards-compliant and readable by external tools.

---

### h5diff
A command-line tool that compares two HDF5 files and reports differences in structure or data.

**Example**:
```bash
h5diff expected.h5 actual.h5
```

**Why it matters**: Used in integration tests to validate correctness of hdf5-agent writes.

---

## AI and Data Transfer Terminology

### Data Pipeline
A sequence of steps that collect, transform, and route data from source systems (e.g., marketing APIs, databases) to destination systems (e.g., data warehouses, ML models, BI tools).

**Example**: Marketing pipeline: Google Analytics API → transform → package as HDF5 → deliver to ML model training.

---

### Data Catalog
A searchable inventory of datasets available in an organization, often including metadata (dataset names, schemas, owners, descriptions, lineage).

**Example**: A marketing data catalog lets analysts search for "campaign datasets" and see available HDF5 files, their shapes, and update timestamps.

**hdf5-agent role**: The catalog calls hdf5-agent's API to extract metadata from HDF5 files.

---

### Schema
The structure and data types of a dataset. For HDF5, schema includes:
- Dataset paths (e.g., `/impressions/daily`)
- Shapes (e.g., `[365, 5]`)
- Data types (e.g., `int64`)

**Why it matters**: ML models expect specific schemas. Schema mismatches (wrong shape, wrong type, missing dataset) cause failures.

---

### Schema Validation
Confirming that a dataset matches expected structure and types before using it downstream.

**Example**: An analyst confirms that `/impressions/daily` has shape `[365, 5]` and dtype `int64` before sending the file to the ML team.

**hdf5-agent role**: Provides UI and API for browsing schemas; future roadmap includes automated validation rules.

---

### Data Transfer
Moving data from one system or team to another. In this context: marketing analysts transferring datasets to AI/ML applications.

**Challenge**: Format fragmentation (CSV, JSON, SQL), type loss, lack of validation tools.

**Solution**: Use HDF5 as interchange format; use hdf5-agent for validation and editing.

---

### Interchange Format
A file format used to transfer data between different systems or teams, designed to be widely supported and preserve important properties (types, structure).

**Example**: HDF5 is an interchange format for numeric data going from marketing to AI applications.

---

### Type Preservation
Ensuring that integers remain integers and floats remain floats when data is transferred, rather than converting everything to text (as in CSV).

**Example**: Campaign IDs stored as `int64` in HDF5 are read as integers by Python, not strings.

**Why it matters**: Prevents type errors in downstream processing (e.g., ML models expecting numeric input).

---

### ML (Machine Learning)
A field of AI focused on building models that learn patterns from data to make predictions or decisions.

**Example**: A marketing ML model predicts customer conversion probability based on campaign impressions and clicks.

**Data needs**: Typed, structured numeric data (features); HDF5 is a natural format.

---

### AI Application
A software application that uses AI models to provide insights, predictions, or automation.

**Example**: A personalization engine that recommends content to users based on their behavior, trained on marketing data.

**Data needs**: Reliable schemas, correct types, validated inputs.

---

## hdf5-agent Terminology

### hdf5-agent
The HTTP service (Go backend + React UI + JSON API) for browsing, validating, and editing HDF5 files. Designed for marketing data analysts preparing datasets for AI applications.

---

### `/api/v1`
The versioned REST API provided by hdf5-agent. Endpoints include:
- `GET /api/v1/files` — list files
- `GET /api/v1/files/{name}` — get file structure
- `GET /api/v1/files/{name}/datasets?path=...` — read dataset
- `PUT /api/v1/files/{name}/datasets` — update dataset

---

### OpenAPI Spec
A machine-readable contract (YAML file) describing the hdf5-agent API: endpoints, request/response schemas, error codes.

**Location**: `api/openapi.yaml` in the repository; also served at `GET /api/v1/openapi.yaml` on a running instance.

**Why it matters**: Enables code generation for client libraries, API documentation, and integration testing.

---

### Request ID
A unique identifier (UUID) for each HTTP request, used to correlate logs across services.

**Header**: `X-Request-ID`

**Example**: Client sends `X-Request-ID: abc-123`; hdf5-agent logs all operations for that request with the same ID; errors include the ID in the response.

---

### Structured Error
A JSON error response with stable error codes and messages, following the envelope format:

```json
{
  "error": {
    "code": "not_found",
    "message": "file not found: missing.h5",
    "request_id": "abc-123"
  }
}
```

**Error codes**: `not_found`, `invalid_path`, `invalid_name`, `conflict`, `too_large`, `unsupported_type`, `internal`

---

### Health Endpoint
An HTTP endpoint that returns the service's liveness or readiness status.

**Endpoints**:
- `/healthz` — liveness (200 if process is running)
- `/readyz` — readiness (200 if data directory is readable)

**Why it matters**: Used by load balancers, Kubernetes, and orchestrators to monitor service health.

---

### pkg/hdf5client
A typed Go client library for calling the hdf5-agent API from other Go services.

**Example usage**:
```go
import "github.com/tacurran/hdf5-agent/pkg/hdf5client"

client := hdf5client.New("http://localhost:8080")
files, err := client.ListFiles(ctx)
```

**Why it matters**: Simplifies integration for Go-based data catalogs, pipelines, and automation.

---

### CGO (C-Go)
A feature of the Go programming language that allows Go code to call C libraries. hdf5-agent currently uses CGO to call the HDF5 C library (libhdf5) via `gonum.org/v1/hdf5`.

**Trade-off**: CGO adds build complexity (compiler flags, distro dependencies). hdf5-agent is evaluating a pure-Go HDF5 library to remove CGO.

**User impact**: None (internal implementation detail). Pure-Go migration would simplify deployment but not change the API or UI.

---

## Common Acronyms

| Acronym | Meaning |
|---------|---------|
| **HDF5** | Hierarchical Data Format version 5 |
| **API** | Application Programming Interface |
| **REST** | Representational State Transfer (HTTP API style) |
| **JSON** | JavaScript Object Notation (data format) |
| **CSV** | Comma-Separated Values (text file format) |
| **ML** | Machine Learning |
| **AI** | Artificial Intelligence |
| **UI** | User Interface |
| **UX** | User Experience |
| **BI** | Business Intelligence |
| **CRUD** | Create, Read, Update, Delete (operations) |
| **CI/CD** | Continuous Integration / Continuous Deployment |
| **CGO** | C-Go (Go's C interop feature) |
| **MVP** | Minimum Viable Product |
| **PRD** | Product Requirements Document |
| **JTBD** | Jobs-to-be-Done (product framework) |

---

## Usage Notes

### For Analysts
You don't need to memorize these terms to use hdf5-agent! The UI is designed to be intuitive. Key concepts:
- **File**: Container (like a folder)
- **Dataset**: Table or array of numbers
- **Shape**: Number of rows and columns

That's enough to browse and validate your data.

### For Ops/Engineers
Refer to this glossary when writing integration code or documentation. Link to it in onboarding materials.

### For External Partners
Share this glossary (or a subset) in partnership discussions to establish shared vocabulary.

---

## Updates

Add terms as needed based on:
- User feedback (e.g., "What does 'flattened index' mean?")
- New features (e.g., if we add CSV ingestion, define "CSV")
- Partner questions (e.g., external stakeholders unfamiliar with HDF5)

Review quarterly; keep definitions concise and jargon-free.
