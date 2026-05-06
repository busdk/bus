#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "015"
setup_dispatch_fixtures

PATH="$TEST_PATH" "$BIN" --perf accounts inspect > "$WS/perf.out" 2> "$WS/perf.err"
diff -u <(printf 'ACCOUNTS:inspect\n') "$WS/perf.out"
grep -q '^INFO perf bus-accounts inspect ' "$WS/perf.err"

PATH="$TEST_PATH" "$BIN" -v accounts inspect > "$WS/debug.out" 2> "$WS/debug.err"
diff -u <(printf 'ACCOUNTS:-v inspect\n') "$WS/debug.out"
grep -q '^DEBUG perf bus-accounts inspect ' "$WS/debug.err"

PATH="$TEST_PATH" "$BIN" -vv accounts inspect > "$WS/repeated-trace.out" 2> "$WS/repeated-trace.err"
diff -u <(printf 'ACCOUNTS:-v -v inspect\n') "$WS/repeated-trace.out"
grep -q '^TRACE perf bus-accounts inspect ' "$WS/repeated-trace.err"

PATH="$TEST_PATH" "$BIN" --trace accounts inspect > "$WS/trace.out" 2> "$WS/trace.err"
diff -u <(printf 'ACCOUNTS:--trace inspect\n') "$WS/trace.out"
grep -q '^TRACE perf bus-accounts inspect ' "$WS/trace.err"

PATH="$TEST_PATH" "$BIN" --quiet --perf accounts inspect > "$WS/quiet.out" 2> "$WS/quiet.err"
diff -u <(printf 'ACCOUNTS:--quiet inspect\n') "$WS/quiet.out"
! test -s "$WS/quiet.err"

set +e
PATH="$TEST_PATH" "$BIN" --quiet --trace accounts inspect > "$WS/conflict.out" 2> "$WS/conflict.err"
conflict_code=$?
set -e

test "$conflict_code" -eq 2
! test -s "$WS/conflict.out"
grep -q '^bus: invalid usage: --quiet and --verbose/--trace are mutually exclusive$' "$WS/conflict.err"
grep -q '^usage: bus <command> \[args...\]$' "$WS/conflict.err"

echo "e2e OK"
