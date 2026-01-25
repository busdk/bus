#!/bin/bash
cd "$(dirname "$0")/.."
set -e
#set -x

MODEL=gpt-5.2
OUTPUT_FORMAT=stream-json
TASK_TIMEOUT=15m
MDC_FILE=.cursor/rules/go-commit.mdc

echo
echo "--- COMMITTING UNCHANGED TO GIT ---"
echo

if timeout "$TASK_TIMEOUT" cursor-agent -p --output-format "$OUTPUT_FORMAT" -f --model "$MODEL" -- "$(cat "$MDC_FILE")"; then
  echo
  echo '--- SUCCESSFUL COMMIT ---'
  echo
else
  ERRNO="$?"
  echo
  echo '--- COMMIT ERROR:'"$ERRNO"' ---'
  echo
  exit 1
fi
