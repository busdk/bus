"use strict";

const fs = require("fs");
const path = require("path");

function fail(message) {
  process.stderr.write(`${message}\n`);
  process.exit(1);
}

function main() {
  const root = path.resolve(__dirname, "..");
  const grammarPath = path.join(root, "editors", "tree-sitter-bus", "grammar.js");
  const queryPath = path.join(root, "editors", "tree-sitter-bus", "queries", "highlights.scm");
  const readmePath = path.join(root, "editors", "tree-sitter-bus", "README.md");

  const grammar = require(grammarPath);
  if (!grammar || grammar.name !== "bus") {
    fail("unexpected tree-sitter grammar name");
  }
  const requiredRules = [
    "source_file",
    "line",
    "shebang",
    "comment",
    "include",
    "directive_line",
    "command_line",
    "command_name",
    "subcommand_name",
    "long_flag",
    "short_flag",
    "assignment",
    "string",
    "date",
    "number",
    "continuation",
  ];
  for (const rule of requiredRules) {
    if (!grammar.rules || typeof grammar.rules[rule] !== "function") {
      fail(`missing tree-sitter rule: ${rule}`);
    }
  }

  const query = fs.readFileSync(queryPath, "utf8");
  for (const capture of ["@comment", "@function", "@keyword", "@property", "@operator", "@string", "@number"]) {
    if (!query.includes(capture)) {
      fail(`missing highlight capture ${capture}`);
    }
  }

  const readme = fs.readFileSync(readmePath, "utf8");
  if (!readme.includes("Neovim") || !readme.includes("Emacs")) {
    fail("tree-sitter README must mention Neovim and Emacs integration");
  }
}

main();
