import * as vscode from "vscode";

export class TypeScriptService {
  private static instance: TypeScriptService;

  static getInstance(): TypeScriptService {
    if (!TypeScriptService.instance) {
      TypeScriptService.instance = new TypeScriptService();
    }
    return TypeScriptService.instance;
  }

  async getUnusedImports(
    filePath: string,
  ): Promise<{ line: number; text: string }[]> {
    const issues: { line: number; text: string }[] = [];

    try {
      const uri = vscode.Uri.file(filePath);
      const diagnostics = vscode.languages.getDiagnostics(uri);

      for (const diag of diagnostics) {
        if (diag.source === "ts" || diag.source === "typescript") {
          const code = diag.code;
          if (typeof code === "number") {
            if (code === 6133 || code === 6192) {
              const line = diag.range.start.line + 1;
              issues.push({
                line,
                text: this.extractImportText(diag.message, line, filePath),
              });
            }
          }
        }
      }
    } catch (error) {
      console.error("[TypeScriptService] Error getting diagnostics:", error);
    }

    return issues;
  }

  private extractImportText(
    message: string,
    line: number,
    filePath: string,
  ): string {
    const lines = require("fs").readFileSync(filePath, "utf-8").split("\n");
    if (line > 0 && line <= lines.length) {
      return lines[line - 1].trim();
    }
    return message;
  }
}
