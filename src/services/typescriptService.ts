import * as vscode from "vscode";
import type { AnalysisResult, CodeIssue } from "../types";

type SupportedLanguage = "javascript" | "javascriptreact" | "typescript" | "typescriptreact" | "svelte" | "astro" | "vue";

interface UnusedSymbol {
  name: string;
  line: number;
  kind: "import" | "variable" | "parameter";
}

export class TypeScriptService {
  private static instance: TypeScriptService;

  static getInstance(): TypeScriptService {
    if (!TypeScriptService.instance) {
      TypeScriptService.instance = new TypeScriptService();
    }
    return TypeScriptService.instance;
  }

  async analyzeFile(document: vscode.TextDocument): Promise<AnalysisResult> {
    if (!this.isSupportedLanguage(document.languageId)) {
      return { imports: [], variables: [], parameters: [] };
    }

    const unusedSymbols = await this.findUnusedSymbolsFromCodeActions(document);

    const result: AnalysisResult = {
      imports: [],
      variables: [],
      parameters: [],
    };

    const filePath = document.uri.fsPath;

    for (const symbol of unusedSymbols) {
      const lineText = document.lineAt(symbol.line).text.trim();
      const issue: CodeIssue = {
        id: this.generateUUID(),
        line: symbol.line + 1,
        text: lineText || symbol.name,
        file: filePath,
      };

      switch (symbol.kind) {
        case "import":
          result.imports.push(issue);
          break;
        case "variable":
          result.variables.push(issue);
          break;
        case "parameter":
          result.parameters.push(issue);
          break;
      }
    }

    return result;
  }

  async analyzeWorkspace(files: vscode.Uri[]): Promise<Map<string, AnalysisResult>> {
    const results = new Map<string, AnalysisResult>();

    for (const uri of files) {
      try {
        const document = await vscode.workspace.openTextDocument(uri);
        const result = await this.analyzeFile(document);
        results.set(uri.fsPath, result);
      } catch (error) {
        console.error(`[TypeScriptService] Error analyzing ${uri.fsPath}:`, error);
      }
    }

    return results;
  }

  private async findUnusedSymbolsFromCodeActions(document: vscode.TextDocument): Promise<UnusedSymbol[]> {
    const symbols: UnusedSymbol[] = [];
    const diagnostics = vscode.languages.getDiagnostics(document.uri);

    const tsDiagnostics = diagnostics.filter(
      (d) => (d.source === "ts" || d.source === "typescript") &&
        [6192, 6133, 6138, 6196, 1484].includes(typeof d.code === "number" ? d.code : 0)
    );

    for (const diagnostic of tsDiagnostics) {
      const code = typeof diagnostic.code === "number" ? diagnostic.code : 0;
      const line = diagnostic.range.start.line;

      const codeActions = await vscode.commands.executeCommand<vscode.CodeAction[]>(
        "vscode.executeCodeActionProvider",
        document.uri,
        diagnostic.range,
        vscode.CodeActionKind.QuickFix.value
      );

      if (!codeActions || codeActions.length === 0) {
        continue;
      }

      const symbol = this.identifySymbolFromCodeActions(document, diagnostic, code, codeActions);
      if (symbol) {
        symbols.push(symbol);
      }
    }

    const codeLensSymbols = await this.findUnusedSymbolsFromCodeLens(document);
    symbols.push(...codeLensSymbols);

    return this.deduplicateSymbols(symbols);
  }

  private async findUnusedSymbolsFromCodeLens(document: vscode.TextDocument): Promise<UnusedSymbol[]> {
    const symbols: UnusedSymbol[] = [];

    try {
      const codeLenses = await vscode.commands.executeCommand<vscode.CodeLens[]>(
        "vscode.executeCodeLensProvider",
        document.uri
      );

      if (!codeLenses || codeLenses.length === 0) {
        return symbols;
      }

      for (const codeLens of codeLenses) {
        if (!codeLens.isResolved) {
          continue;
        }

        const command = codeLens.command;
        if (!command) {
          continue;
        }

        const title = command.title.toLowerCase();

        if (title.includes("0 references") || title.includes("0 ref")) {
          const line = codeLens.range.start.line;
          const identifier = this.getIdentifierAtLine(document, line);

          if (identifier) {
            symbols.push({
              name: identifier,
              line,
              kind: "variable",
            });
          }
        }
      }
    } catch (error) {
      console.error("[TypeScriptService] CodeLens error:", error);
    }

    return symbols;
  }

