#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
TEST_SUBJECT="${ROOT_DIR}/bin/bus"

hash_file() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
    return
  fi
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi
  cksum "$1" | awk '{print $1 ":" $2}'
}

assert_subject_unchanged() {
  local got
  got="$(hash_file "$TEST_SUBJECT")"
  [[ "$got" == "$EXPECTED_SUBJECT_HASH" ]] || {
    echo "e2e failed: test subject changed during run: ${TEST_SUBJECT}" >&2
    echo "expected hash: ${EXPECTED_SUBJECT_HASH}" >&2
    echo "actual hash:   ${got}" >&2
    exit 1
  }
}

[[ -x "$TEST_SUBJECT" ]] || {
  echo "e2e failed: test subject missing or not executable: ${TEST_SUBJECT}" >&2
  exit 1
}
EXPECTED_SUBJECT_HASH="$(hash_file "$TEST_SUBJECT")"

while IFS= read -r script; do
  assert_subject_unchanged
  (if [[ "${BUS_E2E_VERBOSE:-0}" = "1" ]]; then set -x; fi; bash "$script")
  assert_subject_unchanged
done < <(find "${ROOT_DIR}/tests/e2e" -maxdepth 1 -type f -name '[0-9][0-9][0-9]-*.sh' | LC_ALL=C sort)

"$TEST_SUBJECT" help --format opencli | grep -q '"io.busdk.environment"'
"$TEST_SUBJECT" help --format opencli | grep -q '"title": "bus"'
echo "e2e.sh: PASS"
