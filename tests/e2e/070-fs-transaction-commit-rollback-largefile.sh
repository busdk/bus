#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "070"
setup_dispatch_fixtures

cat > "$WS/datapackage.json" <<'EOF_DP_FS_BATCH'
{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"batch","fallback_to_none":false}}}}
EOF_DP_FS_BATCH
cat > "$WS/fs_ok.bus" <<'EOF_FS_OK'
txwrite fs/data.txt one
txwrite fs/data.txt two
EOF_FS_OK
(cd "$WS" && BUS_TEST_ENABLE_TXWRITE=1 PATH="$TEST_PATH" "$BIN" fs_ok.bus > "$WS/fs_ok.out" 2> "$WS/fs_ok.err")
! test -s "$WS/fs_ok.out"
! test -s "$WS/fs_ok.err"
diff -u <(printf 'one\ntwo\n') "$WS/fs/data.txt"

cat > "$WS/fs_fail.bus" <<'EOF_FS_FAIL'
txwrite fs/rollback.txt one
txwrite fs/rollback.txt fail
EOF_FS_FAIL
set +e
(cd "$WS" && BUS_TEST_ENABLE_TXWRITE=1 PATH="$TEST_PATH" "$BIN" fs_fail.bus > "$WS/fs_fail.out" 2> "$WS/fs_fail.err")
fs_fail_code=$?
set -e

test "$fs_fail_code" -eq 1
! test -s "$WS/fs_fail.out"
grep -q 'command failed (exit 1)' "$WS/fs_fail.err"
! test -e "$WS/fs/rollback.txt"

awk 'BEGIN { for (i = 0; i < 100000; i++) printf("%d,alpha,beta\\n", i) }' > "$WS/fs/big.csv"
cat > "$WS/fs_big.bus" <<'EOF_FS_BIG'
txwrite fs/big.csv 100000,omega,zeta
EOF_FS_BIG
(cd "$WS" && BUS_TEST_ENABLE_TXWRITE=1 PATH="$TEST_PATH" "$BIN" fs_big.bus > "$WS/fs_big.out" 2> "$WS/fs_big.err")
! test -s "$WS/fs_big.out"
! test -s "$WS/fs_big.err"
diff -u <(printf '100000,omega,zeta\n') <(tail -n 1 "$WS/fs/big.csv")

REAL_FS_DIR="$WS/fs_real_module"
mkdir -p "$REAL_FS_DIR"
cat > "$REAL_FS_DIR/datapackage.json" <<'EOF_REAL_FS_DP'
{"bus":{"busfile":{"transaction":{"provider":"fs","scope":"batch","fallback_to_none":false}}}}
EOF_REAL_FS_DP
cat > "$REAL_FS_DIR/ok.bus" <<'EOF_REAL_FS_OK'
bank init
EOF_REAL_FS_OK
set +e
(cd "$REAL_FS_DIR" && PATH="$TEST_PATH" "$BIN" ok.bus > "$REAL_FS_DIR/ok.out" 2> "$REAL_FS_DIR/ok.err")
real_fs_code=$?
set -e

test "$real_fs_code" -eq 127
! test -s "$REAL_FS_DIR/ok.out"
grep -q 'dispatch error: unknown target "bank"' "$REAL_FS_DIR/ok.err"
! test -e "$REAL_FS_DIR/bank-imports.csv"
! test -e "$REAL_FS_DIR/bank-transactions.csv"

echo "e2e OK"
