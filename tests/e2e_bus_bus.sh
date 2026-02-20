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

cat > "$WORK_DIR/path_first/bus-status" <<'SH'
#!/bin/sh
for arg in "$@"; do
  if [ "$arg" = "--version" ]; then
    printf 'bus-status e2e\n'
    exit 0
  fi
done
printf 'STATUS:%s\n' "$*"
SH
chmod +x "$WORK_DIR/path_first/bus-status"

cat > "$WORK_DIR/path_first/bus-journal" <<'SH'
#!/bin/sh
printf 'JOURNAL:%s\n' "$*"
SH
chmod +x "$WORK_DIR/path_first/bus-journal"

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
  journal
  ledger
  status
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
  journal
  ledger
  status
EOF_EXPECT_MISSING

set +e
PATH="$TEST_PATH" "$BIN_PATH" missing > "$WORK_DIR/missing.out" 2> "$WORK_DIR/missing.err"
missing_code=$?
set -e

test "$missing_code" -eq 127
! test -s "$WORK_DIR/missing.out"
diff -u "$WORK_DIR/expected_missing.err" "$WORK_DIR/missing.err"
grep -q '^bus: missing subcommand:' "$WORK_DIR/missing.err"

PATH="$TEST_PATH" "$BIN_PATH" --help > "$WORK_DIR/help_global.out" 2> "$WORK_DIR/help_global.err"
grep -q '^usage: bus \[global-flags\] <command> \[args...\]$' "$WORK_DIR/help_global.out"
grep -q '^  -C, --chdir <dir>$' "$WORK_DIR/help_global.out"
grep -q '^available commands:$' "$WORK_DIR/help_global.out"
! test -s "$WORK_DIR/help_global.err"

PATH="$TEST_PATH" "$BIN_PATH" --version > "$WORK_DIR/version_global.out" 2> "$WORK_DIR/version_global.err"
diff -u <(printf 'bus dev\n') "$WORK_DIR/version_global.out"
! test -s "$WORK_DIR/version_global.err"

PATH="$TEST_PATH" "$BIN_PATH" -C / status --version > "$WORK_DIR/global_chdir.out" 2> "$WORK_DIR/global_chdir.err"
diff -u <(printf 'bus-status e2e\n') "$WORK_DIR/global_chdir.out"
! test -s "$WORK_DIR/global_chdir.err"

PATH="$TEST_PATH" "$BIN_PATH" -q status --version > "$WORK_DIR/global_quiet.out" 2> "$WORK_DIR/global_quiet.err"
diff -u <(printf 'bus-status e2e\n') "$WORK_DIR/global_quiet.out"
! test -s "$WORK_DIR/global_quiet.err"

PATH="$TEST_PATH" "$BIN_PATH" -- status --version > "$WORK_DIR/global_dd.out" 2> "$WORK_DIR/global_dd.err"
diff -u <(printf 'bus-status e2e\n') "$WORK_DIR/global_dd.out"
! test -s "$WORK_DIR/global_dd.err"

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

cat > "$WORK_DIR/2024-01.bus" <<'EOF_202401'
accounts jan
EOF_202401

cat > "$WORK_DIR/2024-02.bus" <<'EOF_202402'
journal feb
EOF_202402

cat > "$WORK_DIR/all.bus" <<'EOF_ALL'
2024-01.bus
2024-02.bus
EOF_ALL

PATH="$TEST_PATH" "$BIN_PATH" "$WORK_DIR/all.bus" > "$WORK_DIR/busfile.out" 2> "$WORK_DIR/busfile.err"
diff -u <(printf 'ACCOUNTS:jan\nJOURNAL:feb\n') "$WORK_DIR/busfile.out"
! test -s "$WORK_DIR/busfile.err"

cat > "$WORK_DIR/bad.bus" <<'EOF_BAD'
accounts 'unterminated
EOF_BAD

set +e
PATH="$TEST_PATH" "$BIN_PATH" "$WORK_DIR/2024-01.bus" "$WORK_DIR/bad.bus" > "$WORK_DIR/preflight.out" 2> "$WORK_DIR/preflight.err"
preflight_code=$?
set -e

test "$preflight_code" -eq 65
! test -s "$WORK_DIR/preflight.out"
grep -q '^.*/bad.bus:1: syntax error:' "$WORK_DIR/preflight.err"

PATH="$TEST_PATH" "$BIN_PATH" --check "$WORK_DIR/2024-01.bus" > "$WORK_DIR/check.out" 2> "$WORK_DIR/check.err"
! test -s "$WORK_DIR/check.out"
! test -s "$WORK_DIR/check.err"

PATH="$TEST_PATH" "$BIN_PATH" --trace "$WORK_DIR/2024-01.bus" > "$WORK_DIR/trace.out" 2> "$WORK_DIR/trace.err"
grep -q '2024-01.bus:1: bus accounts jan' "$WORK_DIR/trace.out"
grep -q 'ACCOUNTS:jan' "$WORK_DIR/trace.out"
! test -s "$WORK_DIR/trace.err"

set +e
PATH="$TEST_PATH" "$BIN_PATH" fail sample > "$WORK_DIR/fail.out" 2> "$WORK_DIR/fail.err"
fail_code=$?
set -e

test "$fail_code" -eq 7
diff -u <(printf 'FAIL:sample\n') "$WORK_DIR/fail.out"
! test -s "$WORK_DIR/fail.err"

echo "e2e_bus_bus.sh: PASS"
