# HDF5 Agent - Architecture & Design Document

## System Overview

HDF5 Agent is a full-stack application designed to manipulate HDF5 files through a modern web interface. The system follows a client-server architecture with a Go backend and React frontend.

```
┌─────────────────────────────────────────────────────────────┐
│                    Browser Environment                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  React Frontend (Port 3000)                         │   │
│  │  ├─ File Browser & Navigation                      │   │
│  │  ├─ HDF5 Structure Tree View                       │   │
│  │  ├─ Data Viewer & Editor                           │   │
│  │  └─ Modal Dialogs & Settings                       │   │
│  └─────────────────────────────────────────────────────┘   │
│                           ↓ (REST API)                       │
├─────────────────────────────────────────────────────────────┤
│                    Network Boundary                          │
│                      (HTTP/REST)                             │
├─────────────────────────────────────────────────────────────┤
│                   Go Server Environment                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  HTTP Server (Port 8080)                            │   │
│  │  ├─ Gorilla Mux Router                             │   │
│  │  ├─ CORS Middleware                                │   │
│  │  └─ Request Handlers                               │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │  HDF5 Interface Layer                              │   │
│  │  ├─ File Operations                                │   │
│  │  ├─ Structure Navigation                           │   │
│  │  ├─ Data Read/Write                                │   │
│  │  └─ Attribute Management                           │   │
│  ├─────────────────────────────────────────────────────┤   │
│  │  Data Serialization                                │   │
│  │  ├─ JSON Encoding/Decoding                         │   │
│  │  ├─ Type Conversion                                │   │
│  │  └─ Shape & Metadata Handling                      │   │
│  └─────────────────────────────────────────────────────┘   │
│                           ↓ (File I/O)                       │
├─────────────────────────────────────────────────────────────┤
│                  Filesystem & HDF5 Layer                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  HDF5 Files (.h5, .hdf5)                            │   │
│  │  ├─ File Container                                 │   │
│  │  ├─ Group Hierarchy                                │   │
│  │  ├─ Datasets (Arrays/Matrices)                     │   │
│  │  └─ Attributes (Metadata)                          │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Backend Architecture

### Component Structure

#### 1. Server (main.go)
- **Responsibility:** Orchestrate HTTP server and request routing
- **Key Types:**
  - `Server` - Main server struct with router and config
  - `DatasetInfo` - HDF5 dataset metadata
  - `GroupInfo` - HDF5 group metadata
  - `FileStructure` - Complete file tree structure
  - `DatasetData` - Dataset values with metadata
  - `UpdateRequest` - Data modification payload

#### 2. Request Handlers

**File Operations:**
- `listFiles()` - Enumerate HDF5 files in working directory
- `openFile()` - Initialize file and build structure
- `getFileStructure()` - Retrieve hierarchical file layout
- `createFile()` - Create new empty HDF5 file
- `deleteFile()` - Remove file from filesystem

**Data Operations:**
- `readDataset()` - Fetch dataset values and metadata
- `updateDataset()` - Modify dataset values with validation
- `sliceDataset()` - Extract hyperslab for large datasets
- `health()` - Health check for monitoring

#### 3. HDF5 Integration

Uses **gonum/hdf5** library for native HDF5 support:
```go
import "gonum.org/v1/hdf5"

// File operations
file, _ := hdf5.OpenFile(path, hdf5.F_ACC_RDWR)
dataset := file.OpenDataset(path)

// Data operations
dataset.Read(&data)
dataset.Write(&data)

// Structure traversal
file.VisitObjects(func(name string) error {
    obj := file.Object(name)
    // Handle groups, datasets
})
```

#### 4. Data Type Handling

**Supported Types:**
- **Float:** Converted to `[]float64`, JSON number
- **Integer:** Converted to `[]int64`, JSON number
- **String:** Converted to `[]string`, JSON string

**Conversion Logic:**
```go
switch dtype.Class() {
case hdf5.H5T_FLOAT:
    data := make([]float64, npoints)
    dataset.Read(&data)
    return data
case hdf5.H5T_INTEGER:
    data := make([]int64, npoints)
    dataset.Read(&data)
    return data
case hdf5.H5T_STRING:
    data := make([]string, npoints)
    dataset.Read(&data)
    return data
}
```

### API Endpoints

**Resource Paths:**
```
/api/
├── files                          [GET]   List HDF5 files
├── file/
│   ├── open                       [POST]  Open and structure
│   ├── structure                  [GET]   Get file tree
│   ├── create                     [POST]  Create new file
│   └── delete                     [DELETE] Delete file
└── dataset/
    ├── read                       [GET]   Read dataset
    ├── update                     [POST]  Modify values
    └── slice                      [POST]  Get hyperslab
