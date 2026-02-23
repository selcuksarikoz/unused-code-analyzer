import * as vscode from "vscode";
import * as path from "path";
import type { AnalysisResult, CodeIssue } from "./types";
import { WasmService } from "./services/wasmService";
import { computeHash } from "./utils/hash";
import { isRelevantFile } from "./utils/fileUtils";
import {
  DEFAULT_EXCLUDE_FOLDERS,
  DEFAULT_FILE_EXTENSIONS,
  DEFAULT_AUTO_ANALYZER,
  DEFAULT_AUTO_ANALYZE_DELAY,
  DECORATION_COLOR,
  DECORATION_BORDER,
  DECORATION_TIMEOUT_MS,
} from "./constants";

interface FileIssue {
  file: string;
  issues: AnalysisResult;
}

class ResultsTreeProvider implements vscode.TreeDataProvider<vscode.TreeItem> {
  private results: FileIssue[] = [];
  private _onDidChangeTreeData = new vscode.EventEmitter<
    vscode.TreeItem | undefined | void
  >();
  readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

  setResults(results: FileIssue[]): void {
    this.results = results;
    this._onDidChangeTreeData.fire();
  }

  getResults(): FileIssue[] {
    return this.results;
  }

  addResult(result: FileIssue): void {
    this.results.push(result);
  }

  addResults(results: FileIssue[]): void {
    this.results = results;
    this._onDidChangeTreeData.fire();
  }

  refresh(): void {
    this._onDidChangeTreeData.fire();
  }

  clear(): void {
    this.results = [];
    this._onDidChangeTreeData.fire(undefined);
  }

  getTreeItem(element: vscode.TreeItem): vscode.TreeItem {
    return element;
  }

  getChildren(element?: vscode.TreeItem): vscode.TreeItem[] {
    if (!element) {
      if (this.results.length === 0) {
        const empty = new vscode.TreeItem('Click "Scan Workspace" to analyze');
        empty.contextValue = "empty";
        return [empty];
      }
      return this.results.map((result) => this.createFileItem(result));
    }

    if (element.contextValue === "file") {
      return (element as any).children || [];
    }

    return [];
  }

  private createFileItem(result: FileIssue): vscode.TreeItem {
    const totalIssues =
      result.issues.imports.length +
      result.issues.variables.length +
      result.issues.parameters.length;

    const allIssues = [
      ...result.issues.imports,
      ...result.issues.variables,
      ...result.issues.parameters,
    ];

    const item = new vscode.TreeItem(
      `${result.file} (${totalIssues})`,
      vscode.TreeItemCollapsibleState.Collapsed,
    );
    item.contextValue = "file";
    item.iconPath = new vscode.ThemeIcon("file-code");
    item.command = {
      command: "get-unused-imports.highlightFile",
      arguments: [result.file, allIssues],
      title: "Highlight Issues",
    };

    const children: vscode.TreeItem[] = [];

    result.issues.imports.forEach((issue) => {
      const child = new vscode.TreeItem(
        `Import: ${issue.text} (line ${issue.line})`,
      );
      child.contextValue = "issue";
      child.iconPath = new vscode.ThemeIcon("download");
      child.command = {
        command: "get-unused-imports.goTo",
        arguments: [issue, result.file],
        title: "Go to Issue",
      };
      children.push(child);
    });

    result.issues.variables.forEach((issue) => {
      const child = new vscode.TreeItem(
        `Variable: ${issue.text} (line ${issue.line})`,
      );
      child.contextValue = "issue";
      child.iconPath = new vscode.ThemeIcon("symbol-variable");
      child.command = {
        command: "get-unused-imports.goTo",
        arguments: [issue, result.file],
        title: "Go to Issue",
      };
      children.push(child);
    });

    result.issues.parameters.forEach((issue) => {
      const child = new vscode.TreeItem(
        `Parameter: ${issue.text} (line ${issue.line})`,
      );
      child.contextValue = "issue";
      child.iconPath = new vscode.ThemeIcon("symbol-parameter");
      child.command = {
        command: "get-unused-imports.goTo",
        arguments: [issue, result.file],
        title: "Go to Issue",
      };
      children.push(child);
    });

    (item as any).children = children;
    return item;
  }
}

