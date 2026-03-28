#!/usr/bin/env python3

import json
import subprocess
import sys
from pathlib import Path


def send_message(handle, payload):
    body = json.dumps(payload).encode("utf-8")
    header = f"Content-Length: {len(body)}\r\n\r\n".encode("utf-8")
    handle.write(header)
    handle.write(body)
    handle.flush()


def read_message(handle):
    headers = {}
    while True:
        line = handle.readline()
        if not line:
            raise SystemExit("language server closed stdout unexpectedly")
        if line == b"\r\n":
            break
        name, value = line.decode("utf-8").split(":", 1)
        headers[name.strip().lower()] = value.strip()
    content_length = int(headers["content-length"])
    body = handle.read(content_length)
    return json.loads(body.decode("utf-8"))


def decode_tokens(data, legend):
    token_types = legend["tokenTypes"]
    out = []
    line = 0
    start = 0
    for index in range(0, len(data), 5):
        delta_line, delta_start, length, token_type, _mods = data[index:index + 5]
        if delta_line != 0:
            line += delta_line
            start = delta_start
        else:
            start += delta_start
        out.append({"line": line, "start": start, "length": length, "type": token_types[token_type]})
    return out


def main():
    root = Path(__file__).resolve().parent.parent
    server = root / "editors" / "vscode-bus-language" / "language-server.js"
    proc = subprocess.Popen(
        ["node", str(server), "--stdio"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    try:
        send_message(proc.stdin, {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {"capabilities": {}},
        })
        initialize = read_message(proc.stdout)
        legend = initialize["result"]["capabilities"]["semanticTokensProvider"]["legend"]
        if "function" not in legend["tokenTypes"] or "keyword" not in legend["tokenTypes"]:
            raise SystemExit("unexpected semantic token legend")

        text = "\n".join([
            "#!/usr/bin/env bus",
            "# month end",
            "--chdir data --color auto",
            "journal add --date 2026-03-14 --desc \"Month end\" --debit 1910=100.00",
            "bank add --set bank_txn_id=import-2026-03-0001",
            "",
        ])
        uri = "file:///tmp/sample.bus"
        send_message(proc.stdin, {
            "jsonrpc": "2.0",
            "method": "textDocument/didOpen",
            "params": {
                "textDocument": {
                    "uri": uri,
                    "languageId": "bus",
                    "version": 1,
                    "text": text,
                }
            },
        })
        send_message(proc.stdin, {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "textDocument/semanticTokens/full",
            "params": {"textDocument": {"uri": uri}},
        })
        tokens_response = read_message(proc.stdout)
        tokens = decode_tokens(tokens_response["result"]["data"], legend)
        present = {token["type"] for token in tokens}
        expected = {"comment", "function", "keyword", "property", "operator", "string", "number"}
        missing = sorted(expected - present)
        if missing:
            raise SystemExit(f"missing semantic token kinds: {missing}")

        send_message(proc.stdin, {"jsonrpc": "2.0", "id": 3, "method": "shutdown", "params": None})
        _ = read_message(proc.stdout)
        send_message(proc.stdin, {"jsonrpc": "2.0", "method": "exit", "params": None})
    finally:
        proc.stdin.close()
        proc.wait(timeout=5)
        stderr = proc.stderr.read().decode("utf-8")
        if proc.returncode != 0:
            sys.stderr.write(stderr)
            raise SystemExit(proc.returncode)


if __name__ == "__main__":
    main()
