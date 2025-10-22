#!/bin/bash
set -e

echo "🔄 Generating Bridge bindings from IDL..."

anchor-go \
  --idl bridge-minimal.json \
  --output ./cmd/internal/bridge/bindings/ \
  --name bindings \
  --no-go-mod

echo ""
echo "✅ Bindings generated successfully!"
echo "   📁 Output: cmd/internal/bridge/bindings/"
echo "   📦 Package: bindings"
echo "   ✓ No go.mod created (using root go.mod)"
echo ""
echo "🔨 Building to verify..."
go build ./...
echo "✅ Build successful!"

