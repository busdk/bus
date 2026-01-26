#!/bin/bash
cd "$(dirname "$0")/.."
set -e
#set -x

if git status|grep -qE 'nothing to commit, working tree clean'; then
  echo
  echo "--- nothing to commit, working tree clean ---"
  echo
  exit 0
fi

MODEL=gpt-5.2
OUTPUT_FORMAT=stream-json
TASK_TIMEOUT=15m
MDC_FILE=.cursor/rules/go-commit.mdc

echo
echo "--- COMMITTING UNCHANGED TO GIT ---"
echo

if timeout "$TASK_TIMEOUT" cursor-agent -p --output-format "$OUTPUT_FORMAT" -f --model "$MODEL" -- "$(cat "$MDC_FILE")\nCommit all staged changes using as semantically small commits as possible with meaningful commit messages. Do nothing else." \
    |./scripts/format-cursor-log.sh --only-roles=assistant,user,system; then
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
