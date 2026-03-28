"use strict";

const defineGrammar = globalThis.grammar || ((spec) => spec);

module.exports = defineGrammar({
  name: "bus",
  extras: () => [/[ \t\f]/],
  rules: {
    source_file: ($) => repeat(choice($.line, $.blank_line)),
    blank_line: () => /\r?\n/,
    line: ($) =>
      seq(
        choice($.shebang, $.comment, $.include, $.directive_line, $.command_line),
        optional($.continuation),
        /\r?\n/,
      ),
    shebang: () => token(seq("#!", /.*/)),
    comment: () => token(seq(optional(/[ \t]+/), "#", /.*/)),
    include: ($) => field("path", $.include_path),
    include_path: () => /(?:(?:\.{1,2}\/|\/)?(?:[A-Za-z0-9_.-]+\/)*[A-Za-z0-9_.-]+\.bus)/,
    directive_line: ($) => repeat1(choice($.long_flag, $.short_flag, $.assignment, $.string, $.date, $.number, $.word)),
    command_line: ($) =>
      seq(
        field("command", $.command_name),
        optional(field("subcommand", $.subcommand_name)),
        repeat(choice($.long_flag, $.short_flag, $.assignment, $.string, $.date, $.number, $.word)),
      ),
    command_name: () => /[a-z][a-z0-9-]*/,
    subcommand_name: () => /[a-z][a-z0-9-]*/,
    long_flag: () => /--[A-Za-z0-9][A-Za-z0-9-]*(?:=[^\s"'\\]+)?/,
    short_flag: () => /-[A-Za-z0-9]+/,
    assignment: ($) => seq(field("key", $.assignment_key), "=", field("value", choice($.string, $.date, $.number, $.word))),
    assignment_key: () => /[A-Za-z0-9_][A-Za-z0-9_.-]*/,
    string: () => choice(/"([^"\\]|\\.)*"/, /'[^']*'/),
    date: () => /\d{4}-\d{2}-\d{2}(?:T\d{2}:\d{2}:\d{2}Z)?/,
    number: () => /-?\d+(?:\.\d+)?/,
    word: () => /[^\s"'\\=]+/,
    continuation: () => token("\\"),
  },
});
