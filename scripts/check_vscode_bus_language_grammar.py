#!/usr/bin/env python3
import json
import pathlib
import re
import sys


def fail(message: str) -> None:
    raise SystemExit(message)


def expect_match(label: str, pattern: str, text: str) -> None:
    if re.search(pattern, text) is None:
        fail(f"{label} did not match sample: {text!r}")


def main() -> None:
    if len(sys.argv) != 2:
        fail("usage: check_vscode_bus_language_grammar.py <grammar.json>")

    grammar_path = pathlib.Path(sys.argv[1])
    grammar = json.loads(grammar_path.read_text(encoding="utf-8"))
    if grammar.get("scopeName") != "source.bus":
        fail("unexpected grammar scopeName")

    repository = grammar.get("repository", {})
    required = {
        "shebang",
        "comment",
        "include",
        "directiveLine",
        "commandTarget",
        "subcommand",
        "longFlag",
        "shortFlag",
        "assignment",
        "doubleQuotedString",
        "singleQuotedString",
        "datetime",
        "number",
        "continuation",
    }
    missing = sorted(required - set(repository))
    if missing:
        fail(f"missing grammar repository entries: {missing}")

    expect_match("shebang", repository["shebang"]["match"], "#!/usr/bin/env bus")
    expect_match("comment", repository["comment"]["match"], "  # month-end close")
    expect_match("include", repository["include"]["match"], "2026-03.bus")
    expect_match("directiveLine", repository["directiveLine"]["begin"], "  --chdir data --color auto")
    expect_match("commandTarget", repository["commandTarget"]["match"], "journal add --date 2026-03-14")
    expect_match("subcommand", repository["subcommand"]["match"], "journal add --date 2026-03-14")
    expect_match("longFlag", repository["longFlag"]["match"], "--source-id close:2026-03:1")
    expect_match("shortFlag", repository["shortFlag"]["match"], "-C ./workspace")
    expect_match("assignment", repository["assignment"]["match"], "bank_txn_id=import-2026-03-0001")
    expect_match("assignment", repository["assignment"]["match"], "1910=100.00")
    expect_match("doubleQuotedString", repository["doubleQuotedString"]["begin"], '"Kuukausikatkaisu: siirto"')
    expect_match("doubleQuotedString", repository["doubleQuotedString"]["end"], '"Kuukausikatkaisu: siirto"')
    expect_match("singleQuotedString", repository["singleQuotedString"]["begin"], "'Example Vendor'")
    expect_match("singleQuotedString", repository["singleQuotedString"]["end"], "'Example Vendor'")
    expect_match("datetime", repository["datetime"]["match"], "2026-03-14")
    expect_match("datetime", repository["datetime"]["match"], "2026-03-14T12:30:45Z")
    expect_match("number", repository["number"]["match"], "-861.68")
    expect_match("continuation", repository["continuation"]["match"], "journal add \\")


if __name__ == "__main__":
    main()
