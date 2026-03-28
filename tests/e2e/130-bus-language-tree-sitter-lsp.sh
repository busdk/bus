#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "130"

make -C "$ROOT_DIR" check-tree-sitter-bus-language >/dev/null
make -C "$ROOT_DIR" check-bus-language-server >/dev/null

echo "e2e OK"
