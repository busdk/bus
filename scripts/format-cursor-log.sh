#!/usr/bin/env node
/**
 * ndjson-to-text: Stream-friendly version.
 * - Streams NDJSON line-by-line (prints incrementally).
 * - Falls back to buffered JSON/array parse if no NDJSON objects were parsed.
 * - Preserves all CLI flags and formatting behavior.
 */

"use strict";

// -------- CLI parsing --------
const cli = parseCLI(process.argv.slice(2));
if (cli.help) {
  process.stdout.write(helpText());
  process.exit(0);
}

// -------- streaming NDJSON path with buffered fallback --------
const readline = require("readline");

// We'll try to stream NDJSON first. If we never successfully parse a line,
// we fall back to buffered JSON/array parse using the captured raw text.
const rawLines = [];
let sawParsedNdjson = false;

// Maintain a pending merged event to support mergeConsecutive() semantics
// in a streaming-friendly way.
let pending = null;

// Set up line reader
process.stdin.setEncoding("utf8");
const rl = readline.createInterface({
  input: process.stdin,
  crlfDelay: Infinity,
});

// Backpressure-aware write
function writeOut(s) {
  if (!s) return;
  if (!process.stdout.write(s)) {
    rl.pause();
    process.stdout.once("drain", () => rl.resume());
  }
}

rl.on("line", (line) => {
  rawLines.push(line);
  const trimmed = line.trim();
  if (!trimmed) return;

  let obj = null;
  try { obj = JSON.parse(trimmed); } catch { /* not NDJSON; ignore for now */ }

  if (!obj) return;

  sawParsedNdjson = true;

  const norm = normalizeEvent(obj);
  if (!norm) return;

  // Streaming merge logic (replicates mergeConsecutive for assistant/user)
  if (canMerge(pending, norm)) {
    pending.text = (pending.text || "") + (norm.text || "");
  } else {
    flushPending();          // flush previous merged block (after filter)
    pending = { ...norm };   // start a new block
  }
});

rl.on("close", () => {
  if (sawParsedNdjson) {
    // Finish the last block for NDJSON stream
    flushPending(true);
    return;
  }

  // Fallback: buffered parse of the entire input (JSON or NDJSON-ish)
  const input = rawLines.join("\n");
  try {
    const events = parseInput(input);
    const normalized = events.map(normalizeEvent).filter(Boolean);
    const merged = mergeConsecutive(normalized);
    const filtered = merged.filter((b) => filterEvent(b, cli));
    const out = filtered.map((b) => renderPlain(b, cli)).join("\n");
    writeOut(out ? out + "\n" : "");
  } catch (err) {
    writeOut(`(Could not parse input: ${String((err && err.message) || err)})\n`);
  }
});

// ============== Helpers ==============

// Flush the pending merged block if it passes filters.
function flushPending(finalFlush = false) {
  if (!pending) return;
  if (filterEvent(pending, cli)) {
    const s = renderPlain(pending, cli);
    if (s && s.length) writeOut(s + "\n");
  }
  pending = null;
}

// Decide if two events should be merged (assistant/user consecutive text blocks)
function canMerge(a, b) {
  if (!a || !b) return false;
  if (a.html || b.html) return false; // parity with original guard
  const aIsMsg = a.kind === "assistant" || a.kind === "user";
  const bIsMsg = b.kind === "assistant" || b.kind === "user";
  return aIsMsg && bIsMsg && a.role === b.role;
}

// ---- input parsing (JSON or NDJSON) ----
function parseInput(s) {
  const trimmed = s.trim();
  if (!trimmed) return [];
  try {
    const j = JSON.parse(trimmed);
    return Array.isArray(j) ? j : [j];
  } catch {
    const lines = s.split(/\r?\n/).filter((l) => l.trim().length > 0);
    const out = [];
    for (const line of lines) {
      const obj = safeParse(line);
      if (obj) out.push(obj);
    }
    return out;
  }
}
function safeParse(line) { try { return JSON.parse(line); } catch { return null; } }

// ---- normalize to a compact internal shape ----
function concatTextFromContent(content = []) {
  return content
    .filter((c) => c && c.type === "text" && typeof c.text === "string")
    .map((c) => c.text)
    .join("");
}

function normalizeEvent(ev) {
  if (!ev || typeof ev !== "object") return null;

  if (ev.type === "system" && ev.subtype === "init") {
    const parts = [];
    if (ev.apiKeySource) parts.push(`apiKeySource=${ev.apiKeySource}`);
    if (ev.model) parts.push(`model=${ev.model}`);
    if (ev.cwd) parts.push(`cwd=${ev.cwd}`);
    return { role: "system", kind: "system-init", text: `System initialized (${parts.join(", ")})` };
  }

  if (ev.type === "user" && ev.message?.role === "user") {
    return { role: "user", kind: "user", text: concatTextFromContent(ev.message.content) };
  }

  if (ev.type === "assistant" && ev.message?.role === "assistant") {
    return { role: "assistant", kind: "assistant", text: concatTextFromContent(ev.message.content) };
  }

  if (ev.type === "tool_call") {
    const phase = ev.subtype || "event";
    const tc = ev.tool_call || {};
    if (tc.shellToolCall) {
      const t = tc.shellToolCall;
      const args = t.args || {};
      const res = (t.result && (t.result.success || t.result.error)) || null;
      return {
        role: "tool",
        kind: "tool-shell",
        phase,
        cmd: args.command || renderSimpleCommands(args.simpleCommands),
        cwd: args.workingDirectory || "",
        result: res ? summarizeShellResult(res) : null,
      };
    }
    // generic tool call (e.g., updateTodosToolCall)
    const name = Object.keys(tc)[0] || "tool";
    const node = tc[name] || {};
    return {
      role: "tool",
      kind: "tool-generic",
      phase,
      name,
      args: node.args || null,
      result: node.result || null,
    };
  }

  return null;
}

