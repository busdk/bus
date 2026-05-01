#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BIN="${ROOT_DIR}/bin/bus"
WS=""

fail() {
  echo "e2e failed: $*" >&2
  exit 1
}

e2e_cleanup() {
  if [ -n "${WS}" ] && [ "${BUS_E2E_KEEP:-0}" != "1" ]; then
    rm -rf "$WS"
  fi
}

e2e_setup() {
  local case_name="$1"
  if [ ! -x "$BIN" ]; then
    fail "binary not found or not executable: ${BIN} (run: make build)"
  fi
  WS="${ROOT_DIR}/tests/e2e_bus_bus_workspace_${case_name}"
  rm -rf "$WS"
  mkdir -p "$WS"
  trap e2e_cleanup EXIT
}

setup_dispatch_fixtures() {
  mkdir -p "$WS/path_first" "$WS/path_second"

  cat > "$WS/path_first/bus-accounts" <<'SH'
#!/bin/sh
printf 'ACCOUNTS:%s\n' "$*"
SH
  chmod +x "$WS/path_first/bus-accounts"

  cat > "$WS/path_second/bus-ledger" <<'SH'
#!/bin/sh
printf 'LEDGER:%s\n' "$*"
SH
  chmod +x "$WS/path_second/bus-ledger"

  cat > "$WS/path_first/bus-fail" <<'SH'
#!/bin/sh
printf 'FAIL:%s\n' "$*"
exit 7
SH
  chmod +x "$WS/path_first/bus-fail"

  cat > "$WS/path_second/bus-accounts" <<'SH'
#!/bin/sh
printf 'SECOND:%s\n' "$*"
SH
  chmod +x "$WS/path_second/bus-accounts"

  cat > "$WS/path_first/bus-status" <<'SH'
#!/bin/sh
for arg in "$@"; do
  if [ "$arg" = "--version" ]; then
    printf 'bus-status e2e\n'
    exit 0
  fi
done
printf 'STATUS:%s\n' "$*"
SH
  chmod +x "$WS/path_first/bus-status"

  cat > "$WS/path_first/bus-journal" <<'SH'
#!/bin/sh
printf 'JOURNAL:%s\n' "$*"
SH
  chmod +x "$WS/path_first/bus-journal"

  cat > "$WS/path_first/bus-env" <<'SH'
#!/bin/sh
skip_next=0
for key in "$@"; do
  if [ "$skip_next" = "1" ]; then
    skip_next=0
    continue
  fi
  if [ "$key" = "--chdir" ]; then
    skip_next=1
    continue
  fi
  eval "value=\${$key-}"
  printf '%s=%s\n' "$key" "$value"
done
SH
  chmod +x "$WS/path_first/bus-env"

  cat > "$WS/path_first/bus-nonexec" <<'TXT'
ignored
TXT

  TEST_PATH="${WS}/path_first:${WS}/path_second"
}