class Extension implements vscode.Disposable {
  private wasmService: WasmService;
  private treeProvider: ResultsTreeProvider;
  private treeView: vscode.TreeView<vscode.TreeItem> | undefined;
  private isScanning = false;
  private decorationCollection: vscode.TextEditorDecorationType[] = [];
  private autoAnalyzeTimeout: ReturnType<typeof setTimeout> | undefined;
  private fileHashes: Map<string, string> = new Map();
  private fileIssues: FileIssue[] = [];

  constructor(private context: vscode.ExtensionContext) {
    this.wasmService = new WasmService();
    this.treeProvider = new ResultsTreeProvider();

    this.registerCommands();
    this.registerTreeView();
    this.registerEventHandlers();
  }

  private isAutoAnalyzerEnabled(): boolean {
    const config = vscode.workspace.getConfiguration("get-unused-imports");
    return config.get<boolean>("autoAnalyzer", DEFAULT_AUTO_ANALYZER);
  }

  private getAutoAnalyzeDelay(): number {
    const config = vscode.workspace.getConfiguration("get-unused-imports");
    return config.get<number>("autoAnalyzeDelay", DEFAULT_AUTO_ANALYZE_DELAY);
  }

  private getEnabledExtensions(): string[] {
    const config = vscode.workspace.getConfiguration("get-unused-imports");
    const configured =
      config.get<string[]>("fileExtensions", DEFAULT_FILE_EXTENSIONS) || [];
    const merged = new Set<string>([
      ...DEFAULT_FILE_EXTENSIONS.map((ext) => ext.toLowerCase()),
      ...configured.map((ext) => ext.toLowerCase()),
    ]);
    return [...merged];
  }

  private checkRelevantFile(filePath: string): boolean {
    const config = vscode.workspace.getConfiguration("get-unused-imports");
    const extensions = this.getEnabledExtensions();
    const excludeFolders = config.get<string[]>(
      "excludeFolders",
      DEFAULT_EXCLUDE_FOLDERS,
    );
    return isRelevantFile(filePath, extensions, excludeFolders);
  }

  private registerEventHandlers(): void {
    vscode.window.onDidChangeActiveTextEditor((editor) => {
      if (!editor) {
        this.decorationCollection.forEach((d) => d.dispose());
        this.decorationCollection = [];
      }
    });

    this.context.subscriptions.push(
      vscode.workspace.onDidSaveTextDocument(async (doc) => {
        if (!this.isAutoAnalyzerEnabled()) {
          return;
        }

        const filePath = doc.uri.fsPath;
        console.log(
          "[Extension] onDidSaveTextDocument:",
          filePath,
          "isRelevant:",
          this.checkRelevantFile(filePath),
        );
        if (!this.checkRelevantFile(filePath)) {
          return;
        }

        await this.scheduleAutoAnalyze(filePath, doc.getText(), doc.languageId);
      }),
    );

    this.context.subscriptions.push(
      vscode.workspace.onDidChangeTextDocument(async (event) => {
        if (!this.isAutoAnalyzerEnabled()) {
          return;
        }

        const filePath = event.document.uri.fsPath;
        console.log(
          "[Extension] onDidChangeTextDocument:",
          filePath,
          "isRelevant:",
          this.checkRelevantFile(filePath),
        );
        if (!this.checkRelevantFile(filePath)) {
          return;
        }

        await this.scheduleAutoAnalyze(
          filePath,
          event.document.getText(),
          event.document.languageId,
        );
      }),
    );
  }

  private async scheduleAutoAnalyze(
    filePath: string,
    content: string,
    language: string,
  ): Promise<void> {
    if (this.autoAnalyzeTimeout) {
      clearTimeout(this.autoAnalyzeTimeout);
    }

    this.autoAnalyzeTimeout = setTimeout(async () => {
      await this.analyzeSingleFile(filePath, content, language);
    }, this.getAutoAnalyzeDelay());
  }

