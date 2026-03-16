# Bus Language Support

This extension adds syntax highlighting and `.bus` file associations for BusDK
command files.

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

## Installation

Install the published package from the marketplace or install a `.vsix`
artifact from a BusDK release.

If you are building from source, package the extension from the `bus`
repository:

```sh
make package-vscode-extension
```

That command writes a `.vsix` file into `./bin/`.
