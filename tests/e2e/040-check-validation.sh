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

cat > "$WS/check_journal_row_desc.bus" <<'EOF_CHECK_JOURNAL_ROW_DESC'
journal add --date 2024-10-31 --desc test --debit '1911=924.10=Asiakkaan maksusuoritus pankkiin' --credit '3001=924.10=Oma hostingpalvelu HG-asiakkaalle'
EOF_CHECK_JOURNAL_ROW_DESC
PATH="$TEST_PATH" "$BIN" --check "$WS/check_journal_row_desc.bus" > "$WS/check_journal_row_desc.out" 2> "$WS/check_journal_row_desc.err"
! test -s "$WS/check_journal_row_desc.out"
! test -s "$WS/check_journal_row_desc.err"

cat > "$WS/check_journal_row_desc_quoted_punctuation.bus" <<'EOF_CHECK_JOURNAL_ROW_DESC_QUOTED_PUNCT'
journal add --date 2024-10-31 --desc test --debit '1911=924.10=Asiakkaan maksusuoritus + alv' --credit '3001=924.10=Muistutusmaksut Reminder Fee -rivistä'
EOF_CHECK_JOURNAL_ROW_DESC_QUOTED_PUNCT
PATH="$TEST_PATH" "$BIN" --check "$WS/check_journal_row_desc_quoted_punctuation.bus" > "$WS/check_journal_row_desc_quoted_punctuation.out" 2> "$WS/check_journal_row_desc_quoted_punctuation.err"
! test -s "$WS/check_journal_row_desc_quoted_punctuation.out"
! test -s "$WS/check_journal_row_desc_quoted_punctuation.err"

cat > "$WS/check_journal_row_desc_quoted_semicolon.bus" <<'EOF_CHECK_JOURNAL_ROW_DESC_QUOTED_SEMICOLON'
journal add --date 2024-10-31 --desc test --debit '1911=924.10=Titan 1 GB foo; Rekisteröintimaksu bar' --credit '3001=924.10=Reminder Fee; collection step'
EOF_CHECK_JOURNAL_ROW_DESC_QUOTED_SEMICOLON
PATH="$TEST_PATH" "$BIN" --check "$WS/check_journal_row_desc_quoted_semicolon.bus" > "$WS/check_journal_row_desc_quoted_semicolon.out" 2> "$WS/check_journal_row_desc_quoted_semicolon.err"
! test -s "$WS/check_journal_row_desc_quoted_semicolon.out"
! test -s "$WS/check_journal_row_desc_quoted_semicolon.err"

cat > "$WS/check_journal_row_desc_unquoted.bus" <<'EOF_CHECK_JOURNAL_ROW_DESC_UNQUOTED'
journal add --date 2024-10-31 --desc test --debit 1911=924.10=Asiakkaan maksusuoritus pankkiin --credit 3001=924.10=Oma hostingpalvelu HG-asiakkaalle
EOF_CHECK_JOURNAL_ROW_DESC_UNQUOTED
PATH="$TEST_PATH" "$BIN" --check "$WS/check_journal_row_desc_unquoted.bus" > "$WS/check_journal_row_desc_unquoted.out" 2> "$WS/check_journal_row_desc_unquoted.err"
! test -s "$WS/check_journal_row_desc_unquoted.out"
! test -s "$WS/check_journal_row_desc_unquoted.err"

cat > "$WS/check_journal_plain_posting.bus" <<'EOF_CHECK_JOURNAL_PLAIN_POSTING'
journal add --date 2024-10-31 --desc test --debit 1911=924.10 --credit 3001=924.10 --source-id b26197 --source-entry 1
EOF_CHECK_JOURNAL_PLAIN_POSTING
PATH="$TEST_PATH" "$BIN" --check "$WS/check_journal_plain_posting.bus" > "$WS/check_journal_plain_posting.out" 2> "$WS/check_journal_plain_posting.err"
! test -s "$WS/check_journal_plain_posting.out"
! test -s "$WS/check_journal_plain_posting.err"

echo "e2e OK"
