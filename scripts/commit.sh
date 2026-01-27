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

echo
echo "--- COMMITTING UNCHANGED TO GIT ---"
echo

if timeout "$TASK_TIMEOUT" cursor-agent -p --output-format "$OUTPUT_FORMAT" \
      -f --model "$MODEL" agent \
      -- "$(awk 'f{print} /^### PROMPT ###$/{f=1}' "$0")" \
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
exit 0
### PROMPT ###

This rule governs a commit-only workflow. The assistant must operate strictly on
the current Git index (staging area) of the local working copy. The assistant
must not interact with any remotes or networks in any way, including (but not
limited to) push, pull, fetch, clone, submodule update/init, remote add/set-url,
or any operation that could contact a remote.

Scope: The assistant’s only allowed state-changing operation is creating new
commits from already-staged changes. The assistant must not modify files, must
not auto-stage anything, must not amend commits, must not rebase, and must not
rewrite history. If there are no staged changes anywhere, the assistant must do
nothing and say so.

Submodules: If the repository contains Git submodules, the assistant commits
staged changes in depth-first order:
1) For each submodule that itself contains staged changes, process that submodule
   first. If that submodule contains nested submodules with staged changes,
   process those nested submodules first, recursively.
2) Enter the submodule (logically, via its path) and commit only what is already
   staged in that submodule. Do not stage, reset, checkout, update, initialize,
   or reconfigure submodules.
3) After committing inside a submodule, return to the parent repository and
   continue until all submodules with staged changes have been committed.

Superproject: After all relevant submodule commits are complete, the assistant
may commit the containing repository (superproject) only if the superproject
itself has staged changes. This may include staged gitlink pointer updates. If a
submodule commit caused the superproject to have an unstaged gitlink change that
must be recorded, the assistant must explicitly point this out and stop short of
staging it. The assistant must never stage gitlink updates automatically.

Per-commit review and atomicity: Before creating each commit, the assistant must
review the staged changes at a high level (what files and what intent). If the
staged set appears to contain multiple logical changes, the assistant must
propose an atomic commit split plan. The assistant must not alter the staging to
perform the split unless explicitly instructed; by default it commits exactly
what is currently staged.

Commit messages: For every commit, write a concise, action-oriented subject line
in the imperative mood. Optionally include a body separated by a blank line. The
message should explain what changed and why, mention user-visible impact or risk
when relevant, and include traceability (issue IDs or full URLs) when helpful.
Avoid vague summaries. Conventional prefixes such as feat, fix, docs, refactor,
test, or chore may be used when they improve clarity, but never at the expense
of a precise subject.

Failures and hooks: If a commit is rejected by hooks, commit-msg policies, or
checks (including inside any submodule), the assistant must report the exact
failure reason/output and suggest the minimal correction needed to satisfy the
policy. The assistant must not retry automatically unless explicitly instructed.

Hard prohibitions: Regardless of outcome, the assistant must not push, tag,
publish, synchronize, or perform any other remote operation. Pushing is out of
scope for this rule and must be left for manual action or separate rules.

Instruction: Commit all currently staged changes using the smallest semantically
meaningful commits possible with high-quality messages. Do nothing else.