  private async analyzeSingleFile(
    filePath: string,
    content: string,
    language: string,
  ): Promise<void> {
    try {
      console.log(
        "[Extension] analyzeSingleFile called:",
        filePath,
        "language:",
        language,
      );

      const contentHash = computeHash(content);
      const cachedHash = this.fileHashes.get(filePath);

      if (cachedHash === contentHash) {
        console.log("[Extension] File unchanged, skipping:", filePath);
        return;
      }

      console.log("[Extension] File changed, analyzing:", filePath);
      this.fileHashes.set(filePath, contentHash);

      const result = await this.wasmService.analyze({
        content,
        filename: filePath,
        language,
      });

      console.log(
        "[Extension] Analysis result:",
        filePath,
        "imports:",
        result.imports.length,
        "variables:",
        result.variables.length,
        "parameters:",
        result.parameters.length,
      );

      const totalIssues =
        result.imports.length +
        result.variables.length +
        result.parameters.length;

      const relativePath = vscode.workspace.asRelativePath(filePath);

      const existingResults = this.treeProvider.getResults();
      const existingIndex = existingResults.findIndex(
        (r) => r.file === relativePath,
      );

      if (totalIssues > 0) {
        const fileIssue: FileIssue = {
          file: relativePath,
          issues: result,
        };

        if (existingIndex >= 0) {
          existingResults[existingIndex] = fileIssue;
        } else {
          existingResults.push(fileIssue);
        }
      } else if (existingIndex >= 0) {
        existingResults.splice(existingIndex, 1);
      }

      this.treeProvider.setResults([...existingResults]);

      const allAnalysisResults: AnalysisResult = {
        imports: [],
        variables: [],
        parameters: [],
      };

      for (const fileResult of existingResults) {
        allAnalysisResults.imports.push(...fileResult.issues.imports);
        allAnalysisResults.variables.push(...fileResult.issues.variables);
        allAnalysisResults.parameters.push(...fileResult.issues.parameters);
      }
    } catch (error) {
      console.error("[Extension] Auto-analyze error:", error);
    }
  }

  private getExcludePattern(): string {
    const config = vscode.workspace.getConfiguration("get-unused-imports");
    const excludeFolders = config.get<string[]>(
      "excludeFolders",
      DEFAULT_EXCLUDE_FOLDERS,
    );
    return "**/{" + excludeFolders.join(",") + "}/**";
  }

  private getFileExtensions(): string {
    const extensions = this.getEnabledExtensions();
    return "**/*." + extensions[0];
  }

  private async getAllFiles(): Promise<vscode.Uri[]> {
    const extensions = this.getEnabledExtensions();
    const excludePattern = this.getExcludePattern();

    const allFiles: vscode.Uri[] = [];

    for (const ext of extensions) {
      const pattern = `**/*.${ext}`;
      const files = await vscode.workspace.findFiles(pattern, excludePattern);
      allFiles.push(...files);
    }

    return allFiles;
  }

  private async buildWorkspaceAnalysisInput(): Promise<{
    files: { content: string; filename: string }[];
    scannedCount: number;
  }> {
    const uris = await this.getAllFiles();
    const unique = new Map<string, vscode.Uri>();
    for (const uri of uris) {
      unique.set(uri.fsPath, uri);
    }

    const workspaceFiles: { content: string; filename: string }[] = [];
    for (const uri of unique.values()) {
      try {
        const doc = await vscode.workspace.openTextDocument(uri.fsPath);
        const content = doc.getText();
        const hash = computeHash(content);
        this.fileHashes.set(uri.fsPath, hash);
        workspaceFiles.push({
          content,
          filename: uri.fsPath,
        });
      } catch (error) {
        console.error(`Error reading ${uri.fsPath}:`, error);
      }
    }

    return {
      files: workspaceFiles,
      scannedCount: unique.size,
    };
  }

  private isPathInFolder(folderPath: string, filePath: string): boolean {
    const relative = path.relative(folderPath, filePath);
    return (
      relative === "" ||
      (!relative.startsWith("..") && !path.isAbsolute(relative))
    );
  }

  async init(): Promise<void> {
    await this.wasmService.initialize();
    if (
      this.isAutoAnalyzerEnabled() &&
      vscode.workspace.workspaceFolders &&
      vscode.workspace.workspaceFolders.length > 0
    ) {
      this.scanWorkspace().catch((error) => {
        console.error("[Extension] Initial workspace scan failed:", error);
      });
    }
  }

  private registerCommands(): void {
    this.context.subscriptions.push(
      vscode.commands.registerCommand(
        "get-unused-imports.scanWorkspace",
        async () => {
          await this.scanWorkspace();
        },
      ),
    );

    this.context.subscriptions.push(
      vscode.commands.registerCommand(
        "get-unused-imports.scanFile",
        async (uri?: vscode.Uri) => {
          await this.scanFile(uri);
        },
      ),
    );

    this.context.subscriptions.push(
      vscode.commands.registerCommand(
        "get-unused-imports.scanFolder",
        async (uri?: vscode.Uri) => {
          await this.scanFolder(uri);
        },
      ),
    );

    this.context.subscriptions.push(
      vscode.commands.registerCommand(
        "get-unused-imports.goTo",
        async (issue: CodeIssue, filePath: string) => {
          await this.highlightIssues(filePath, [issue]);
        },
      ),
    );

    this.context.subscriptions.push(
      vscode.commands.registerCommand(
        "get-unused-imports.highlightFile",
        async (filePath: string, issues: CodeIssue[]) => {
          await this.highlightIssues(filePath, issues);
        },
      ),
    );
  }

