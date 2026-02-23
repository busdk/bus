#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "020"
setup_dispatch_fixtures

PATH="$TEST_PATH" "$BIN" accounts alpha beta > "$WS/dispatch.out" 2> "$WS/dispatch.err"
diff -u <(printf 'ACCOUNTS:alpha beta\n') "$WS/dispatch.out"
! test -s "$WS/dispatch.err"

echo "e2e OK"