```

### Error Handling

**Strategy:**
- Validate input parameters
- Return HTTP status codes (4xx, 5xx)
- Include error messages in response body
- Log errors for debugging

**Common Errors:**
- 400 Bad Request - Missing/invalid parameters
- 404 Not Found - File/dataset doesn't exist
- 500 Internal Server Error - HDF5 operation failure

## Frontend Architecture

### Component Hierarchy

```
App
├── Header
│   ├── Title
│   └── CreateFileButton
├── ErrorBanner
├── Container
│   ├── Sidebar
│   │   ├── FileList
│   │   └── FileItem (×N)
│   └── MainContent
│       ├── TreePane
│       │   ├── TreeView
│       │   └── TreeNode (×N recursive)
│       └── DataPane
│           ├── DataHeader
│           └── DataViewer
│               └── DataCell (×N)
└── Modals
    └── NewFileModal
```

### React Components

#### 1. TreeNode (Recursive)
Renders hierarchical HDF5 structure with expansion.

**Props:**
- `node` - HDF5 node (file, group, dataset)
- `path` - Full path to node
- `onSelect` - Selection callback
- `selectedPath` - Currently selected path
- `level` - Recursion depth for indentation

**Features:**
- Lazy expansion with toggle state
- Icon indicators (file, folder, database)
- Metadata badges (shape, type)
- Click to select datasets

#### 2. DataViewer
Displays dataset contents with edit capability.

**Features:**
- Grid layout for array data
- Cell-level editing
- Type-safe input validation
- Real-time sync with backend
- Truncation for large datasets

#### 3. App (Main)
Orchestrates entire application state and API calls.

**State Management:**
- `files` - Available HDF5 files
- `selectedFile` - Currently open file
- `fileStructure` - File tree structure
- `selectedDataset` - Active dataset
- `datasetData` - Dataset contents
- `loading` - Loading state
- `error` - Error messages

**Methods:**
- `loadFiles()` - Fetch file list
- `loadFileStructure()` - Build tree
- `loadDataset()` - Fetch data
- `handleUpdateData()` - Write changes
- `handleCreateFile()` - Create new
- `handleDeleteFile()` - Remove file

### Styling Architecture

**Design System:**
- CSS Variables for theming
- Scientific instrument aesthetic
- Dark mode (charcoal + neon cyan)
- Responsive breakpoints

**Color Palette:**
```css
Primary:    #00d9ff (Cyan - Data, Accents)
Secondary:  #fbbf24 (Amber - Metadata)
Danger:     #ef4444 (Red - Destructive)
Success:    #10b981 (Green - Confirmation)
```

**Typography:**
- Display: IBM Plex Mono (technical focus)
- UI: Segoe UI (clarity)
- Data: SF Mono (consistency)

**Spacing System:**
- Base unit: 0.25rem (4px)
- Increments: xs (2), sm (4), md (6), lg (8), xl (12), 2xl (16)

### State Flow

```
User Action
    ↓
Event Handler (onClick, onChange, etc.)
    ↓
API Call (fetch)
    ↓
Backend Processing
    ↓
HTTP Response
    ↓
State Update (setState)
    ↓
Re-render Components
    ↓
UI Update
```

## Data Flow Examples

### Reading a Dataset

```
User Clicks Dataset
    ↓
handleSelectDataset()
    ↓
loadDataset(filePath, datasetPath)
    ↓
fetch(/api/dataset/read?file=...&path=...)
    ↓
Backend: readDataset()
    ├─ Open HDF5 file
    ├─ Get dataset
    ├─ Read data values
    ├─ Serialize to JSON
    └─ Return DatasetData
    ↓
setDatasetData(response)
    ↓
DataViewer renders data
```

### Updating a Value

```
User Edits Cell
    ↓
handleUpdateData(index, newValue)
    ↓
