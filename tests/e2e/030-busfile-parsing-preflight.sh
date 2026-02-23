#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "030"
setup_dispatch_fixtures

cat > "$WS/2024-01.bus" <<'EOF_202401'
accounts jan
EOF_202401

cat > "$WS/2024-02.bus" <<'EOF_202402'
ledger feb
EOF_202402

cat > "$WS/all.bus" <<'EOF_ALL'
2024-01.bus
2024-02.bus
EOF_ALL

PATH="$TEST_PATH" "$BIN" "$WS/all.bus" > "$WS/busfile.out" 2> "$WS/busfile.err"
diff -u <(printf 'ACCOUNTS:jan\nLEDGER:feb\n') "$WS/busfile.out"
! test -s "$WS/busfile.err"

cat > "$WS/bad.bus" <<'EOF_BAD'
accounts 'unterminated
EOF_BAD

set +e
PATH="$TEST_PATH" "$BIN" "$WS/2024-01.bus" "$WS/bad.bus" > "$WS/preflight.out" 2> "$WS/preflight.err"
preflight_code=$?
set -e

test "$preflight_code" -eq 65
! test -s "$WS/preflight.out"
grep -q '^.*/bad.bus:1: syntax error:' "$WS/preflight.err"

cat > "$WS/unknown_target.bus" <<'EOF_UNKNOWN_TARGET'
bnak add --id 1
EOF_UNKNOWN_TARGET

set +e
PATH="$TEST_PATH" "$BIN" "$WS/2024-01.bus" "$WS/unknown_target.bus" > "$WS/unknown_target.out" 2> "$WS/unknown_target.err"
unknown_target_code=$?
set -e

test "$unknown_target_code" -eq 127
! test -s "$WS/unknown_target.out"
grep -q 'dispatch error: unknown target "bnak"' "$WS/unknown_target.err"

echo "e2e OK"
