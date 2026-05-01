# Contributing to HDF5 Agent

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## Getting Started

### 1. Fork the Repository
```bash
git clone https://github.com/YOUR-USERNAME/hdf5-agent.git
cd hdf5-agent
```

### 2. Create a Feature Branch
```bash
git checkout -b feature/your-feature-name
```

### 3. Set Up Development Environment

**Backend:**
```bash
export HDF5_DIR=$(brew --prefix hdf5)
export CGO_CPPFLAGS="-I$(brew --prefix hdf5)/include"
export CGO_LDFLAGS="-L$(brew --prefix hdf5)/lib"
go run main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## Development Workflow

### Backend Changes (Go)

1. Edit `main.go`
2. Test locally: `go run main.go`
3. Build and test: `go build -o hdf5-agent`
4. Run tests: `go test ./...`

Key files:
- `main.go` - Server and handlers
- `go.mod` - Dependencies

### Frontend Changes (React)

1. Edit files in `frontend/src/`
2. Changes auto-reload in dev mode
3. Build: `npm run build`
4. Test: Manual testing in browser

Key files:
- `frontend/src/App.jsx` - Main component
- `frontend/src/App.css` - Styling
- `frontend/package.json` - Dependencies

### Docker Changes

1. Edit `Dockerfile` or `docker-compose.yml`
2. Rebuild: `docker-compose down && docker-compose up --build`
3. Test: `http://localhost:3000`

## Testing

### Manual Testing

1. Start services (local or Docker)
2. Create test HDF5 file:
```bash
python3 << 'EOF'
import h5py
import numpy as np

with h5py.File('test_data.h5', 'w') as f:
    grp = f.create_group('measurements')
    grp.create_dataset('waveform', data=np.sin(np.linspace(0, 2*np.pi, 1000)))
    grp.create_dataset('matrix', data=np.arange(200).reshape(10, 20))
EOF
```
3. Test in browser at `http://localhost:3000`
4. Test API endpoints:
```bash
curl http://localhost:8080/api/files
curl "http://localhost:8080/api/file/structure?path=test_data.h5"
```

### Testing HDF5 Compatibility

Test with different HDF5 file types:
- Simple datasets
- Nested groups
- Different datatypes (int, float, string)
- Large datasets
- Special datatypes (compound, variable-length)

## Code Style

### Go
- Follow standard Go conventions
- Run `gofmt` before committing
- Use meaningful variable names
- Add comments for exported functions

### JavaScript/React
- Use consistent indentation (2 spaces)
- Use descriptive component names
- Add PropTypes or TypeScript where helpful
- Keep components small and focused

## Commit Messages

Write clear, concise commit messages:

```
# Good
git commit -m "Add dataset preview feature for large files"
git commit -m "Fix: handle empty groups in file structure"
git commit -m "Refactor: simplify datatype detection"

# Avoid
git commit -m "fix stuff"
git commit -m "updates"
git commit -m "wip"
```

## Pull Request Process

1. **Update your fork:**
```bash
git fetch upstream
git rebase upstream/main
```

2. **Push to your fork:**
```bash
git push origin feature/your-feature-name
```

3. **Create Pull Request:**
   - Go to GitHub
   - Click "New Pull Request"
   - Describe your changes clearly
   - Reference related issues

4. **PR Description Template:**
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement

## Testing
- [ ] Tested locally with Go backend
- [ ] Tested frontend in browser
- [ ] Tested with Docker
- [ ] Tested with multiple HDF5 file types

## Screenshots (if UI changes)
Include before/after if relevant

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-reviewed my code
- [ ] Added comments where needed
- [ ] Updated documentation if needed
- [ ] No new warnings generated
```

5. **Address feedback:**
   - Review comments
   - Make requested changes
   - Push additional commits
   - Request re-review

## Architecture Decisions

### HDF5 Library
- Using `gonum.org/v1/hdf5` for Go bindings
- Handles most common HDF5 file structures
- Known limitations documented in README

### UI Framework
- React for interactivity
- Vite for fast dev experience
- Tailwind for styling (if added)

### API Design
- RESTful endpoints
- JSON responses
- Error handling with HTTP status codes

## Documentation

### Code Comments
```go
// Recover from panics in HDF5 library
defer func() {
    if r := recover(); r != nil {
        log.Printf("Recovered from panic: %v", r)
    }
}()
```

### README Updates
Update README.md if you:
- Add new features
- Change API endpoints
- Modify setup instructions
- Add dependencies

### Docstrings (Go)
```go
// listGroupObjects recursively traverses an HDF5 group
// and returns its contents (subgroups and datasets)
func (s *Server) listGroupObjects(group *hdf5.Group) ([]map[string]interface{}, error) {
    // implementation
}
```

## Issues and Feature Requests

### Reporting Issues
Include:
1. OS and Go/Node versions
2. Steps to reproduce
3. Expected vs actual behavior
4. Error messages and logs
5. Screenshot if UI-related

### Requesting Features
Include:
1. Use case and motivation
2. Proposed solution
3. Alternative solutions considered
4. Examples or wireframes

## Performance Considerations

### Backend
- Datasets > 100k points show warning
- Use pagination for large results
- Stream data where possible

### Frontend
- Virtual scrolling for large lists
- Debounce search queries
- Lazy load components

## Security

- Don't commit secrets or credentials
- Validate user input on backend
- Use CORS carefully in production
- Keep dependencies updated

## Deployment

Before deploying:
1. Test on local machine
2. Test with Docker
3. Run lint/format checks
4. Create release notes
5. Tag version in git

## Questions?

- Open an issue for questions
- Check existing issues for answers
- Discuss in PR comments

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to HDF5 Agent! 🚀
