# Creating the GitHub Repository

Follow these steps to create and push the HDF5 Agent project to GitHub:

## Step 1: Create Repository on GitHub

1. Go to https://github.com/new
2. Fill in details:
   - **Repository name:** `hdf5-agent`
   - **Description:** "Full-stack Go + React application for browsing and manipulating HDF5 files"
   - **Visibility:** Public
   - **Initialize with:** None (we'll push existing files)
3. Click **Create repository**

## Step 2: Initialize Git Locally

```bash
cd /Users/tacurran/project/hdf5-agent

# Initialize git if not already done
git init

# Add all files
git add .

# Create initial commit
git commit -m "Initial commit: HDF5 Agent full-stack application

- Go backend with gonum/hdf5
- React frontend with Vite
- Docker support
- File browser and data viewer"
```

## Step 3: Add Remote and Push

```bash
# Replace YOUR-USERNAME with your GitHub username
git remote add origin https://github.com/YOUR-USERNAME/hdf5-agent.git

# Push to main branch
git branch -M main
git push -u origin main
```

## Step 4: Verify Repository

1. Go to https://github.com/YOUR-USERNAME/hdf5-agent
2. Verify all files are present
3. Check README displays correctly

## Step 5: Configure GitHub Settings

### Repository Settings
1. Go to Settings → General
2. Set:
   - Default branch: `main`
   - Allow squash merging: ✓
   - Allow rebase merging: ✓
   - Delete branch on merge: ✓

### Branch Protection (Optional)
1. Go to Settings → Branches
2. Click "Add rule"
3. Branch name pattern: `main`
4. Enable:
   - Require pull request reviews before merging: 1
   - Require status checks to pass before merging
   - Require branches to be up to date

### Add Topics
1. Go to Repository details (top right)
2. Add topics:
   - `hdf5`
   - `data-visualization`
   - `go`
   - `react`
   - `docker`
   - `scientific-computing`

## Step 6: Create Releases

### Create Release v1.0.0

```bash
# Tag the version
git tag -a v1.0.0 -m "HDF5 Agent v1.0.0 - Initial Release"

# Push tag
git push origin v1.0.0
```

Then on GitHub:
1. Go to Releases
2. Click "Create a release"
3. Select tag `v1.0.0`
4. Add release notes:

```markdown
# HDF5 Agent v1.0.0

## Features
- Browse HDF5 files in web interface
- View hierarchical file structure
- Display dataset contents
- Support for multiple data types
- Docker deployment

## Installation
```bash
docker-compose up
```

## Requirements
- Docker & Docker Compose

## Known Limitations
- Datasets > 100k points show preview only
- Some advanced HDF5 features not supported

## Bug Fixes & Improvements
- Fixed panel layout responsiveness
- Improved error handling
- Added comprehensive logging

For complete changelog, see CHANGELOG.md
```

## Step 7: Optional - Add GitHub Pages

Create a simple landing page in `docs/index.html`:

```html
<!DOCTYPE html>
<html>
<head>
    <title>HDF5 Agent</title>
    <style>
        body { font-family: Arial; margin: 40px; }
        h1 { color: #333; }
        code { background: #f4f4f4; padding: 2px 6px; }
    </style>
</head>
<body>
    <h1>HDF5 Agent</h1>
    <p>Full-stack Go + React application for browsing HDF5 files</p>
    <h2>Quick Start</h2>
    <code>docker-compose up</code>
    <p><a href="https://github.com/YOUR-USERNAME/hdf5-agent">View on GitHub</a></p>
</body>
</html>
```

Then enable Pages:
1. Settings → Pages
2. Source: Deploy from a branch
3. Branch: `main`
4. Folder: `/docs`

## Step 8: Create Development Workflow

### Create develop branch
```bash
git checkout -b develop
git push -u origin develop
```

### Git Workflow
- `main` - Production releases only
- `develop` - Integration branch for features
- `feature/*` - Feature branches off develop
- `bugfix/*` - Bug fix branches off develop

### Example workflow:
```bash
# Create feature branch
git checkout -b feature/dataset-editing develop

# Make changes
git add .
git commit -m "Add dataset editing feature"

# Push and create PR
git push origin feature/dataset-editing
# Open PR on GitHub to merge into develop

# After merging, create release PR:
git checkout -b release/v1.1.0 main
# Update version numbers
git commit -m "Release v1.1.0"
git push origin release/v1.1.0
# Merge to main and tag
```

## Step 9: Set Up CI/CD

The `.github/workflows/ci.yml` will automatically:
- Run on every push to `main` and `develop`
- Run on pull requests
- Build backend and frontend
- Build Docker images
- Run tests (when added)

View results in Actions tab.

## Step 10: Add Collaborators (Optional)

1. Settings → Collaborators
2. Click "Add people"
3. Enter GitHub usernames
4. Set permission level (Maintain/Write/Triage)

## Useful Commands

### Update local repo
```bash
git fetch origin
git pull origin main
```

### Create release branch
```bash
git checkout -b release/v1.1.0 main
git push -u origin release/v1.1.0
```

### View tags
```bash
git tag
git show v1.0.0
```

### Create signed tag (optional)
```bash
git tag -s v1.0.0 -m "Version 1.0.0"
```

## GitHub Actions Secrets (if needed)

1. Settings → Secrets and variables → Actions
2. Create secrets for:
   - `DOCKER_USERNAME` (if pushing to Docker Hub)
   - `DOCKER_PASSWORD`
   - `DEPLOY_KEY` (if auto-deploying)

## Maintenance

### Weekly
- Review open issues and PRs
- Run tests locally
- Update dependencies

### Monthly
- Review metrics in Insights
- Plan next release
- Update documentation

### Quarterly
- Major feature planning
- Roadmap update
- Community feedback review

---

**Your GitHub repository is ready!** 🎉

Share the link: `https://github.com/YOUR-USERNAME/hdf5-agent`