  private async highlightIssues(
    filePath: string,
    issues: CodeIssue[],
  ): Promise<void> {
    this.decorationCollection.forEach((d) => d.dispose());
    this.decorationCollection = [];

    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      return;
    }
    const absolutePath = path.join(workspaceFolder.uri.fsPath, filePath);
    const doc = await vscode.workspace.openTextDocument(absolutePath);
    const editor = await vscode.window.showTextDocument(doc);

    const decorationType = vscode.window.createTextEditorDecorationType({
      backgroundColor: DECORATION_COLOR,
      border: DECORATION_BORDER,
    });

    const ranges: vscode.Range[] = issues.map((issue) => {
      const start = new vscode.Position(issue.line - 1, 0);
      const end = new vscode.Position(issue.line, 0);
      return new vscode.Range(start, end);
    });

    editor.setDecorations(decorationType, ranges);
    this.decorationCollection.push(decorationType);

    if (issues.length > 0) {
      const firstPosition = new vscode.Position(issues[0].line - 1, 0);
      editor.selection = new vscode.Selection(firstPosition, firstPosition);
      editor.revealRange(new vscode.Range(firstPosition, firstPosition));
    }

    setTimeout(() => {
      this.decorationCollection.forEach((d) => d.dispose());
      this.decorationCollection = [];
    }, DECORATION_TIMEOUT_MS);
  }

  private registerTreeView(): void {
    this.treeView = vscode.window.createTreeView("get-unused-imports.results", {
      treeDataProvider: this.treeProvider,
      showCollapseAll: true,
    });

    this.treeView.onDidChangeSelection(async (e) => {
      if (e.selection.length > 0) {
        const item = e.selection[0];
        if (item.contextValue === "issue" && item.command) {
          vscode.commands.executeCommand(
            item.command.command,
            ...(item.command.arguments || []),
          );
        }
      }
    });

    this.treeView.onDidChangeVisibility(async (e) => {
      console.log("[Extension] TreeView visibility changed:", e.visible);
      if (e.visible && this.isAutoAnalyzerEnabled()) {
        const results = this.treeProvider.getResults();
        console.log("[Extension] Current results count:", results.length);
        if (results.length === 0) {
          console.log("[Extension] No results, running initial scan...");
          await this.scanWorkspace();
        }
      }
    });

    this.context.subscriptions.push(this.treeView);
  }

  private async scanWorkspace(): Promise<void> {
    if (this.isScanning) {
      return;
    }

    const workspaceFolders = vscode.workspace.workspaceFolders;
    if (!workspaceFolders || workspaceFolders.length === 0) {
      vscode.window.showInformationMessage("No workspace folder open");
      return;
    }

    this.isScanning = true;
    this.treeProvider.clear();
    this.fileHashes.clear();

    try {
      let scannedCount = 0;
      let totalResults = 0;

      const {
        files: workspaceFiles,
        scannedCount: totalScannedFiles,
      } = await this.buildWorkspaceAnalysisInput();
      scannedCount = totalScannedFiles;

      if (workspaceFiles.length > 0) {
        const resultsMap = await this.wasmService.analyzeWorkspace(workspaceFiles);

        const allResults: AnalysisResult = {
          imports: [],
          variables: [],
          parameters: [],
        };
        const fileResults: FileIssue[] = [];

        for (const [filename, result] of resultsMap) {
          const issuesCount =
            result.imports.length +
            result.variables.length +
            result.parameters.length;

          if (issuesCount > 0) {
            fileResults.push({
              file: vscode.workspace.asRelativePath(filename),
              issues: result,
            });

            allResults.imports.push(...result.imports);
            allResults.variables.push(...result.variables);
            allResults.parameters.push(...result.parameters);
          }
        }

        this.treeProvider.addResults(fileResults);
      }

      totalResults = this.treeProvider.getResults().length || 0;
      vscode.window.showInformationMessage(
        `Scan complete. Scanned ${scannedCount} files, found ${totalResults} files with issues.`,
      );
    } finally {
      this.isScanning = false;
    }
  }

  private async scanFile(uri?: vscode.Uri): Promise<void> {
    try {
      vscode.commands.executeCommand(
        "workbench.view.extension.get-unused-imports",
      );

      const targetUri = uri ?? vscode.window.activeTextEditor?.document.uri;
      if (!targetUri) {
        vscode.window.showInformationMessage(
          "No file selected or active editor found",
        );
        return;
      }

      const targetStat = await vscode.workspace.fs.stat(targetUri);
      if ((targetStat.type & vscode.FileType.File) === 0) {
        vscode.window.showInformationMessage(
          "Please select a file, not a folder",
        );
        return;
      }

      const {
        files: workspaceFiles,
        scannedCount,
      } = await this.buildWorkspaceAnalysisInput();
      const resultsMap = await this.wasmService.analyzeWorkspace(workspaceFiles);

      const result = resultsMap.get(targetUri.fsPath);
      if (
        !result ||
        result.imports.length + result.variables.length + result.parameters.length ===
          0
      ) {
        this.treeProvider.clear();
        vscode.window.showInformationMessage(
          `Analyzed file with workspace cross-reference (${scannedCount} files): no issues found.`,
        );
        return;
      }

      const fileIssue: FileIssue = {
        file: vscode.workspace.asRelativePath(targetUri.fsPath),
        issues: result,
      };

      this.treeProvider.addResults([fileIssue]);
      vscode.window.showInformationMessage(
        `Analyzed file with workspace cross-reference (${scannedCount} files): ${fileIssue.file}`,
      );
    } catch (error) {
      vscode.window.showErrorMessage(`Analysis failed: ${error}`);
    }
  }

  private async scanFolder(uri?: vscode.Uri): Promise<void> {
    if (this.isScanning) {
      vscode.window.showInformationMessage("Already scanning...");
      return;
    }

    vscode.commands.executeCommand(
      "workbench.view.extension.get-unused-imports",
    );

    const targetUri = uri ?? vscode.workspace.workspaceFolders?.[0]?.uri;
    if (!targetUri) {
      vscode.window.showInformationMessage(
        "No folder selected and no workspace is open",
      );
      return;
    }

    const targetStat = await vscode.workspace.fs.stat(targetUri);
    if ((targetStat.type & vscode.FileType.Directory) === 0) {
      vscode.window.showInformationMessage("Please select a folder");
      return;
    }

    this.isScanning = true;
    this.treeProvider.clear();

    try {
      let scannedCount = 0;
      let totalResults = 0;

      const {
        files: workspaceFiles,
      } = await this.buildWorkspaceAnalysisInput();

      if (workspaceFiles.length > 0) {
        const resultsMap = await this.wasmService.analyzeWorkspace(workspaceFiles);

        const fileResults: FileIssue[] = [];

        for (const [filename, result] of resultsMap) {
          if (!this.isPathInFolder(targetUri.fsPath, filename)) {
            continue;
          }

          scannedCount++;

          const issuesCount =
            result.imports.length +
            result.variables.length +
            result.parameters.length;

          if (issuesCount > 0) {
            fileResults.push({
              file: vscode.workspace.asRelativePath(filename),
              issues: result,
            });
          }
        }

        this.treeProvider.addResults(fileResults);
      }

      totalResults = this.treeProvider.getResults().length || 0;
      vscode.window.showInformationMessage(
        `Scan complete. Scanned ${scannedCount} files, found ${totalResults} files with issues.`,
      );
    } finally {
      this.isScanning = false;
    }
  }

  dispose(): void {
    this.wasmService.terminate();
  }
}
export async function activate(
  context: vscode.ExtensionContext,
): Promise<void> {
  console.log("[Extension] Starting activation...");
  const ext = new Extension(context);
  context.subscriptions.push(ext);
  console.log("[Extension] Extension created, initializing analyzers...");
  try {
    await ext.init();
    console.log("[Extension] Activation complete");
  } catch (error) {
    console.error(
      "[Extension] Initialization failed, extension will continue in degraded mode:",
      error,
    );
    vscode.window.showWarningMessage(
      "Get Unused Imports initialized with limited capabilities. JS/TS analysis is still available.",
    );
  }
}

export function deactivate(): void {}
