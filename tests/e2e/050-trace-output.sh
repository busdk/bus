#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "050"
setup_dispatch_fixtures

cat > "$WS/2024-01.bus" <<'EOF_202401'
accounts jan
EOF_202401

PATH="$TEST_PATH" "$BIN" --trace "$WS/2024-01.bus" > "$WS/trace.out" 2> "$WS/trace.err"
grep -q '2024-01.bus:1: bus accounts jan' "$WS/trace.out"
grep -q 'ACCOUNTS:jan' "$WS/trace.out"
! test -s "$WS/trace.err"

echo "e2e OK"
