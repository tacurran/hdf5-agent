# Changelog

All notable changes to the HDF5 Agent project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Dataset editing feature (in development)
- Batch operations on multiple files (planned)
- Export to CSV/JSON (planned)
- Advanced search and filtering (planned)

### Changed
- Improved error handling in HDF5 operations
- Better logging for debugging

### Fixed
- TBD

### Deprecated
- TBD

### Removed
- TBD

### Security
- TBD

## [1.0.0] - 2026-04-30

### Added
- Initial release of HDF5 Agent
- Go backend with RESTful API
- React frontend with interactive UI
- File browser for HDF5 files
- Hierarchical structure explorer
- Dataset viewer for multiple data types
- Docker and Docker Compose support
- CORS middleware for cross-origin requests
- Panic recovery for HDF5 library errors
- Support for int64, float64, and string datatypes
- Comprehensive error handling

### Features
#### Backend (Go)
- `GET /api/health` - Health check endpoint
- `GET /api/files` - List HDF5 files in directory
- `GET /api/file/structure` - Get file hierarchy
- `GET /api/dataset/read` - Read dataset contents
- Gorilla Mux for routing
- CORS support with rs/cors
- gonum/v1/hdf5 bindings

#### Frontend (React)
- Three-pane layout (files, structure, data)
- Expandable tree view for groups
- Dataset preview with data grid
- Dark theme with cyan accents
- Responsive design
- Real-time API integration

#### Deployment
- Multi-stage Docker builds
- Docker Compose orchestration
- Environment-based configuration
- Volume mounting for file access

### Documentation
- Comprehensive README with setup instructions
- API documentation
- Contributing guidelines
- Issue templates (bug, feature)
- Pull request template

### Testing
- Manual testing with sample HDF5 files
- Docker build verification
- API endpoint testing
- Cross-browser testing

### Known Limitations
- Datasets > 100k points show warning
- Some advanced HDF5 features not fully supported
- Variable-length strings have limited support
- Compound datatypes show as "unknown"
- Does not support editing/writing HDF5 files

### Dependencies

#### Go
- github.com/gorilla/mux v1.8.1
- github.com/rs/cors v1.11.1
- gonum.org/v1/hdf5 v0.0.0-20210714002203

#### Node.js
- react ^18.0.0
- react-dom ^18.0.0
- vite ^4.0.0

### Infrastructure
- GitHub repository with CI/CD
- GitHub Actions workflows
- Release management with semantic versioning

---

## Version History

### How to Update

#### Minor updates (v1.0.0 → v1.1.0)
- New features
- Bug fixes
- Performance improvements

```bash
git checkout -b release/v1.1.0 main
# Update version in package.json and go files
git commit -m "Bump version to 1.1.0"
git push origin release/v1.1.0
# Create PR, merge to main
git tag v1.1.0
git push origin v1.1.0
```

#### Patch updates (v1.0.0 → v1.0.1)
- Bug fixes only
- No new features

#### Major updates (v1.0.0 → v2.0.0)
- Breaking changes
- Major refactoring
- New architecture

---

## Support

For more information:
- GitHub: https://github.com/YOUR-USERNAME/hdf5-agent
- Issues: https://github.com/YOUR-USERNAME/hdf5-agent/issues
- Discussions: https://github.com/YOUR-USERNAME/hdf5-agent/discussions
