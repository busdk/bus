#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "021"
setup_dispatch_fixtures

cat > "$WS/.env" <<'ENV'
# working-directory dispatcher environment
BUS_E2E_DOTENV=loaded
export BUS_E2E_DOTENV_EXPORTED=exported
BUS_E2E_DOTENV_SPACED = spaced value
BUS_E2E_DOTENV_EXISTING=from-dotenv
GENERIC_E2E_DOTENV=generic
ENV

(cd "$WS" && PATH="$TEST_PATH" BUS_E2E_DOTENV_EXISTING=from-process "$BIN" env BUS_E2E_DOTENV BUS_E2E_DOTENV_EXPORTED BUS_E2E_DOTENV_SPACED BUS_E2E_DOTENV_EXISTING GENERIC_E2E_DOTENV > "$WS/dotenv.out" 2> "$WS/dotenv.err")
diff -u <(printf 'BUS_E2E_DOTENV=loaded\nBUS_E2E_DOTENV_EXPORTED=exported\nBUS_E2E_DOTENV_SPACED=spaced value\nBUS_E2E_DOTENV_EXISTING=from-process\nGENERIC_E2E_DOTENV=generic\n') "$WS/dotenv.out"
! test -s "$WS/dotenv.err"

mkdir -p "$WS/app"
printf 'BUS_E2E_DOTENV_CHDIR=from-chdir\n' > "$WS/app/.env"
PATH="$TEST_PATH" "$BIN" -C "$WS/app" env BUS_E2E_DOTENV_CHDIR > "$WS/dotenv_chdir.out" 2> "$WS/dotenv_chdir.err"
diff -u <(printf 'BUS_E2E_DOTENV_CHDIR=from-chdir\n') "$WS/dotenv_chdir.out"
! test -s "$WS/dotenv_chdir.err"

printf '1INVALID=value\n' > "$WS/.env"
if (cd "$WS" && PATH="$TEST_PATH" "$BIN" env BUS_E2E_DOTENV > "$WS/dotenv_invalid.out" 2> "$WS/dotenv_invalid.err"); then
  fail "invalid .env unexpectedly succeeded"
fi
! test -s "$WS/dotenv_invalid.out"
grep -q 'bus: failed to load .env: .env:1: invalid environment variable name "1INVALID"' "$WS/dotenv_invalid.err"

echo "e2e OK"
