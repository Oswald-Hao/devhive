#!/usr/bin/env bash
# DevHive Release Script
# Bumps version across all files, builds binaries, creates archives, and publishes to npm.
#
# USAGE:
#   VERSION=0.3.0 bash scripts/release.sh          # Full release
#   VERSION=0.3.0 bash scripts/release.sh --dry-run  # Build only, skip npm publish

set -euo pipefail

# Ensure Go is available
export PATH="$HOME/go/bin:$HOME/.go/bin:/usr/local/go/bin:$PATH"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION="${VERSION:-}"

if [ -z "$VERSION" ]; then
    echo "Usage: VERSION=<version> bash scripts/release.sh [--dry-run]"
    echo "Example: VERSION=0.3.0 bash scripts/release.sh"
    exit 1
fi

DRY_RUN=false
if [ "${1:-}" = "--dry-run" ]; then
    DRY_RUN=true
fi

# Strip leading 'v' if present
VERSION="${VERSION#v}"

echo "=== DevHive Release v${VERSION} ==="
echo ""

# ── 1. Bump version in all files ────────────────────────────────

echo "[1/5] Bumping version to ${VERSION}..."

# VERSION file
echo "${VERSION}" > "$ROOT/VERSION"

# cmd/dh/main.go: const version = "X.Y.Z"
sed -i '' "s/const version = \".*\"/const version = \"${VERSION}\"/" "$ROOT/cmd/dh/main.go" 2>/dev/null || \
    sed -i "s/const version = \".*\"/const version = \"${VERSION}\"/" "$ROOT/cmd/dh/main.go"

# package.json
sed -i '' "s/\"version\": \".*\"/\"version\": \"${VERSION}\"/" "$ROOT/package.json" 2>/dev/null || \
    sed -i "s/\"version\": \".*\"/\"version\": \"${VERSION}\"/" "$ROOT/package.json"

# install.sh
sed -i '' "s/VERSION=\"\${DEVHIVE_VERSION:-.*}\"/VERSION=\"\${DEVHIVE_VERSION:-${VERSION}}\"/" "$ROOT/install.sh" 2>/dev/null || \
    sed -i "s/VERSION=\"\${DEVHIVE_VERSION:-.*}\"/VERSION=\"\${DEVHIVE_VERSION:-${VERSION}}\"/" "$ROOT/install.sh"

echo "  ✓ Version bumped in VERSION, cmd/dh/main.go, package.json, install.sh"

# ── 2. Build Go binaries ────────────────────────────────────────

echo "[2/5] Building Go binaries..."

BUILD_DIR="$ROOT/build/$VERSION"
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for platform in "${platforms[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"
    output="$BUILD_DIR/dh-${GOOS}-${GOARCH}"
    echo "  → ${GOOS}/${GOARCH}"
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
        go build -ldflags="-s -w" -o "$output" ./cmd/dh/
done

echo "  ✓ Binaries built"

# ── 3. Create tar.gz archives ───────────────────────────────────

echo "[3/5] Creating archives..."

ARCHIVE_DIR="$ROOT/build/archives"
rm -rf "$ARCHIVE_DIR"
mkdir -p "$ARCHIVE_DIR"

for platform in "${platforms[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"
    binary="dh-${GOOS}-${GOARCH}"
    tarball="devhive-${VERSION}-${GOOS}-${GOARCH}.tar.gz"

    cp "$BUILD_DIR/$binary" "$BUILD_DIR/dh"
    tar -czf "$ARCHIVE_DIR/$tarball" -C "$BUILD_DIR" dh
    rm "$BUILD_DIR/dh"
    echo "  ✓ $tarball"
done

echo "  ✓ Archives created in build/archives/"

# ── 4. Prepare npm package ──────────────────────────────────────

echo "[4/5] Preparing npm package..."

NPM_DIR="$ROOT/build/npm"
rm -rf "$NPM_DIR"
mkdir -p "$NPM_DIR/bin"

# Copy npm package files
cp "$ROOT/package.json" "$NPM_DIR/"
cp "$ROOT/install.sh" "$NPM_DIR/"
cp "$ROOT/config.example.yaml" "$NPM_DIR/"

# Create the bin/dh wrapper (same as current but with updated find logic)
cp "$ROOT/bin/dh" "$NPM_DIR/bin/dh"

echo "  ✓ npm package prepared in build/npm/"

# ── 5. Publish ──────────────────────────────────────────────────

if [ "$DRY_RUN" = true ]; then
    echo ""
    echo "=== Dry run complete ==="
    echo "Binaries:   build/${VERSION}/"
    echo "Archives:   build/archives/"
    echo "npm pkg:    build/npm/"
    echo ""
    echo "To publish:"
    echo "  1. Upload archives to GitHub Releases (v${VERSION})"
    echo "  2. cd build/npm && npm publish"
    exit 0
fi

echo "[5/5] Publishing..."

# Publish to npm
cd "$NPM_DIR"
if command -v npm &>/dev/null; then
    npm publish --access public
    echo "  ✓ Published to npm"
else
    echo "  ✗ npm not found, skipping npm publish"
fi

echo ""
echo "=== DevHive v${VERSION} released ==="
echo ""
echo "Next steps:"
echo "  - Upload archives from build/archives/ to GitHub Releases"
echo "  - Push version bump commit:"
echo "    git add VERSION cmd/dh/main.go package.json install.sh"
echo "    git commit -m 'release: v${VERSION}'"
echo "    git tag v${VERSION}"
echo "    git push && git push --tags"
