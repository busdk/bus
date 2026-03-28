"use strict";

const fs = require("fs");
const core = require("./bus_language_core");

const documents = new Map();
let shutdownRequested = false;
let buffer = Buffer.alloc(0);

// writeMessage emits one JSON-RPC message with LSP Content-Length framing.
// Used by: respond, respondError.
function writeMessage(payload) {
  const body = Buffer.from(JSON.stringify(payload), "utf8");
  const header = Buffer.from(`Content-Length: ${body.length}\r\n\r\n`, "utf8");
  process.stdout.write(Buffer.concat([header, body]));
}

// respond emits one successful JSON-RPC response.
// Used by: handleRequest.
function respond(id, result) {
  writeMessage({ jsonrpc: "2.0", id, result });
}

// respondError emits one JSON-RPC error response for unsupported requests.
// Used by: handleRequest.
function respondError(id, code, message) {
  writeMessage({ jsonrpc: "2.0", id, error: { code, message } });
}

// handleRequest serves supported LSP requests for initialize, shutdown, and semantic tokens.
// Used by: handleMessage.
function handleRequest(message) {
  const { id, method, params } = message;
  switch (method) {
    case "initialize":
      respond(id, {
        capabilities: {
          textDocumentSync: 1,
          semanticTokensProvider: {
            legend: core.legend,
            full: true,
          },
        },
        serverInfo: {
          name: "bus-language-server",
          version: "0.1.0",
        },
      });
      return;
    case "shutdown":
      shutdownRequested = true;
      respond(id, null);
      return;
    case "textDocument/semanticTokens/full": {
      const uri = params && params.textDocument ? params.textDocument.uri : "";
      respond(id, { data: core.encodeSemanticTokens(documents.get(uri) || "") });
      return;
    }
    default:
      respondError(id, -32601, `method not found: ${method}`);
  }
}

// handleNotification updates document state and lifecycle notifications.
// Used by: handleMessage.
function handleNotification(message) {
  const { method, params } = message;
  switch (method) {
    case "initialized":
      return;
    case "textDocument/didOpen":
      documents.set(params.textDocument.uri, params.textDocument.text || "");
      return;
    case "textDocument/didChange":
      if (params.contentChanges && params.contentChanges.length > 0) {
        documents.set(params.textDocument.uri, params.contentChanges[params.contentChanges.length - 1].text || "");
      }
      return;
    case "textDocument/didClose":
      documents.delete(params.textDocument.uri);
      return;
    case "exit":
      process.exit(shutdownRequested ? 0 : 1);
      return;
    default:
      return;
  }
}

// handleMessage routes one decoded JSON-RPC message to request or notification handling.
// Used by: drainBuffer.
function handleMessage(message) {
  if (Object.prototype.hasOwnProperty.call(message, "id")) {
    handleRequest(message);
    return;
  }
  handleNotification(message);
}

// drainBuffer decodes as many framed JSON-RPC messages as are currently buffered.
// Used by: runStdio.
function drainBuffer() {
  while (true) {
    const delimiter = buffer.indexOf("\r\n\r\n");
    if (delimiter < 0) {
      return;
    }
    const header = buffer.slice(0, delimiter).toString("utf8");
    const lengthMatch = header.match(/Content-Length:\s*(\d+)/i);
    if (!lengthMatch) {
      throw new Error("missing Content-Length header");
    }
    const contentLength = parseInt(lengthMatch[1], 10);
    const messageStart = delimiter + 4;
    if (buffer.length < messageStart + contentLength) {
      return;
    }
    const body = buffer.slice(messageStart, messageStart + contentLength).toString("utf8");
    buffer = buffer.slice(messageStart + contentLength);
    handleMessage(JSON.parse(body));
  }
}

// runStdio starts the LSP stdio loop for editor clients.
// Used by: main when --stdio is selected.
function runStdio() {
  process.stdin.on("data", (chunk) => {
    buffer = Buffer.concat([buffer, chunk]);
    drainBuffer();
  });
  process.stdin.on("end", () => {
    if (!shutdownRequested) {
      process.exit(1);
    }
  });
}

// printTokens writes semantic token output for one file in offline inspection mode.
// Used by: main when --semantic-tokens is selected.
function printTokens(path) {
  const text = fs.readFileSync(path, "utf8");
  process.stdout.write(`${JSON.stringify({
    legend: core.legend,
    tokens: core.semanticTokens(text),
    data: core.encodeSemanticTokens(text),
  }, null, 2)}\n`);
}

// main selects stdio LSP mode or offline token-dump mode.
// Used by: process entrypoint.
function main(argv) {
  if (argv.length === 2 && argv[0] === "--semantic-tokens") {
    printTokens(argv[1]);
    return;
  }
  if (argv.length === 1 && argv[0] === "--stdio") {
    runStdio();
    return;
  }
  process.stderr.write("usage: language-server.js --stdio | --semantic-tokens <file>\n");
  process.exit(2);
}

main(process.argv.slice(2));
