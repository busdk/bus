#!/bin/bash
cd "$(dirname "$0")/.."
set -e
#set -x

NAME="$(basename "$(pwd)")"

MODEL=gpt-5.2-codex-high
OUTPUT_FORMAT=stream-json
TASK_TIMEOUT=60m
MDC_FILE=".cursor/rules/$NAME.mdc"

if test -f $MDC_FILE; then
  :
else
  echo 'ERROR: No MDC file found: '"$MDC_FILE"
  exit 2
fi

echo
echo "--- $NAME ---"
echo

if timeout "$TASK_TIMEOUT" cursor-agent -p --output-format "$OUTPUT_FORMAT" -f \
    --model "$MODEL" agent -- \
    "$(cat "$MDC_FILE")\n\n$(awk 'f{print} /^### PROMPT ###$/{f=1}' "$0")" \
    |./scripts/format-cursor-log.sh --only-roles=assistant,user,system; then
  echo
  echo '--- SUCCESSFUL:$NAME ---'
  echo
else
  ERRNO="$?"
  echo
  echo '--- ERROR:$NAME:'"$ERRNO"' ---'
  echo
  exit 1
fi

exit 0
### PROMPT ###
------

Project specification and sources. Implement this module according to the BusDK 
specifications and conventions at https://docs.busdk.com/. Treat the 
documentation as the primary reference, but not as infallible: if you find a 
conflict between the repository’s reality and the written spec, do not guess. 
Document the mismatch clearly (what the repo does today vs what the spec says), 
then implement the smallest, most deterministic change that preserves 
auditable, agent-friendly BusDK workflows. At minimum, align with these specs 
and link them from README.md where relevant: 
https://docs.busdk.com/spec/03-data-formats-and-storage, 
https://docs.busdk.com/spec/data/csv-conventions, 
https://docs.busdk.com/spec/data/table-schema-contract, 
https://docs.busdk.com/spec/04-cli-workflow, 
https://docs.busdk.com/spec/cli/error-handling-dry-run-diagnostics, and 
https://docs.busdk.com/spec/layout/repository-readme-expectations. If a 
module-specific spec page exists for this repository, follow it and link it 
prominently in the README.

Definition of Done. Every task must be completed as a single coherent change 
that includes working implementation, unit tests, and README documentation 
together. Do not defer tests or documentation. The task is not done until 
implementation, tests, and documentation exist in the same change and all 
checks pass.

Implementation and unit testing discipline. Keep the CLI entrypoint thin and 
push behavior into testable packages/functions that can be called directly by 
unit tests. Avoid global state. Do not call os.Exit outside main; return 
errors/status codes so tests can execute deterministically. Tests must be fast 
and hermetic: no network, no external services, and no reliance on the 
developer machine’s state. Use temporary directories for filesystem work, fully 
control inputs, and assert success paths, failure modes, and edge cases. Treat 
failing tests as the primary signal for what to fix next.

No untested code at completion. Do not introduce meaningful new or changed 
logic without unit test coverage. Add tests for all new behavior and for every 
bug fix (including regression tests). If a tiny piece of glue code cannot be 
reasonably unit tested, keep it minimal and explicitly justify it in the README 
(or the PR description if applicable), but do not leave untested core logic. 
Run the full unit test suite at the end and treat uncovered or unexercised new 
logic as a failure of the Definition of Done.

Build tooling and agent workflow. This repository must include a Makefile with 
targets named build, test, lint, and fmt. These targets are the canonical 
workflow for validating changes. make fmt must format the codebase, make lint 
must run static checks, make test must run the full unit test suite, and make 
build must produce the module binary. Before finishing the task, run make fmt, 
make lint, make test, and make build, then fix failures until everything is 
green.

README documentation and structure. Update README.md in the same task. Do not 
include a table of contents. The README must be as short as possible while 
still covering every important detail. Start with end-user content first: 
clearly state what the module is and what it does, then explain how to use it 
with at least one realistic example. After usage, document what files/formats 
it reads and writes at a high level (consistent with BusDK CSV + Table Schema 
conventions), and only then include contributor/developer information: how to 
build, test, lint, and format, and how to run the binary locally. Avoid bullet 
point lists unless the content is genuinely best expressed as a small grouped 
list (for example, a compact set of commands or options); prefer clear 
paragraphs. Include links to https://docs.busdk.com/ for all relevant specs and 
conventions (including any module-specific spec page you discover). If the 
implementation intentionally deviates from the spec, document the deviation 
briefly with rationale and implications.
