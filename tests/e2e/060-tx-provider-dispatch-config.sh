#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "060"
setup_dispatch_fixtures

cat > "$WS/tx-config.bus" <<'EOF_TX_CONFIG'
accounts list
EOF_TX_CONFIG
cat > "$WS/datapackage.json" <<'EOF_DP_FALLBACK'
{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"file","fallback_to_none":true}}}}
EOF_DP_FALLBACK
(cd "$WS" && PATH="$TEST_PATH" "$BIN" tx-config.bus > "$WS/tx_fallback.out" 2> "$WS/tx_fallback.err")
grep -q 'ACCOUNTS:list' "$WS/tx_fallback.out"
grep -q 'falling back to "none"' "$WS/tx_fallback.err"

cat > "$WS/datapackage.json" <<'EOF_DP_STRICT'
{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"file","fallback_to_none":false}}}}
EOF_DP_STRICT
set +e
(cd "$WS" && PATH="$TEST_PATH" "$BIN" tx-config.bus > "$WS/tx_strict.out" 2> "$WS/tx_strict.err")
tx_strict_code=$?
set -e

test "$tx_strict_code" -eq 2
! test -s "$WS/tx_strict.out"
grep -q 'provider "fs" requires in-process tx runners' "$WS/tx_strict.err"

cat > "$WS/datapackage.json" <<'EOF_DP_SHELL_OFF'
{"bus":{"busfile":{"dispatch":{"shell_lookup_enabled":false}}}}
EOF_DP_SHELL_OFF
set +e
(cd "$WS" && PATH="$TEST_PATH" "$BIN" tx-config.bus > "$WS/shell_off.out" 2> "$WS/shell_off.err")
shell_off_code=$?
set -e

test "$shell_off_code" -eq 127
! test -s "$WS/shell_off.out"
grep -q 'shell lookup disabled and no in-process runner' "$WS/shell_off.err"

echo "e2e OK"
