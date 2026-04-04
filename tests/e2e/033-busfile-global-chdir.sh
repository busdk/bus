#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "033"
setup_dispatch_fixtures

mkdir -p "$WS/workspace/data"
cat > "$WS/workspace/replay.bus" <<'EOF_REPLAY'
accounts alpha
EOF_REPLAY

(cd "$WS" && PATH="$TEST_PATH" "$BIN" -C workspace/data ../replay.bus > "$WS/chdir_busfile.out" 2> "$WS/chdir_busfile.err")
diff -u <(printf 'ACCOUNTS:alpha\n') "$WS/chdir_busfile.out"
! test -s "$WS/chdir_busfile.err"

echo "e2e OK"