function renderSimpleCommands(list) {
  if (!Array.isArray(list) || list.length === 0) return "";
  return list.map((c) => c.fullText || c.name || "").filter(Boolean).join(" && ");
}

function summarizeShellResult(res) {
  return {
    exitCode: res.exitCode ?? null,
    signal: res.signal || "",
    stdout: typeof res.stdout === "string" ? stripAnsi(res.stdout) : "",
    stderr: typeof res.stderr === "string" ? stripAnsi(res.stderr) : "",
  };
}

// ---- merge consecutive assistant/user fragments (buffered fallback path) ----
function mergeConsecutive(items) {
  const out = [];
  for (const it of items) {
    const last = out[out.length - 1];
    if (
      last &&
      !last.html && !it.html &&
      last.role === it.role &&
      (it.kind === "assistant" || it.kind === "user") &&
      (last.kind === "assistant" || last.kind === "user")
    ) {
      last.text = (last.text || "") + (it.text || "");
    } else {
      out.push({ ...it });
    }
  }
  return out.filter((x) => (typeof x.text === "string" ? x.text.trim().length > 0 : true));
}

// ---- filtering ----
function filterEvent(b, opts) {
  // Role filters
  if (opts.onlyRoles.size && !opts.onlyRoles.has(b.role)) return false;
  if (opts.excludeRoles.has(b.role)) return false;

  // Kind filters
  if (opts.onlyKinds.size && !opts.onlyKinds.has(b.kind)) return false;
  if (opts.excludeKinds.has(b.kind)) return false;

  // Regex include/exclude
  const hay = eventTextForMatch(b);
  if (opts.includeRe && !opts.includeRe.test(hay)) return false;
  if (opts.excludeRe && opts.excludeRe.test(hay)) return false;

  return true;
}

function eventTextForMatch(b) {
  if (b.role === "user" || b.role === "assistant" || b.role === "system") {
    return (b.text || "").toString();
  }
  if (b.kind === "tool-shell") {
    return [
      b.phase || "",
      b.cmd || "",
      b.cwd || "",
      b.result?.stdout || "",
      b.result?.stderr || "",
    ].join("\n");
  }
  if (b.kind === "tool-generic") {
    return [
      b.phase || "",
      b.name || "",
      safePretty(b.args) || "",
      safePretty(b.result) || "",
    ].join("\n");
  }
  return "";
}

// ---- rendering (plain text) ----
function renderPlain(b, opts) {
  switch (b.role) {
    case "system":
      return `[System] ${b.text || ""}`.trim();

    case "user":
      return formatBlock("User", b.text);

    case "assistant":
      return formatBlock("Assistant", b.text);

    case "tool": {
      if (b.kind === "tool-shell") {
        const lines = [];
        lines.push(`[Tool] Shell ${b.phase}`);
        if (b.cmd) lines.push(`Command: ${b.cmd}`);
        if (b.cwd) lines.push(`Directory: ${b.cwd}`);
        if (b.result) {
          const meta = [
            `exitCode=${b.result.exitCode === null ? "?" : String(b.result.exitCode)}`,
            b.result.signal ? `signal=${b.result.signal}` : null,
          ].filter(Boolean).join(", ");
          lines.push(`Result: ${meta || "(no metadata)"}`);

          if (!opts.toolsMetaOnly && opts.showStdout) {
            if (b.result.stdout && b.result.stdout.trim()) {
              lines.push(`stdout:\n${indent(b.result.stdout.trim(), "  ")}`);
            } else if (opts.showEmptyIO) {
              lines.push(`stdout: (empty)`);
            }
          }
          if (!opts.toolsMetaOnly && opts.showStderr && b.result.stderr && b.result.stderr.trim()) {
            lines.push(`stderr:\n${indent(b.result.stderr.trim(), "  ")}`);
          }
        }
        return lines.join("\n");
      }
      // generic tool call
      const lines = [];
      lines.push(`[Tool] ${b.name || "tool"} ${b.phase}`);
      if (!opts.toolsMetaOnly) {
        if (b.args && Object.keys(b.args).length) {
          lines.push(`args:\n${indent(safePretty(b.args), "  ")}`);
        }
        if (b.result && Object.keys(b.result).length) {
          lines.push(`result:\n${indent(safePretty(b.result), "  ")}`);
        }
      }
      return lines.join("\n");
    }

    default:
      return `(unknown event)`;
  }
}

