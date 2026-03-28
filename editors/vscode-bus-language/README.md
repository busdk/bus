# Bus Language Support

This extension adds syntax highlighting, semantic tokens, and `.bus` file
associations for BusDK command files.

It is intended for VS Code compatible editors such as VS Code, Cursor,
VSCodium, and Windsurf.

## What it highlights

- shebang lines such as `#!/usr/bin/env bus`
- comment lines beginning with `#`
- `.bus` include lines
- sticky directive lines that begin with dispatcher-global flags
- command targets and first subcommands
- long and short flags
- `key=value` assignments such as `--set field=value` or `1910=100.00`
- quoted strings
- ISO-style dates and datetimes
- trailing line-continuation backslashes

Semantic token support is provided from the same local parser contract and
surfaces standard token classes for commands, flags, assignments, strings,
dates, and numbers.

## Installation

Install the published package from the marketplace or install a `.vsix`
artifact from a BusDK release.

If you are building from source, package the extension from the `bus`
repository:

```sh
make package-vscode-extension
```

That command writes a `.vsix` file into `./bin/`.

Supported release surfaces are:

- `.vsix` release asset for VS Code, Cursor, Windsurf, and other editors that
  accept local VS Code extension packages
- Open VSX-compatible metadata under the extension id `busdk.language-bus`,
  with namespace/name `busdk/language-bus`

Maintainers can validate that release metadata and packaging still match those
surfaces from the `bus` repository root:

```sh
make check-vscode-extension-release
```

## Language server

The extension directory also ships a lightweight stdio language server:

```sh
node editors/vscode-bus-language/language-server.js --stdio
```

Editors that can start an arbitrary stdio language server can use that script
directly for semantic-token support without depending on the VS Code extension
host.
