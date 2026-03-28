#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"
e2e_setup "120"

REPORT_PATH="$WS/release-surface.json"
python3 "$ROOT_DIR/scripts/check_vscode_bus_language_release.py" --format json > "$REPORT_PATH"

python3 - "$REPORT_PATH" <<'PY'
import json
import sys

report_path = sys.argv[1]
with open(report_path, "r", encoding="utf-8") as handle:
    report = json.load(handle)

if report["extension_id"] != "busdk.language-bus":
    raise SystemExit(f"unexpected extension_id: {report['extension_id']}")
if report["openvsx_slug"] != "busdk/language-bus":
    raise SystemExit(f"unexpected openvsx_slug: {report['openvsx_slug']}")
if not report["vsix_path"].startswith("bin/"):
    raise SystemExit(f"unexpected vsix_path: {report['vsix_path']}")
if not report["vsix_release_asset"].endswith(".vsix"):
    raise SystemExit(f"unexpected vsix_release_asset: {report['vsix_release_asset']}")
if "vscode" not in report["supported_editors"] or "vscodium" not in report["supported_editors"]:
    raise SystemExit(f"unexpected supported_editors: {report['supported_editors']}")
PY

echo "e2e OK"
