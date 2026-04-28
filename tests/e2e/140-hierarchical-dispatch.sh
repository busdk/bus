#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "140"

mkdir -p "$WS/path_first"

cat > "$WS/path_first/bus-operator" <<'SH'
#!/bin/sh
printf 'OPERATOR:%s\n' "$*"
SH
chmod +x "$WS/path_first/bus-operator"

cat > "$WS/path_first/bus-operator-billing" <<'SH'
#!/bin/sh
printf 'BILLING:%s\n' "$*"
SH
chmod +x "$WS/path_first/bus-operator-billing"

PATH="$WS/path_first" "$BIN" operator billing catalog sync > "$WS/nested.out" 2> "$WS/nested.err"
diff -u <(printf 'OPERATOR:billing catalog sync\n') "$WS/nested.out"
! test -s "$WS/nested.err"

PATH="$WS/path_first" "$BIN" --perf operator billing catalog sync > "$WS/perf.out" 2> "$WS/perf.err"
diff -u <(printf 'OPERATOR:billing catalog sync\n') "$WS/perf.out"
grep -q '^INFO perf bus-operator billing ' "$WS/perf.err"

rm "$WS/path_first/bus-operator-billing"
PATH="$WS/path_first" "$BIN" operator billing catalog sync > "$WS/direct.out" 2> "$WS/direct.err"
diff -u <(printf 'OPERATOR:billing catalog sync\n') "$WS/direct.out"
! test -s "$WS/direct.err"

echo "e2e OK"
