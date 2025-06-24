#!/bin/bash
set -e

echo "🔧 Simulating CI environment..."

# Step 1: Clean slate
echo "1. Cleaning module cache..."
go clean -modcache

# Step 2: Reset to default Go proxy settings (like fresh CI)
echo "2. Setting default Go proxy settings..."
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org
export GO111MODULE=on

# Step 3: Show environment
echo "3. Go environment:"
echo "   GOPROXY: $GOPROXY"
echo "   GOSUMDB: $GOSUMDB"
echo "   GO111MODULE: $GO111MODULE"

# Step 4: Try to build (this would have failed with pkg/myshift)
echo "4. Attempting build..."
if go build ./...; then
    echo "✅ Build successful!"
else
    echo "❌ Build failed - this is what CI experienced!"
    exit 1
fi

echo "🎉 CI simulation complete!"