function formatBlock(label, body) {
  if (body == null || String(body).trim() === "") return `[${label}]`;
  const s = String(body);
  if (!s.includes("\n")) return `[${label}] ${s}`;
  return `[${label}]\n` + indent(s.trimEnd(), "  ");
}

// ---- misc utils ----
function indent(s, pad = "  ") {
  return String(s).split("\n").map((l) => pad + l).join("\n");
}
function safePretty(obj) { try { return JSON.stringify(obj, null, 2); } catch { return String(obj); } }
function stripAnsi(s) { return String(s).replace(/\u001b\[[0-9;]*m/g, ""); }

// ---- CLI parsing & help ----
function parseCLI(argv) {
  const opts = {
    help: false,
    onlyRoles: new Set(),     // e.g. system,user,assistant,tool
    excludeRoles: new Set(),
    onlyKinds: new Set(),     // e.g. system-init,assistant,user,tool-shell,tool-generic
    excludeKinds: new Set(),
    includeRe: null,          // RegExp to include
    excludeRe: null,          // RegExp to exclude
    showStdout: true,
    showStderr: true,
    toolsMetaOnly: false,
    showEmptyIO: false,
  };

  for (const a of argv) {
    if (a === "-h" || a === "--help") { opts.help = true; continue; }
    if (a === "--no-tools") { opts.excludeRoles.add("tool"); continue; }
    if (a === "--no-system") { opts.excludeRoles.add("system"); continue; }
    if (a === "--no-user") { opts.excludeRoles.add("user"); continue; }
    if (a === "--no-assistant") { opts.excludeRoles.add("assistant"); continue; }

    if (a.startsWith("--only-roles=")) addCsv(opts.onlyRoles, a.split("=")[1]);
    else if (a.startsWith("--exclude-roles=")) addCsv(opts.excludeRoles, a.split("=")[1]);
    else if (a.startsWith("--only-kinds=")) addCsv(opts.onlyKinds, a.split("=")[1]);
    else if (a.startsWith("--exclude-kinds=")) addCsv(opts.excludeKinds, a.split("=")[1]);

    else if (a.startsWith("--match=")) opts.includeRe = parseRegexArg(a.split("=")[1]);
    else if (a.startsWith("--not-match=")) opts.excludeRe = parseRegexArg(a.split("=")[1]);

    else if (a === "--no-stdout") opts.showStdout = false;
    else if (a === "--no-stderr") opts.showStderr = false;
    else if (a === "--tools-meta-only") opts.toolsMetaOnly = true;
    else if (a === "--show-empty-io") opts.showEmptyIO = true;
    else {
      if (!opts._unknown) opts._unknown = [];
      opts._unknown.push(a);
    }
  }
  return opts;
}

function addCsv(set, csv) {
  for (const x of String(csv).split(",").map((s) => s.trim()).filter(Boolean)) set.add(x);
}

function parseRegexArg(val) {
  // Accept "/expr/flags" or plain text (treated as case-insensitive substring).
  const s = String(val);
  const m = /^\/(.+)\/([a-z]*)$/.exec(s);
  if (m) {
    try { return new RegExp(m[1], m[2]); } catch { /* fall through */ }
  }
  return new RegExp(escapeRegex(s), "i");
}
function escapeRegex(s) { return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"); }

function helpText() {
  return `
ndjson-to-text — Convert JSON/NDJSON to readable plain text with filters (streaming).
- NDJSON is processed incrementally; output is printed as events arrive.
- If no NDJSON objects are parsed, falls back to buffered JSON/array parsing.

USAGE
  producer | ndjson-to-text [options]
  ndjson-to-text [options] < events.ndjson

FILTER OPTIONS
  --no-tools                Exclude all tool events (role=tool)
  --no-system               Exclude system events
  --no-user                 Exclude user messages
  --no-assistant            Exclude assistant messages
  --only-roles=A,B          Only include these roles (system,user,assistant,tool)
  --exclude-roles=A,B       Exclude these roles
  --only-kinds=A,B          Only include these kinds (e.g. system-init,assistant,user,tool-shell,tool-generic)
  --exclude-kinds=A,B       Exclude these kinds
  --match=REGEX             Include only events whose content matches REGEX.
                            Accepts "/expr/flags" or plain text (case-insensitive).
  --not-match=REGEX         Exclude events whose content matches REGEX.

TOOL OUTPUT TUNING
  --no-stdout               Hide tool stdout
  --no-stderr               Hide tool stderr
  --tools-meta-only         Show tool headers/meta, hide args/result/stdout/stderr
  --show-empty-io           When stdout is empty, print "stdout: (empty)"

MISC
  -h, --help                Show this help

EXAMPLES
  ndjson-to-text --no-tools < events.ndjson
  ndjson-to-text --only-roles=assistant,user < events.ndjson
  ndjson-to-text --match="/pull request/i" --no-tools < events.ndjson
  ndjson-to-text --only-kinds=tool-shell --no-stdout < events.ndjson

`;
}
