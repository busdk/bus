#!/bin/bash
cd "$(dirname "$0")/.."
set -e
#set -x

NAME="$(basename "$(pwd)")"

MODEL=gpt-5.2-codex-high
OUTPUT_FORMAT=stream-json
TASK_TIMEOUT=60m
MDC_FILE=".cursor/rules/$NAME.mdc"

if test -f "$MDC_FILE"; then
  :
else
  echo "ERROR: No MDC file found: $MDC_FILE"
  exit 2
fi

echo
echo "--- refine-mdc-spec: $NAME ---"
echo

# Build an expanded, concrete prompt (no ${NAME} placeholders)
PROMPT="$(
  cat "$MDC_FILE"
  printf "\n\n"
  # Expand {{MODULE_NAME}} and {{MDC_FILE}} inside the embedded prompt
  awk 'f{print} /^### PROMPT ###$/{f=1}' "$0" \
    | sed "s/{{MODULE_NAME}}/$NAME/g" \
    | sed "s|{{MDC_FILE}}|$MDC_FILE|g"
)"

if timeout "$TASK_TIMEOUT" cursor-agent -p --output-format "$OUTPUT_FORMAT" -f \
    --model "$MODEL" agent -- \
    "$PROMPT" \
    | ./scripts/format-cursor-log.sh --only-roles=assistant,user,system; then
  echo
  echo "--- SUCCESSFUL: $NAME ---"
  echo
else
  ERRNO="$?"
  echo
  echo "--- ERROR: $NAME: $ERRNO ---"
  echo
  exit 1
fi

exit 0
### PROMPT ###
------

Goal. Refine only the Cursor MDC rule file at {{MDC_FILE}} for module 
{{MODULE_NAME}} so that it accurately reflects the latest BusDK specifications 
at https://docs.busdk.com/. Do not change source code, tests, schemas, 
README.md, or any other files in this task. The task is complete only when the 
MDC file is updated and saved with improved, spec-aligned guidance.

Spec review. Read and use the BusDK documentation as the primary reference. 
Start from https://docs.busdk.com/ and locate the most relevant pages for this 
module, including general cross-cutting specs (data formats and storage, CSV 
conventions, Table Schema contract, CLI workflow, error handling, dry-run 
behavior, diagnostics conventions, repository layout and README expectations) 
and any module-specific spec pages that apply to this repository. Prefer 
linking to the most specific and canonical pages you can find for the rules you 
are encoding.

Refinement approach. Compare the current MDC content against the spec and 
update it so it is correct, unambiguous, and immediately actionable for an AI 
agent working inside this repo. Keep the MDC short, deterministic, and focused 
on what the agent must do and must not do. Where the BusDK spec already defines 
a rule, prefer a short statement plus a link instead of re-explaining the full 
spec. Where this module has unique requirements beyond shared conventions, 
spell them out clearly in the MDC.

Conflict handling. Treat the docs as authoritative but not infallible. If you 
find a conflict between the repository’s current reality (file layout, naming, 
existing behavior) and the written spec, do not guess. In the MDC, document the 
mismatch clearly as a small dedicated section describing what the repo does 
today versus what the spec says, and what the agent should follow going forward 
for new work. Do not attempt to fix the mismatch in code in this task; just 
document it inside the MDC.

MDC structure discipline. Preserve valid MDC frontmatter and keep alwaysApply 
and globs behavior consistent with how this repo uses Cursor rules. Ensure the 
description remains accurate for this module. Ensure the file is 
well-formatted, with a final newline, and uses the same tone and conventions as 
other BusDK module MDC files (direct instructions, deterministic constraints, 
and minimal ambiguity).

Module scope. Ensure the MDC states the exact purpose of this module, its 
inputs and outputs at a high level (datasets, schemas, and side effects), and 
its explicit non-goals. Ensure the MDC states how this module is invoked from 
the bus dispatcher (binary name and invocation pattern), and how it should 
behave regarding stdout/stderr, exit codes, and dry-run.

Quality gates. Ensure the MDC instructs the agent to follow BusDK’s 
deterministic workflow expectations, with fast hermetic tests, no network 
dependence, and a standard Makefile interface (build, test, lint, fmt) if that 
is part of the BusDK spec for modules. Only include requirements that are 
supported by the spec or clearly established by this repo’s conventions.

Deliverable. Save the refined MDC file at {{MDC_FILE}}. After updating it, 
provide a short summary of what changed and which spec pages were most 
important, so reviewers can verify the refinement quickly.
