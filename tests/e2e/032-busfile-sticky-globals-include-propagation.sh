#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "032"
setup_dispatch_fixtures

cat > "$WS/another.bus" <<'EOF_ANOTHER'
accounts beta
accounts gamma
EOF_ANOTHER

cat > "$WS/all.bus" <<'EOF_ALL'
--perf
--chdir data
another.bus
--no-perf
accounts delta
EOF_ALL

PATH="$TEST_PATH" "$BIN" "$WS/all.bus" > "$WS/include.out" 2> "$WS/include.err"
diff -u <(printf 'ACCOUNTS:--chdir data beta\nACCOUNTS:--chdir data gamma\nACCOUNTS:--chdir data delta\n') "$WS/include.out"
grep -q '^INFO perf bus-accounts beta ' "$WS/include.err"
grep -q '^INFO perf bus-accounts gamma ' "$WS/include.err"
! grep -q 'bus-accounts delta' "$WS/include.err"

echo "e2e OK"