fetch(/api/dataset/update, POST)
    ├─ Request: { path, data, indices }
    └─ Body: JSON serialized
    ↓
Backend: updateDataset()
    ├─ Open file (read-write)
    ├─ Get dataset
    ├─ Write values
    ├─ Close file
    └─ Return success
    ↓
loadDataset() (refresh)
    ↓
DataViewer updates
```

## Design Decisions

### 1. Go Backend
**Why:**
- Fast compilation and execution
- Excellent for I/O-bound operations
- Native HDF5 bindings available
- Simple deployment (single binary)

**Alternative:** Python/Flask
- More HDF5 libraries, but slower
- Larger runtime requirements

### 2. React Frontend
**Why:**
- Component-based architecture
- Real-time UI updates
- Rich ecosystem
- Excellent developer experience

**Alternative:** Vue/Svelte
- Smaller bundle size, but less adoption

### 3. REST API (not GraphQL)
**Why:**
- Simple to implement
- Good for CRUD operations
- Easier to debug and monitor
- Standard HTTP methods map well

**Alternative:** GraphQL
- Over-engineering for HDF5 use case
- Increased complexity

### 4. File-based Storage
**Why:**
- Direct HDF5 file access
- No database translation layer
- Maximum compatibility
- Users control file location

**Alternative:** Database Backend
- Loses HDF5-specific features
- Network overhead

### 5. CSS-in-CSS (not CSS-in-JS)
**Why:**
- Faster styling performance
- Simpler deployment
- Works with static builds
- Standard tooling

## Performance Optimizations

### Backend
1. **Dataset Truncation**
   - Limit memory for large datasets
   - Show preview instead of full data
   - Hyperslab selection for querying

2. **Streaming (Future)**
   - Server-Sent Events for progress
   - Chunked response for large files

3. **Caching (Future)**
   - Cache file structures
   - Metadata indexing
   - LRU cache for datasets

### Frontend
1. **Virtual Scrolling (Future)**
   - Only render visible cells
   - Infinite scroll for large arrays

2. **Memoization**
   - React.memo for TreeNode
   - useCallback for handlers

3. **Code Splitting (Future)**
   - Lazy load heavy components
   - Dynamic imports

## Security Considerations

### Input Validation
- File path sanitization (prevent traversal)
- Type checking on data updates
- Size limits on file operations

### API Security
- CORS configured (restrict in production)
- No authentication (add for production)
- No SQL injection (no SQL used)

### Data Integrity
- Read-only mode default
- Confirmation for destructive operations
- Error handling prevents crashes

## Deployment Models

### 1. Local Development
```
npm run dev (Frontend on :3000)
go run main.go (Backend on :8080)
```

### 2. Docker Compose
```
docker-compose up
Both services orchestrated
```

### 3. Production Docker
```
docker build -t hdf5-agent .
docker run -p 8080:8080 -v data:/root/data hdf5-agent
```

### 4. Kubernetes (Future)
- Containerized backend service
- StaticFiles for frontend
- Persistent volume for data

## Testing Strategy

### Unit Tests (Backend)
- Handler functions
- Data serialization
- Type conversions

### Integration Tests
- File operations
- API endpoints
- HDF5 reading/writing

### E2E Tests (Frontend)
- User workflows
- Data modification
- Error handling

## Future Enhancements

### Immediate
- Authentication/Authorization
- Data export (JSON, CSV)
- Dataset statistics
- Attribute editing

### Medium-term
- Advanced visualization (3D plots)
- Batch operations
- Real-time collaboration
- Full-text search

### Long-term
- Version control integration
- Custom plugin system
- Distributed processing
- ML pipeline integration

## Monitoring & Logging

### Backend
```go
log.Printf("Info: %v", event)
log.Fatal(err)  // Exit on critical error
```

### Frontend
```javascript
console.log(message)  // Development
// Production: Replace with analytics
```

### Health Checks
```
GET /api/health → { status: "healthy" }
```

## Conclusion

HDF5 Agent provides a modern, user-friendly interface for HDF5 file manipulation. The architecture balances simplicity with functionality, using proven technologies (Go, React) to deliver a production-ready application suitable for scientific computing and data analysis workflows.

---

**Version:** 1.0.0  
**Last Updated:** 2024
