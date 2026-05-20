#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "090"
setup_dispatch_fixtures

cat > "$WS/path_first/bus-validate" <<'SH'
#!/bin/sh
printf 'VALIDATE:%s\n' "$*"
SH
chmod +x "$WS/path_first/bus-validate"

PATH="$TEST_PATH" "$BIN" audit evidence-coverage --year 2026 > "$WS/audit_alias.out" 2> "$WS/audit_alias.err"
diff -u <(printf 'VALIDATE:evidence-coverage --year 2026\n') "$WS/audit_alias.out"
! test -s "$WS/audit_alias.err"

PATH="$TEST_PATH" "$BIN" audit evidence-coverage -h > "$WS/audit_alias_help_short.out" 2> "$WS/audit_alias_help_short.err"
diff -u <(printf 'VALIDATE:--help evidence-coverage\n') "$WS/audit_alias_help_short.out"
! test -s "$WS/audit_alias_help_short.err"

PATH="$TEST_PATH" "$BIN" audit evidence-coverage --help > "$WS/audit_alias_help_long.out" 2> "$WS/audit_alias_help_long.err"
diff -u <(printf 'VALIDATE:--help evidence-coverage\n') "$WS/audit_alias_help_long.out"
! test -s "$WS/audit_alias_help_long.err"

PATH="$TEST_PATH" "$BIN" --help > "$WS/help.out" 2> "$WS/help.err"
! test -s "$WS/help.err"
grep -q '^  bus audit \+Dispatch audit evidence-coverage workflows$' "$WS/help.out"

audit_code=0
PATH="$TEST_PATH" "$BIN" audit > "$WS/audit_missing.out" 2> "$WS/audit_missing.err" || audit_code=$?
test "$audit_code" -eq 2
grep -q 'audit requires subcommand evidence-coverage' "$WS/audit_missing.err"

echo "e2e OK"
