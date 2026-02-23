#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "040"
setup_dispatch_fixtures

cat > "$WS/2024-01.bus" <<'EOF_202401'
accounts jan
EOF_202401

PATH="$TEST_PATH" "$BIN" --check "$WS/2024-01.bus" > "$WS/check.out" 2> "$WS/check.err"
! test -s "$WS/check.out"
! test -s "$WS/check.err"

cat > "$WS/check_unbalanced.bus" <<'EOF_CHECK_UNBAL'
journal add --date 2024-02-29 --debit 1910=10.00 --credit 3000=9.99
EOF_CHECK_UNBAL
set +e
PATH="$TEST_PATH" "$BIN" --check "$WS/check_unbalanced.bus" > "$WS/check_unbalanced.out" 2> "$WS/check_unbalanced.err"
check_unbalanced_code=$?
set -e

test "$check_unbalanced_code" -eq 1
! test -s "$WS/check_unbalanced.out"
grep -q 'validation error: journal add unbalanced entry' "$WS/check_unbalanced.err"

cat > "$WS/check_bank_invalid.bus" <<'EOF_CHECK_BANK_INVALID'
bank add transactions --set booked_date=2024-99-99 --set amount=NaN --set currency=EURO
EOF_CHECK_BANK_INVALID
set +e
PATH="$TEST_PATH" "$BIN" --check "$WS/check_bank_invalid.bus" > "$WS/check_bank_invalid.out" 2> "$WS/check_bank_invalid.err"
check_bank_invalid_code=$?
set -e

test "$check_bank_invalid_code" -eq 1
! test -s "$WS/check_bank_invalid.out"
grep -q 'validation error: bank add transactions invalid booked_date' "$WS/check_bank_invalid.err"

echo "e2e OK"
