#!/usr/bin/env python3

import argparse
import json
import os
from pathlib import Path
import zipfile
from xml.sax.saxutils import escape


def repo_root() -> Path:
    return Path(__file__).resolve().parent.parent


def extension_root(root: Path) -> Path:
    return root / "editors" / "vscode-bus-language"


def load_package_manifest(ext_root: Path) -> dict:
    package_path = ext_root / "package.json"
    with package_path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def validate_extension_sources(ext_root: Path, package_manifest: dict) -> None:
    required_paths = [
        ext_root / "package.json",
        ext_root / "README.md",
        ext_root / "LICENSE.md",
        ext_root / "language-configuration.json",
        ext_root / "syntaxes" / "bus.tmLanguage.json",
    ]
    for path in required_paths:
        if not path.is_file():
            raise SystemExit(f"missing required extension file: {path}")
    with (ext_root / "language-configuration.json").open("r", encoding="utf-8") as handle:
        json.load(handle)
    with (ext_root / "syntaxes" / "bus.tmLanguage.json").open("r", encoding="utf-8") as handle:
        json.load(handle)
    languages = package_manifest.get("contributes", {}).get("languages", [])
    if not any(".bus" in language.get("extensions", []) for language in languages):
        raise SystemExit("package.json does not register the .bus extension")


def default_output_path(root: Path, package_manifest: dict) -> Path:
    publisher = package_manifest["publisher"]
    name = package_manifest["name"]
    version = package_manifest["version"]
    return root / "bin" / f"{publisher}.{name}-{version}.vsix"


def build_content_types_xml() -> str:
    return """<?xml version="1.0" encoding="utf-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="json" ContentType="application/json"/>
  <Default Extension="md" ContentType="text/markdown"/>
  <Default Extension="txt" ContentType="text/plain"/>
  <Default Extension="xml" ContentType="text/xml"/>
  <Override PartName="/extension.vsixmanifest" ContentType="text/xml"/>
</Types>
"""


def build_vsix_manifest(package_manifest: dict) -> str:
    identity = f"{package_manifest['publisher']}.{package_manifest['name']}"
    version = escape(package_manifest["version"])
    publisher = escape(package_manifest["publisher"])
    display_name = escape(package_manifest["displayName"])
    description = escape(package_manifest["description"])
    vscode_range = escape(package_manifest["engines"]["vscode"])
    return f"""<?xml version="1.0" encoding="utf-8"?>
<PackageManifest Version="2.0.0" xmlns="http://schemas.microsoft.com/developer/vsx-schema/2011">
  <Metadata>
    <Identity Language="en-US" Id="{escape(identity)}" Version="{version}" Publisher="{publisher}"/>
    <DisplayName>{display_name}</DisplayName>
    <Description xml:space="preserve">{description}</Description>
    <Categories>Programming Languages</Categories>
    <Tags>bus busdk busfile cli</Tags>
  </Metadata>
  <Installation>
    <InstallationTarget Id="Microsoft.VisualStudio.Code" Version="{vscode_range}"/>
  </Installation>
  <Dependencies/>
  <Assets>
    <Asset Type="Microsoft.VisualStudio.Code.Manifest" Path="extension/package.json"/>
    <Asset Type="Microsoft.VisualStudio.Services.Content.Details" Path="extension/README.md"/>
    <Asset Type="Microsoft.VisualStudio.Services.Content.License" Path="extension/LICENSE.md"/>
  </Assets>
</PackageManifest>
"""


def iter_extension_files(ext_root: Path):
    for path in sorted(ext_root.rglob("*")):
        if path.is_file():
            yield path


def write_vsix(output_path: Path, ext_root: Path, package_manifest: dict) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(output_path, "w", compression=zipfile.ZIP_DEFLATED) as archive:
        archive.writestr("[Content_Types].xml", build_content_types_xml())
        archive.writestr("extension.vsixmanifest", build_vsix_manifest(package_manifest))
        for path in iter_extension_files(ext_root):
            relative_path = path.relative_to(ext_root).as_posix()
            archive.write(path, arcname=f"extension/{relative_path}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Package the BusDK VS Code-compatible .bus language extension into a .vsix artifact."
    )
    parser.add_argument(
        "--output",
        help="Write the .vsix artifact to this path. Defaults to ./bin/<publisher>.<name>-<version>.vsix.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    root = repo_root()
    ext_root = extension_root(root)
    package_manifest = load_package_manifest(ext_root)
    validate_extension_sources(ext_root, package_manifest)
    output_path = Path(args.output) if args.output else default_output_path(root, package_manifest)
    if not output_path.is_absolute():
        output_path = root / output_path
    write_vsix(output_path, ext_root, package_manifest)
    print(os.fspath(output_path))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
