"use strict";

const legend = Object.freeze({
  tokenTypes: ["comment", "function", "keyword", "property", "operator", "string", "number"],
  tokenModifiers: [],
});

const tokenTypeSet = new Set(legend.tokenTypes);
const commandWord = /^[a-z][a-z0-9-]*$/;
const longFlag = /^--[A-Za-z0-9][A-Za-z0-9-]*/;
const shortFlag = /^-[A-Za-z0-9]+$/;
const assignmentKey = /^[A-Za-z0-9_][A-Za-z0-9_.-]*$/;
const dateLike = /^\d{4}-\d{2}-\d{2}(?:T\d{2}:\d{2}:\d{2}Z)?$/;
const numberLike = /^-?\d+(?:\.\d+)?$/;
const includeLike = /^(?:(?:\.{1,2}\/|\/)?(?:[A-Za-z0-9_.-]+\/)*[A-Za-z0-9_.-]+\.bus)$/;

// pushToken appends one semantic token when the token type is valid and the span is non-empty.
// Used by: appendLineTokens, classifyValue.
function pushToken(tokens, line, start, length, type) {
  if (length <= 0 || !tokenTypeSet.has(type)) {
    return;
  }
  tokens.push({ line, start, length, type });
}

// tokenizeLine splits one busfile line into whitespace-delimited or quoted token spans with offsets.
// Used by: appendLineTokens.
function tokenizeLine(text) {
  const tokens = [];
  let index = 0;
  while (index < text.length) {
    while (index < text.length && /\s/.test(text[index])) {
      index += 1;
    }
    if (index >= text.length) {
      break;
    }
    const start = index;
    if (text[index] === '"' || text[index] === "'") {
      const quote = text[index];
      index += 1;
      while (index < text.length) {
        if (text[index] === "\\" && quote === '"' && index + 1 < text.length) {
          index += 2;
          continue;
        }
        if (text[index] === quote) {
          index += 1;
          break;
        }
        index += 1;
      }
      tokens.push({ text: text.slice(start, index), start, end: index, quoted: true });
      continue;
    }
    while (index < text.length && !/\s/.test(text[index])) {
      if (text[index] === "\\" && index + 1 < text.length) {
        index += 2;
        continue;
      }
      index += 1;
    }
    tokens.push({ text: text.slice(start, index), start, end: index, quoted: false });
  }
  return tokens;
}

// classifyValue appends semantic tokens for an assignment or flag value segment.
// Used by: appendLineTokens.
function classifyValue(tokens, line, token, offset) {
  const value = token.text.slice(offset);
  const start = token.start + offset;
  if (value.length === 0) {
    return;
  }
  if (value.startsWith('"') || value.startsWith("'")) {
    pushToken(tokens, line, start, value.length, "string");
    return;
  }
  if (dateLike.test(value) || numberLike.test(value)) {
    pushToken(tokens, line, start, value.length, "number");
  }
}

// appendLineTokens classifies one busfile line into semantic token spans.
// Used by: semanticTokens.
function appendLineTokens(out, line, text) {
  const trimmed = text.trim();
  if (trimmed === "") {
    return;
  }
  if (trimmed.startsWith("#!")) {
    pushToken(out, line, 0, text.length, "comment");
    return;
  }
  if (/^\s*#/.test(text)) {
    pushToken(out, line, 0, text.length, "comment");
    return;
  }

  const continuationMatch = text.match(/\\\s*$/);
  if (continuationMatch) {
    pushToken(out, line, continuationMatch.index, 1, "operator");
  }

  const parts = tokenizeLine(text);
  if (parts.length === 0) {
    return;
  }
  const directive = parts[0].text.startsWith("-");
  const includeLine = parts.length === 1 && includeLike.test(parts[0].text);

  for (let index = 0; index < parts.length; index += 1) {
    const part = parts[index];
    if (part.quoted) {
      pushToken(out, line, part.start, part.end - part.start, "string");
      continue;
    }

    const raw = part.text;
    if (includeLine) {
      pushToken(out, line, part.start, part.end - part.start, "string");
      continue;
    }
    if (longFlag.test(raw)) {
      const match = raw.match(longFlag);
      pushToken(out, line, part.start, match[0].length, "keyword");
      if (raw.length > match[0].length && raw[match[0].length] === "=") {
        pushToken(out, line, part.start + match[0].length, 1, "operator");
        classifyValue(out, line, part, match[0].length + 1);
      }
      continue;
    }
    if (shortFlag.test(raw)) {
      pushToken(out, line, part.start, part.end - part.start, "keyword");
      continue;
    }
    const eq = raw.indexOf("=");
    if (eq > 0) {
      const key = raw.slice(0, eq);
      if (assignmentKey.test(key)) {
        pushToken(out, line, part.start, eq, "property");
        pushToken(out, line, part.start + eq, 1, "operator");
        classifyValue(out, line, part, eq + 1);
        continue;
      }
    }
    if (!directive && index < 2 && commandWord.test(raw)) {
      pushToken(out, line, part.start, part.end - part.start, "function");
      continue;
    }
    if (dateLike.test(raw) || numberLike.test(raw)) {
      pushToken(out, line, part.start, part.end - part.start, "number");
    }
  }
}

// semanticTokens returns absolute semantic token spans for one full document.
// Used by: VS Code extension provider, language server, semantic token checks.
function semanticTokens(text) {
  const out = [];
  const lines = text.split(/\r?\n/);
  for (let line = 0; line < lines.length; line += 1) {
    appendLineTokens(out, line, lines[line]);
  }
  out.sort((left, right) => {
    if (left.line !== right.line) {
      return left.line - right.line;
    }
    if (left.start !== right.start) {
      return left.start - right.start;
    }
    return left.length - right.length;
  });
  return out;
}

// encodeSemanticTokens converts absolute token spans into LSP delta-encoded integers.
// Used by: language server semanticTokens/full responses and semantic token checks.
function encodeSemanticTokens(text) {
  const tokens = semanticTokens(text);
  const data = [];
  let previousLine = 0;
  let previousStart = 0;
  for (const token of tokens) {
    const deltaLine = token.line - previousLine;
    const deltaStart = deltaLine === 0 ? token.start - previousStart : token.start;
    data.push(deltaLine, deltaStart, token.length, legend.tokenTypes.indexOf(token.type), 0);
    previousLine = token.line;
    previousStart = token.start;
  }
  return data;
}

module.exports = {
  legend,
  semanticTokens,
  encodeSemanticTokens,
};
