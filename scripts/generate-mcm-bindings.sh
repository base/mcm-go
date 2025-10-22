#!/bin/bash
set -e

echo "🔄 Generating MCM bindings from IDL..."

anchor-go \
  --idl mcm.json \
  --output ./pkg/bindings/ \
  --name bindings \
  --no-go-mod

echo ""
echo "✅ Bindings generated successfully!"
echo "   📁 Output: pkg/bindings/"
echo "   📦 Package: bindings"
echo "   ✓ No go.mod created (using root go.mod)"
echo ""
echo "🔨 Building to verify..."
go build ./...
echo "✅ Build successful!"
