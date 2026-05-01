# Quick Start Guide

Get HDF5 Agent running in 5 minutes.

## Fastest Way: Docker

```bash
# 1. Clone repository
git clone https://github.com/yourusername/hdf5-agent.git
cd hdf5-agent

# 2. Run with Docker Compose
docker-compose up

# 3. Open browser
# Frontend:  http://localhost:3000
# Backend:   http://localhost:8080
```

Done! Start with the application.

## Alternative: Local Setup

### macOS

```bash
# Install dependencies
brew install go hdf5 node

# Clone
git clone https://github.com/yourusername/hdf5-agent.git
cd hdf5-agent

# Run
make dev

# Open http://localhost:3000
```

### Ubuntu/Debian

```bash
# Install dependencies
sudo apt-get install golang-go libhdf5-dev nodejs npm

# Clone
git clone https://github.com/yourusername/hdf5-agent.git
cd hdf5-agent

# Run
make dev

# Open http://localhost:3000
```

### Windows

```powershell
# Install using Chocolatey
choco install go mingw hdf5 nodejs

# Clone
git clone https://github.com/yourusername/hdf5-agent.git
cd hdf5-agent

# Run (in separate terminals)
go run main.go
# and
cd frontend && npm run dev
```

## First Steps

### 1. Create Test Data

```bash
go run scripts/create_test_data.go
```

This creates `test_data.h5` with sample datasets.

### 2. Open Application

Open http://localhost:3000 in your browser.

### 3. Load File

1. Click on `test_data.h5` in the left sidebar
2. Explore the file structure in the tree view
3. Click on a dataset to view its data

### 4. Edit Data

1. Click on any data cell to edit
2. Type new value and press Enter
3. Changes sync to file immediately

### 5. Create New File

1. Click "New File" button
2. Enter filename (e.g., `mydata.h5`)
3. Click Create

## Common Commands

```bash
# Check prerequisites
make check

# Build everything
make build

# Run development servers
make dev

# Run tests
make test

# Clean build artifacts
make clean

# Format code
make fmt

# Create test data
make run-test-server
```

## Troubleshooting

### Port Already in Use

```bash
# Use different port
PORT=9000 go run main.go
```

### HDF5 Not Found

```bash
# macOS
brew install hdf5

# Ubuntu
sudo apt-get install libhdf5-dev

# CentOS
sudo yum install hdf5-devel
```

### Node Modules Issues

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Docker Issues

```bash
# Rebuild images
docker-compose build

# Check logs
docker-compose logs -f

# Stop and clean
docker-compose down
docker system prune
```

## Project Layout

```
├── main.go              # Backend server
├── go.mod              # Go dependencies
├── Makefile            # Commands
├── Dockerfile          # Container
├── docker-compose.yml  # Compose config
│
├── frontend/           # React app
│   ├── src/
│   │   ├── App.jsx
│   │   ├── App.css
│   │   └── main.jsx
│   ├── package.json
│   └── vite.config.js
│
├── scripts/            # Helper scripts
│   └── create_test_data.go
│
└── data/              # Your HDF5 files
```

## API Quick Reference

```bash
# List files
curl http://localhost:8080/api/files

# Health check
curl http://localhost:8080/api/health

# Create file
curl -X POST http://localhost:8080/api/file/create \
  -H "Content-Type: application/json" \
  -d '{"filename": "test.h5"}'

# Read dataset
curl "http://localhost:8080/api/dataset/read?file=test.h5&path=/data"

# Update dataset
curl -X POST http://localhost:8080/api/dataset/update?file=test.h5 \
  -H "Content-Type: application/json" \
  -d '{"path": "/data", "data": [1,2,3]}'
```

## Next Steps

- Read [README.md](README.md) for full documentation
- Check [ARCHITECTURE.md](ARCHITECTURE.md) for design details
- Join the discussions for questions
- Report bugs on GitHub Issues

## Support

- 📖 Full docs: See README.md
- 🐛 Issues: GitHub Issues
- 💬 Discussions: GitHub Discussions
- 📚 HDF5 Info: https://www.hdfgroup.org/HDF5/

## Tips

1. **Large Files:** Files >1GB may be slow. Use preview mode.
2. **Backups:** Always backup important HDF5 files before editing.
3. **Performance:** Keep fewer than 100 groups for best experience.
4. **Organization:** Use groups to organize related datasets.
5. **Metadata:** Add attributes to datasets for documentation.

---

**Happy Data Manipulating!** 🚀
