#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "110"
setup_dispatch_fixtures

cat > "$WS/path_first/bus-update" <<'SH'
#!/bin/sh
printf 'UPDATE:%s\n' "$*"
if [ "${BUS_SUBCMD_EXIT_CODE:-0}" -ne 0 ]; then
  exit "${BUS_SUBCMD_EXIT_CODE}"
fi
SH
chmod +x "$WS/path_first/bus-update"

PATH="$TEST_PATH" "$BIN" update package install --module bus-ledger > "$WS/update.out" 2> "$WS/update.err"
diff -u <(printf 'UPDATE:package install --module bus-ledger\n') "$WS/update.out"
! test -s "$WS/update.err"

set +e
PATH="$TEST_PATH" BUS_SUBCMD_EXIT_CODE=9 "$BIN" update package verify > "$WS/verify.out" 2> "$WS/verify.err"
status=$?
set -e
test "$status" -eq 9
diff -u <(printf 'UPDATE:package verify\n') "$WS/verify.out"
! test -s "$WS/verify.err"

rm -f "$WS/path_first/bus-update"
set +e
PATH="$TEST_PATH" "$BIN" update package install --module bus-ledger > "$WS/missing.out" 2> "$WS/missing.err"
status=$?
set -e
test "$status" -eq 127
! test -s "$WS/missing.out"
grep -q 'expected executable named bus-update in PATH' "$WS/missing.err"

echo "e2e OK"
