#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "080"
setup_dispatch_fixtures

set +e
PATH="$TEST_PATH" "$BIN" fail sample > "$WS/fail.out" 2> "$WS/fail.err"
fail_code=$?
set -e

test "$fail_code" -eq 7
diff -u <(printf 'FAIL:sample\n') "$WS/fail.out"
! test -s "$WS/fail.err"

echo "e2e OK"
