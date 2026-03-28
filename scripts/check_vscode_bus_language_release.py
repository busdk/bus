#!/usr/bin/env python3

import argparse
import json
import re
from pathlib import Path


SEMVER_RE = re.compile(r"^\d+\.\d+\.\d+$")
NAME_RE = re.compile(r"^[a-z0-9][a-z0-9-]*$")


def repo_root() -> Path:
    return Path(__file__).resolve().parent.parent


def extension_root(root: Path) -> Path:
    return root / "editors" / "vscode-bus-language"


def load_package_manifest(ext_root: Path) -> dict:
    with (ext_root / "package.json").open("r", encoding="utf-8") as handle:
        return json.load(handle)


def validate_release_manifest(package_manifest: dict) -> None:
    publisher = package_manifest.get("publisher", "")
    name = package_manifest.get("name", "")
    version = package_manifest.get("version", "")
    display_name = package_manifest.get("displayName", "")
    description = package_manifest.get("description", "")
    repository = package_manifest.get("repository", {}).get("url", "")
    homepage = package_manifest.get("homepage", "")
    bugs_url = package_manifest.get("bugs", {}).get("url", "")
    vscode_engine = package_manifest.get("engines", {}).get("vscode", "")
    categories = package_manifest.get("categories", [])
    contributes = package_manifest.get("contributes", {})
    languages = contributes.get("languages", [])

    if not NAME_RE.match(publisher):
        raise SystemExit(f"invalid publisher for release surface: {publisher!r}")
    if not NAME_RE.match(name):
        raise SystemExit(f"invalid extension name for release surface: {name!r}")
    if not SEMVER_RE.match(version):
        raise SystemExit(f"invalid semver version for release surface: {version!r}")
    if not display_name or not description:
        raise SystemExit("displayName and description are required for release surfaces")
    if not repository or not homepage or not bugs_url:
        raise SystemExit("repository, homepage, and bugs.url are required for release surfaces")
    if not vscode_engine:
        raise SystemExit("engines.vscode is required for release surfaces")
    if "Programming Languages" not in categories:
        raise SystemExit("categories must include Programming Languages")
    if not any(language.get("id") == "bus" and ".bus" in language.get("extensions", []) for language in languages):
        raise SystemExit("package.json does not register the .bus extension")


def release_surface(root: Path, package_manifest: dict) -> dict:
    publisher = package_manifest["publisher"]
    name = package_manifest["name"]
    version = package_manifest["version"]
    vsix_name = f"{publisher}.{name}-{version}.vsix"
    return {
        "extension_id": f"{publisher}.{name}",
        "openvsx_namespace": publisher,
        "openvsx_name": name,
        "openvsx_slug": f"{publisher}/{name}",
        "version": version,
        "vsix_path": str((root / "bin" / vsix_name).relative_to(root)),
        "vsix_release_asset": vsix_name,
        "supported_editors": ["vscode", "cursor", "vscodium", "windsurf"],
    }


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Validate and print the supported release surfaces for the BusDK VS Code-compatible .bus extension."
    )
    parser.add_argument(
        "--format",
        choices=("text", "json"),
        default="text",
        help="Render the release surface report as text or JSON.",
    )
    return parser.parse_args()


def render_text(surface: dict) -> str:
    lines = [
        f"extension_id {surface['extension_id']}",
        f"version {surface['version']}",
        f"vsix_path {surface['vsix_path']}",
        f"vsix_release_asset {surface['vsix_release_asset']}",
        f"openvsx_slug {surface['openvsx_slug']}",
        f"supported_editors {' '.join(surface['supported_editors'])}",
    ]
    return "\n".join(lines) + "\n"


def main() -> int:
    args = parse_args()
    root = repo_root()
    manifest = load_package_manifest(extension_root(root))
    validate_release_manifest(manifest)
    surface = release_surface(root, manifest)
    if args.format == "json":
        print(json.dumps(surface, indent=2))
    else:
        print(render_text(surface), end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
