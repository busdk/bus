#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "031"
setup_dispatch_fixtures

cat > "$WS/sticky.bus" <<'EOF_STICKY'
--perf
-v
--chdir data
accounts alpha
--chdir reports
--no-verbose
ledger beta
--no-perf
accounts gamma
--quiet
accounts delta
--no-quiet
accounts epsilon
EOF_STICKY

PATH="$TEST_PATH" "$BIN" "$WS/sticky.bus" > "$WS/sticky.out" 2> "$WS/sticky.err"
diff -u <(printf 'ACCOUNTS:-v --chdir data alpha\nLEDGER:--chdir reports beta\nACCOUNTS:--chdir reports gamma\nACCOUNTS:--quiet --chdir reports delta\nACCOUNTS:--chdir reports epsilon\n') "$WS/sticky.out"
grep -q '^INFO perf bus-accounts alpha ' "$WS/sticky.err"
grep -q '^INFO perf bus-ledger beta ' "$WS/sticky.err"
! grep -q 'bus-accounts gamma' "$WS/sticky.err"
! grep -q 'bus-accounts delta' "$WS/sticky.err"
! grep -q 'bus-accounts epsilon' "$WS/sticky.err"

echo "e2e OK"
