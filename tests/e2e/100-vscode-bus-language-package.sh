#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "100"

vsix_path="$WS/busdk.language-bus.vsix"
make -C "$ROOT_DIR" package-vscode-extension VSIX_OUT="$vsix_path" >/dev/null
test -s "$vsix_path"

python3 - "$vsix_path" <<'PY'
import json
import pathlib
import sys
import tempfile
import zipfile
import subprocess

vsix_path = sys.argv[1]
required_entries = {
    "[Content_Types].xml",
    "extension.vsixmanifest",
    "extension/bus_language_core.js",
    "extension/extension.js",
    "extension/language-server.js",
    "extension/package.json",
    "extension/README.md",
    "extension/LICENSE.md",
    "extension/language-configuration.json",
    "extension/syntaxes/bus.tmLanguage.json",
}
with zipfile.ZipFile(vsix_path) as archive:
    names = set(archive.namelist())
    missing = sorted(required_entries - names)
    if missing:
        raise SystemExit(f"missing entries in VSIX: {missing}")
    manifest = json.loads(archive.read("extension/package.json"))
    languages = manifest["contributes"]["languages"]
    if not any(lang["id"] == "bus" and ".bus" in lang["extensions"] for lang in languages):
        raise SystemExit("extension package.json does not register .bus")
    grammar = json.loads(archive.read("extension/syntaxes/bus.tmLanguage.json"))
    if grammar.get("scopeName") != "source.bus":
        raise SystemExit("unexpected grammar scopeName")
    with tempfile.TemporaryDirectory() as tmp:
        grammar_path = pathlib.Path(tmp) / "bus.tmLanguage.json"
        grammar_path.write_text(json.dumps(grammar), encoding="utf-8")
        subprocess.run(
            ["python3", str(pathlib.Path("scripts/check_vscode_bus_language_grammar.py")), str(grammar_path)],
            check=True,
        )
PY

echo "e2e OK"
