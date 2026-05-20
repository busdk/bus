#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "010"
setup_dispatch_fixtures

cat > "$WS/expected_usage.err" <<'EOF_EXPECT_USAGE'
usage: bus <command> [args...]
tip: did you mean `bus shell`?

available commands:
  bus accounts           Run bus-accounts from PATH
  bus env                Run bus-env from PATH
  bus fail               Run bus-fail from PATH
  bus journal            Run bus-journal from PATH
  bus ledger             Run bus-ledger from PATH
  bus status             Run bus-status from PATH
EOF_EXPECT_USAGE

set +e
PATH="$TEST_PATH" "$BIN" > "$WS/noargs.out" 2> "$WS/noargs.err"
noargs_code=$?
set -e

test "$noargs_code" -eq 2
! test -s "$WS/noargs.out"
diff -u "$WS/expected_usage.err" "$WS/noargs.err"
grep -q '^available commands:$' "$WS/noargs.err"

cat > "$WS/expected_missing.err" <<'EOF_EXPECT_MISSING'
bus: missing subcommand: missing; expected executable named bus-missing in PATH
usage: bus <command> [args...]
tip: did you mean `bus shell`?

available commands:
  bus accounts           Run bus-accounts from PATH
  bus env                Run bus-env from PATH
  bus fail               Run bus-fail from PATH
  bus journal            Run bus-journal from PATH
  bus ledger             Run bus-ledger from PATH
  bus status             Run bus-status from PATH
EOF_EXPECT_MISSING

set +e
PATH="$TEST_PATH" "$BIN" missing > "$WS/missing.out" 2> "$WS/missing.err"
missing_code=$?
set -e

test "$missing_code" -eq 127
! test -s "$WS/missing.out"
diff -u "$WS/expected_missing.err" "$WS/missing.err"
grep -q '^bus: missing subcommand:' "$WS/missing.err"

PATH="$TEST_PATH" "$BIN" --help > "$WS/help_global.out" 2> "$WS/help_global.err"
grep -q '^Usage:$' "$WS/help_global.out"
grep -q '^  bus \[global flags\] <command> \[args...\]$' "$WS/help_global.out"
grep -q '^  -C, --chdir <dir>    Change working directory before dispatch$' "$WS/help_global.out"
grep -q '^Available commands:$' "$WS/help_global.out"
grep -q '^  bus accounts \+Run bus-accounts from PATH$' "$WS/help_global.out"
! test -s "$WS/help_global.err"

PATH="$TEST_PATH" "$BIN" --version > "$WS/version_global.out" 2> "$WS/version_global.err"
diff -u <(printf 'bus dev\n') "$WS/version_global.out"
! test -s "$WS/version_global.err"

PATH="$TEST_PATH" "$BIN" -C / status --version > "$WS/global_chdir.out" 2> "$WS/global_chdir.err"
diff -u <(printf 'bus-status e2e\n') "$WS/global_chdir.out"
! test -s "$WS/global_chdir.err"

PATH="$TEST_PATH" "$BIN" -q status --version > "$WS/global_quiet.out" 2> "$WS/global_quiet.err"
diff -u <(printf 'bus-status e2e\n') "$WS/global_quiet.out"
! test -s "$WS/global_quiet.err"

PATH="$TEST_PATH" "$BIN" -- status --version > "$WS/global_dd.out" 2> "$WS/global_dd.err"
diff -u <(printf 'bus-status e2e\n') "$WS/global_dd.out"
! test -s "$WS/global_dd.err"

set +e
PATH="$TEST_PATH" "$BIN" help > "$WS/help.out" 2> "$WS/help.err"
help_code=$?
set -e

test "$help_code" -eq 0
grep -q '^bus exposes live dispatcher metadata\.$' "$WS/help.out"
grep -q '^  bus help \[--format text|opencli|json\]$' "$WS/help.out"
! test -s "$WS/help.err"

echo "e2e OK"
