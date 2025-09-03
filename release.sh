#!/bin/bash

# Release script for azctl
# Usage:
#   ./release.sh <version> [major_tag] [--with-release]
# Examples:
#   ./release.sh v1.1.0
#   ./release.sh v1.1.0 v1
#   ./release.sh v1.1.0 v1 --with-release

set -euo pipefail

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Log helpers
log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }

show_help() {
  echo "Usage: $0 <version> [major_tag] [--with-release]"
  echo
  echo "Arguments:"
  echo "  version       Semantic version (e.g. v1.1.0)"
  echo "  major_tag     Optional major tag alias (e.g. v1)"
  echo
  echo "Options:"
  echo "  --with-release  Also delete GitHub release objects before retagging (requires gh CLI)"
  echo "  --help          Show this help message"
  echo
  echo "Examples:"
  echo "  $0 v1.1.0"
  echo "  $0 v1.1.0 v1"
  echo "  $0 v1.1.0 v1 --with-release"
}

# Parse args
WITH_RELEASE=false
VERSION=""
MAJOR_TAG=""

for arg in "$@"; do
  case "$arg" in
    --with-release) WITH_RELEASE=true ;;
    --help) show_help; exit 0 ;;
    v[0-9]*.[0-9]*.[0-9]*) VERSION="$arg" ;;
    v[0-9]*) MAJOR_TAG="$arg" ;;
    *) log_error "Unknown argument: $arg"; show_help; exit 1 ;;
  esac
done

if [ -z "$VERSION" ]; then
  log_error "Version argument is required"
  show_help
  exit 1
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  log_error "Invalid version format. Expected: vX.Y.Z"
  exit 1
fi

log_info "Starting release process for version: $VERSION (major tag: ${MAJOR_TAG:-none})"

# Ensure git repo
if ! git rev-parse --git-dir > /dev/null 2>&1; then
  log_error "Not in a git repository"
  exit 1
fi

# Warn if dirty
if ! git diff-index --quiet HEAD --; then
  log_warning "Uncommitted changes found"
  read -p "Continue anyway? (y/N): " -n 1 -r
  echo
  [[ ! $REPLY =~ ^[Yy]$ ]] && exit 1
fi

# Function: delete tag locally & remotely
delete_tag() {
  local TAG=$1
  if git tag -l "$TAG" | grep -q "$TAG"; then
    log_warning "Tag $TAG exists locally. Removing..."
    git tag -d "$TAG"
  fi
  if git ls-remote --tags origin | grep -q "refs/tags/$TAG"; then
    log_warning "Tag $TAG exists remotely. Removing..."
    git push origin ":refs/tags/$TAG"
  fi
}

# Function: delete GitHub release
delete_release() {
  local TAG=$1
  if $WITH_RELEASE && command -v gh >/dev/null 2>&1; then
    log_info "Deleting GitHub release for $TAG..."
    gh release delete "$TAG" -y || true
  fi
}

# Clean up existing tags/releases
delete_tag "$VERSION"
delete_release "$VERSION"
[ -n "$MAJOR_TAG" ] && delete_tag "$MAJOR_TAG" && delete_release "$MAJOR_TAG"

# Create & push tags
log_info "Creating new tag: $VERSION"
git tag "$VERSION"
git push origin "$VERSION"

if [ -n "$MAJOR_TAG" ]; then
  log_info "Creating/Updating major tag: $MAJOR_TAG"
  git tag -f "$MAJOR_TAG"
  git push origin "$MAJOR_TAG" --force
fi

log_success "Release process complete for $VERSION"