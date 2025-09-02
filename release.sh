#!/bin/bash

# Release script for azctl
# Usage: ./release.sh v0.2.0

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if version argument is provided
if [ $# -eq 0 ]; then
    log_error "Version argument is required"
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.2.0"
    exit 1
fi

VERSION="$1"

# Validate version format (should start with 'v' followed by semantic version)
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    log_error "Invalid version format. Expected format: vX.Y.Z"
    echo "Example: v0.2.0"
    exit 1
fi

log_info "Starting release process for version: $VERSION"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    log_error "Not in a git repository"
    exit 1
fi

# Check if we have uncommitted changes
if ! git diff-index --quiet HEAD --; then
    log_warning "You have uncommitted changes. Please commit or stash them before releasing."
    echo "Uncommitted files:"
    git status --porcelain
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Release cancelled"
        exit 1
    fi
fi

# Check if tag exists locally
if git tag -l "$VERSION" | grep -q "$VERSION"; then
    log_warning "Tag $VERSION exists locally. Removing..."
    git tag -d "$VERSION"
    log_success "Local tag $VERSION removed"
else
    log_info "Tag $VERSION does not exist locally"
fi

# Check if tag exists remotely
if git ls-remote --tags origin | grep -q "refs/tags/$VERSION"; then
    log_warning "Tag $VERSION exists remotely. Removing..."
    git push origin ":refs/tags/$VERSION"
    log_success "Remote tag $VERSION removed"
else
    log_info "Tag $VERSION does not exist remotely"
fi

# Create new tag
log_info "Creating new tag: $VERSION"
git tag "$VERSION"

# Push the new tag
log_info "Pushing tag to remote..."
git push origin "$VERSION"

log_success "Release tag $VERSION created and pushed successfully!"

# Show the new tag
log_info "New tag details:"
git show "$VERSION" --no-patch

echo ""
log_info "Next steps:"
echo "1. Create a GitHub release at: https://github.com/furiatona/azctl/releases/new"
echo "2. Upload the built binaries for this release"
echo "3. Update the download links in documentation"
