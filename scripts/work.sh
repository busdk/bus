#!/bin/bash
cd "$(dirname "$0")/.."
set -e
#set -x

NAME="$(basename "$(pwd)")"

MODEL=gpt-5.2-codex-high
OUTPUT_FORMAT=stream-json
TASK_TIMEOUT=15m
MDC_FILE=".cursor/rules/$NAME.mdc"

if test -f $MDC_FILE; then
  :
else
  echo 'ERROR: No MDC file found: '"$MDC_FILE"
  exit 2
fi

echo
echo "--- COMMITTING UNCHANGED TO GIT ---"
echo

if timeout "$TASK_TIMEOUT" cursor-agent -p --output-format "$OUTPUT_FORMAT" -f --model "$MODEL" -- "$(cat "$MDC_FILE")\nCommit all staged changes using as semantically small commits as possible with meaningful commit messages. Do nothing else."; then
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
