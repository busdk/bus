# tree-sitter-bus

This directory contains the Tree-sitter source grammar for `.bus` files.

It is the parser-backed editor tier for BusDK command files outside the shipped
VS Code-compatible extension. The grammar and highlight query are intended for
editors that consume Tree-sitter grammars directly, such as Neovim and Emacs.

Files:

- `grammar.js` defines the canonical `.bus` parser grammar contract.
- `queries/highlights.scm` maps parser nodes to standard highlight captures.

Typical local integration flow:

```sh
tree-sitter generate
tree-sitter test
```

Then point your editor's Tree-sitter runtime at this grammar and query
directory.