  private deduplicateSymbols(symbols: UnusedSymbol[]): UnusedSymbol[] {
    const seen = new Set<string>();
    return symbols.filter((symbol) => {
      const key = `${symbol.line}:${symbol.name}:${symbol.kind}`;
      if (seen.has(key)) {
        return false;
      }
      seen.add(key);
      return true;
    });
  }

  private getIdentifierAtLine(document: vscode.TextDocument, line: number): string | null {
    const lineText = document.lineAt(line).text;
    const trimmed = lineText.trim();

    let start = 0;
    while (start < lineText.length && /\s/.test(lineText[start])) {
      start++;
    }

    let end = start;
    while (end < lineText.length && /[a-zA-Z0-9_$]/.test(lineText[end])) {
      end++;
    }

    return lineText.substring(start, end) || null;
  }

  private identifySymbolFromCodeActions(
    document: vscode.TextDocument,
    diagnostic: vscode.Diagnostic,
    code: number,
    codeActions: vscode.CodeAction[]
  ): UnusedSymbol | null {
    const line = diagnostic.range.start.line;
    const message = diagnostic.message.toLowerCase();

    for (const action of codeActions) {
      if (!action.title.toLowerCase().includes("unused")) {
        continue;
      }

      switch (code) {
        case 6192:
        case 1484:
          return {
            name: action.title,
            line,
            kind: "import",
          };

        case 6133:
          if (message.includes("import") || message.includes("type")) {
            return {
              name: action.title,
              line,
              kind: "import",
            };
          }
          return {
            name: action.title,
            line,
            kind: "variable",
          };

        case 6138:
        case 6196:
          return {
            name: action.title,
            line,
            kind: "parameter",
          };
      }
    }

    return this.identifySymbolFromDiagnosticOnly(document, diagnostic, code);
  }

  private identifySymbolFromDiagnosticOnly(
    document: vscode.TextDocument,
    diagnostic: vscode.Diagnostic,
    code: number
  ): UnusedSymbol | null {
    const line = diagnostic.range.start.line;
    const character = diagnostic.range.start.character;

    const lineText = document.lineAt(line).text;

    let start = character;
    let end = character;

    while (start > 0 && /[a-zA-Z0-9_$]/.test(lineText[start - 1] || "")) {
      start--;
    }
    while (end < lineText.length && /[a-zA-Z0-9_$]/.test(lineText[end] || "")) {
      end++;
    }

    const identifier = lineText.substring(start, end);

    let kind: "import" | "variable" | "parameter";
    switch (code) {
      case 6192:
      case 1484:
        kind = "import";
        break;
      case 6133:
        kind = "variable";
        break;
      case 6138:
      case 6196:
        kind = "parameter";
        break;
      default:
        return null;
    }

    return {
      name: identifier || "unused symbol",
      line,
      kind,
    };
  }

  private isSupportedLanguage(languageId: string): languageId is SupportedLanguage {
    return [
      "javascript",
      "javascriptreact",
      "typescript",
      "typescriptreact",
      "svelte",
      "astro",
      "vue",
    ].includes(languageId);
  }

  private generateUUID(): string {
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0;
      const v = c === "x" ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
  }

  async organizeImports(document: vscode.TextDocument): Promise<void> {
    const command = this.getCommandForLanguage(document.languageId, "organizeImports");
    if (!command) {
      return;
    }

    try {
      await vscode.commands.executeCommand(command, document.uri);
    } catch (error) {
      console.error("[TypeScriptService] Organize imports failed:", error);
    }
  }

  async removeUnusedImports(document: vscode.TextDocument): Promise<void> {
    const command = this.getCommandForLanguage(document.languageId, "removeUnusedImports");
    if (!command) {
      return;
    }

    try {
      await vscode.commands.executeCommand(command, document.uri);
    } catch (error) {
      console.error("[TypeScriptService] Remove unused imports failed:", error);
    }
  }

  private getCommandForLanguage(
    languageId: string,
    action: "organizeImports" | "removeUnusedImports"
  ): string | null {
    const frameworkCommands: Record<string, string> = {
      svelte: `svelte.${action}`,
      astro: `astro.${action}`,
      vue: `vue.${action}`,
    };

    if (frameworkCommands[languageId]) {
      return frameworkCommands[languageId];
    }

    if (languageId === "typescript" || languageId === "typescriptreact") {
      return `typescript.${action}`;
    }

    if (
      languageId === "javascript" ||
      languageId === "javascriptreact"
    ) {
      return `javascript.${action}`;
    }

    return null;
  }
}
