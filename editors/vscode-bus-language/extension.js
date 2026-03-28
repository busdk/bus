"use strict";

const vscode = require("vscode");
const core = require("./bus_language_core");

// activate registers semantic tokens for .bus documents inside VS Code-compatible editors.
// Used by: VS Code-compatible extension package activation from package.json.
function activate(context) {
  const legend = new vscode.SemanticTokensLegend(core.legend.tokenTypes, core.legend.tokenModifiers);
  const provider = {
    provideDocumentSemanticTokens(document) {
      const builder = new vscode.SemanticTokensBuilder(legend);
      for (const token of core.semanticTokens(document.getText())) {
        builder.push(token.line, token.start, token.length, token.type, []);
      }
      return builder.build();
    },
  };
  context.subscriptions.push(
    vscode.languages.registerDocumentSemanticTokensProvider({ language: "bus" }, provider, legend),
  );
}

// deactivate exists for VS Code-compatible extension lifecycle completeness.
// Used by: VS Code-compatible extension host shutdown.
function deactivate() {}

module.exports = {
  activate,
  deactivate,
};
