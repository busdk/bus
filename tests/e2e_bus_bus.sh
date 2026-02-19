#!/usr/bin/env bash
set -e
set -x

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_NAME="$(basename "$ROOT_DIR")"
BIN_PATH="${ROOT_DIR}/bin/${REPO_NAME}"
WORK_DIR="${ROOT_DIR}/tests/e2e_bus_bus_workspace"

cleanup() {
  if [ "${BUS_E2E_KEEP:-0}" = "1" ]; then
    return
  fi
  rm -rf "$WORK_DIR"
}
trap cleanup EXIT

rm -rf "$WORK_DIR"
mkdir -p "$WORK_DIR/path_first" "$WORK_DIR/path_second"

test -f "$BIN_PATH"
test -x "$BIN_PATH"

cat > "$WORK_DIR/path_first/bus-accounts" <<'SH'
#!/bin/sh
printf 'ACCOUNTS:%s\n' "$*"
SH
chmod +x "$WORK_DIR/path_first/bus-accounts"

cat > "$WORK_DIR/path_second/bus-ledger" <<'SH'
#!/bin/sh
printf 'LEDGER:%s\n' "$*"
SH
chmod +x "$WORK_DIR/path_second/bus-ledger"

cat > "$WORK_DIR/path_first/bus-fail" <<'SH'
#!/bin/sh
printf 'FAIL:%s\n' "$*"
exit 7
SH
chmod +x "$WORK_DIR/path_first/bus-fail"

# Duplicate command in later PATH entry must not affect dispatch selection.
cat > "$WORK_DIR/path_second/bus-accounts" <<'SH'
#!/bin/sh
printf 'SECOND:%s\n' "$*"
SH
chmod +x "$WORK_DIR/path_second/bus-accounts"

# Non-executable command-like file must be ignored in discovery.
cat > "$WORK_DIR/path_first/bus-nonexec" <<'TXT'
ignored
TXT

TEST_PATH="${WORK_DIR}/path_first:${WORK_DIR}/path_second"

cat > "$WORK_DIR/expected_usage.err" <<'EOF_EXPECT_USAGE'
usage: bus <command> [args...]

available commands:
  accounts
  fail
  ledger
EOF_EXPECT_USAGE

set +e
PATH="$TEST_PATH" "$BIN_PATH" > "$WORK_DIR/noargs.out" 2> "$WORK_DIR/noargs.err"
noargs_code=$?
set -e

test "$noargs_code" -eq 2
! test -s "$WORK_DIR/noargs.out"
diff -u "$WORK_DIR/expected_usage.err" "$WORK_DIR/noargs.err"
grep -q '^available commands:$' "$WORK_DIR/noargs.err"

cat > "$WORK_DIR/expected_missing.err" <<'EOF_EXPECT_MISSING'
bus: missing subcommand: missing; expected executable named bus-missing in PATH
usage: bus <command> [args...]

available commands:
  accounts
  fail
  ledger
EOF_EXPECT_MISSING

set +e
PATH="$TEST_PATH" "$BIN_PATH" missing > "$WORK_DIR/missing.out" 2> "$WORK_DIR/missing.err"
missing_code=$?
set -e

test "$missing_code" -eq 127
! test -s "$WORK_DIR/missing.out"
diff -u "$WORK_DIR/expected_missing.err" "$WORK_DIR/missing.err"
grep -q '^bus: missing subcommand:' "$WORK_DIR/missing.err"

# If bus-help is absent, "bus help" should print usage and discovered commands.
set +e
PATH="$TEST_PATH" "$BIN_PATH" help > "$WORK_DIR/help.out" 2> "$WORK_DIR/help.err"
help_code=$?
set -e

test "$help_code" -eq 2
! test -s "$WORK_DIR/help.out"
diff -u "$WORK_DIR/expected_usage.err" "$WORK_DIR/help.err"

PATH="$TEST_PATH" "$BIN_PATH" accounts alpha beta > "$WORK_DIR/dispatch.out" 2> "$WORK_DIR/dispatch.err"
diff -u <(printf 'ACCOUNTS:alpha beta\n') "$WORK_DIR/dispatch.out"
! test -s "$WORK_DIR/dispatch.err"

set +e
PATH="$TEST_PATH" "$BIN_PATH" fail sample > "$WORK_DIR/fail.out" 2> "$WORK_DIR/fail.err"
fail_code=$?
set -e

test "$fail_code" -eq 7
diff -u <(printf 'FAIL:sample\n') "$WORK_DIR/fail.out"
! test -s "$WORK_DIR/fail.err"

echo "e2e_bus_bus.sh: PASS"
