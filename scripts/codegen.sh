#!/bin/bash
set -e

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT"

echo "==> Generating Kitex code from IDL..."
echo "    IDL path: $ROOT/idl/"
echo ""
echo "    Skipped: processor.thrift not yet defined (see P5 plan)"
echo ""
echo "    To regenerate after adding IDL:"
echo "      cd idl && kitex -module github.com/angelapytao/diffgram-go processor.thrift"
echo ""
echo "Codegen complete."