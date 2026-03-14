#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "100"

vsix_path="$WS/busdk.language-bus.vsix"
make -C "$ROOT_DIR" package-vscode-extension VSIX_OUT="$vsix_path" >/dev/null
test -s "$vsix_path"

python3 - "$vsix_path" <<'PY'
import json
import sys
import zipfile

vsix_path = sys.argv[1]
required_entries = {
    "[Content_Types].xml",
    "extension.vsixmanifest",
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
PY

echo "e2e OK"
