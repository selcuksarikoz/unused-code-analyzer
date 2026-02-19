import * as vscode from 'vscode';
import * as path from 'path';
import type { AnalysisResult, CodeIssue } from './types';
import { WasmService } from './services/wasmService';
import { WebviewProvider } from './providers/webviewProvider';
import { computeHash } from './utils/hash';
import { isRelevantFile, detectLanguage } from './utils/fileUtils';
import {
    DEFAULT_EXCLUDE_FOLDERS,
    DEFAULT_FILE_EXTENSIONS,
    DEFAULT_AUTO_ANALYZER,
    DEFAULT_AUTO_ANALYZE_DELAY,
    DECORATION_COLOR,
    DECORATION_BORDER,
    DECORATION_TIMEOUT_MS
} from './constants';

interface FileIssue {
    file: string;
    issues: AnalysisResult;
}

class ResultsTreeProvider implements vscode.TreeDataProvider<vscode.TreeItem> {
    private results: FileIssue[] = [];
    private _onDidChangeTreeData = new vscode.EventEmitter<vscode.TreeItem | undefined | void>();
    readonly onDidChangeTreeData = this._onDidChangeTreeData.event;

    setResults(results: FileIssue[]): void {
        this.results = results;
        this._onDidChangeTreeData.fire();
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
                empty.contextValue = 'empty';
                return [empty];
            }
            return this.results.map(result => this.createFileItem(result));
        }

        if (element.contextValue === 'file') {
            return (element as any).children || [];
        }

        return [];
    }

    private createFileItem(result: FileIssue): vscode.TreeItem {
        const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
        const absolutePath = workspaceFolder ? path.join(workspaceFolder.uri.fsPath, result.file) : result.file;
        const totalIssues = 
            result.issues.imports.length + 
            result.issues.variables.length + 
            result.issues.parameters.length;
        
        const allIssues = [...result.issues.imports, ...result.issues.variables, ...result.issues.parameters];
        
        const item = new vscode.TreeItem(
            `${result.file} (${totalIssues})`,
            vscode.TreeItemCollapsibleState.Collapsed
        );
        item.contextValue = 'file';
        item.iconPath = new vscode.ThemeIcon('file-code');
        item.command = {
            command: 'get-unused-imports.highlightFile',
            arguments: [result.file, allIssues],
            title: 'Highlight Issues'
        };

        const children: vscode.TreeItem[] = [];

        result.issues.imports.forEach(issue => {
            const child = new vscode.TreeItem(`Import: ${issue.text} (line ${issue.line})`);
            child.contextValue = 'issue';
            child.iconPath = new vscode.ThemeIcon('download');
            child.command = {
                command: 'get-unused-imports.goTo',
                arguments: [issue, result.file],
                title: 'Go to Issue'
            };
            children.push(child);
        });

        result.issues.variables.forEach(issue => {
            const child = new vscode.TreeItem(`Variable: ${issue.text} (line ${issue.line})`);
            child.contextValue = 'issue';
            child.iconPath = new vscode.ThemeIcon('symbol-variable');
            child.command = {
                command: 'get-unused-imports.goTo',
                arguments: [issue, result.file],
                title: 'Go to Issue'
            };
            children.push(child);
        });

        result.issues.parameters.forEach(issue => {
            const child = new vscode.TreeItem(`Parameter: ${issue.text} (line ${issue.line})`);
            child.contextValue = 'issue';
            child.iconPath = new vscode.ThemeIcon('symbol-parameter');
            child.command = {
                command: 'get-unused-imports.goTo',
                arguments: [issue, result.file],
                title: 'Go to Issue'
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
    private webviewProvider: WebviewProvider;
    private isScanning = false;
    private decorationCollection: vscode.TextEditorDecorationType[] = [];
    private autoAnalyzeTimeout: ReturnType<typeof setTimeout> | undefined;
    private fileHashes: Map<string, string> = new Map();

    constructor(private context: vscode.ExtensionContext) {
        this.wasmService = new WasmService();
        this.treeProvider = new ResultsTreeProvider();
        this.webviewProvider = new WebviewProvider();

        this.registerCommands();
        this.registerTreeView();
        this.registerEventHandlers();
    }

    private isAutoAnalyzerEnabled(): boolean {
        const config = vscode.workspace.getConfiguration('get-unused-imports');
        return config.get<boolean>('autoAnalyzer', DEFAULT_AUTO_ANALYZER);
    }

    private getAutoAnalyzeDelay(): number {
        const config = vscode.workspace.getConfiguration('get-unused-imports');
        return config.get<number>('autoAnalyzeDelay', DEFAULT_AUTO_ANALYZE_DELAY);
    }

    private checkRelevantFile(filePath: string): boolean {
        const config = vscode.workspace.getConfiguration('get-unused-imports');
        const extensions = config.get<string[]>('fileExtensions', DEFAULT_FILE_EXTENSIONS);
        const excludeFolders = config.get<string[]>('excludeFolders', DEFAULT_EXCLUDE_FOLDERS);
        return isRelevantFile(filePath, extensions, excludeFolders);
    }

    private registerEventHandlers(): void {
        vscode.window.onDidChangeActiveTextEditor((editor) => {
            if (!editor) {
                this.decorationCollection.forEach(d => d.dispose());
                this.decorationCollection = [];
            }
        });

        this.context.subscriptions.push(
            vscode.workspace.onDidSaveTextDocument(async (doc) => {
                if (!this.isAutoAnalyzerEnabled()) {
                    return;
                }
                
                const filePath = doc.uri.fsPath;
                if (!this.checkRelevantFile(filePath)) {
                    return;
                }

                await this.scheduleAutoAnalyze(filePath, doc.getText(), doc.languageId);
            })
        );

        this.context.subscriptions.push(
            vscode.workspace.onDidChangeTextDocument(async (event) => {
                if (!this.isAutoAnalyzerEnabled()) {
                    return;
                }
                
                const filePath = event.document.uri.fsPath;
                if (!this.checkRelevantFile(filePath)) {
                    return;
                }

                await this.scheduleAutoAnalyze(filePath, event.document.getText(), event.document.languageId);
            })
        );
    }

    private async scheduleAutoAnalyze(filePath: string, content: string, language: string): Promise<void> {
        if (this.autoAnalyzeTimeout) {
            clearTimeout(this.autoAnalyzeTimeout);
        }

        this.autoAnalyzeTimeout = setTimeout(async () => {
            await this.analyzeSingleFile(filePath, content, language);
        }, this.getAutoAnalyzeDelay());
    }

    private async analyzeSingleFile(filePath: string, content: string, language: string): Promise<void> {
        try {
            const contentHash = computeHash(content);
            const cachedHash = this.fileHashes.get(filePath);
            
            if (cachedHash === contentHash) {
                console.log('[Extension] File unchanged, skipping:', filePath);
                return;
            }
            
            console.log('[Extension] File changed, analyzing:', filePath);
            this.fileHashes.set(filePath, contentHash);

            const result = await this.wasmService.analyze({ 
                content, 
                filename: filePath, 
                language 
            });

            const totalIssues = result.imports.length + result.variables.length + result.parameters.length;
            
            const relativePath = vscode.workspace.asRelativePath(filePath);
            const allIssues = [...result.imports, ...result.variables, ...result.parameters];

            const existingResults = (this.treeProvider as any).results as FileIssue[] | undefined;
            const existingIndex = existingResults?.findIndex(r => r.file === relativePath);

            if (totalIssues > 0) {
                const fileIssue: FileIssue = {
                    file: relativePath,
                    issues: result
                };

                if (existingIndex !== undefined && existingIndex >= 0) {
                    existingResults![existingIndex] = fileIssue;
                } else if (existingResults) {
                    existingResults.push(fileIssue);
                }
            } else if (existingIndex !== undefined && existingIndex >= 0) {
                existingResults?.splice(existingIndex, 1);
            }

            this.treeProvider.refresh();
            
            const currentResults = this.treeProvider.getChildren() as vscode.TreeItem[];
            const allAnalysisResults: AnalysisResult = { imports: [], variables: [], parameters: [] };
            
            for (const item of currentResults) {
                if (item.contextValue === 'file') {
                    const fileResult = existingResults?.find(r => vscode.workspace.asRelativePath(r.file) === (item as any).file);
                    if (fileResult) {
                        allAnalysisResults.imports.push(...fileResult.issues.imports);
                        allAnalysisResults.variables.push(...fileResult.issues.variables);
                        allAnalysisResults.parameters.push(...fileResult.issues.parameters);
                    }
                }
            }
            
            this.webviewProvider.updateContent(allAnalysisResults);
        } catch (error) {
            console.error('[Extension] Auto-analyze error:', error);
        }
    }

    private getExcludePattern(): string {
        const config = vscode.workspace.getConfiguration('get-unused-imports');
        const excludeFolders = config.get<string[]>('excludeFolders', DEFAULT_EXCLUDE_FOLDERS);
        return '**/{' + excludeFolders.join(',') + '}/**';
    }

    private getFileExtensions(): string {
        const config = vscode.workspace.getConfiguration('get-unused-imports');
        const extensions = config.get<string[]>('fileExtensions', DEFAULT_FILE_EXTENSIONS);
        return '**/*.{' + extensions.join(',') + '}';
    }

    async init(): Promise<void> {
        await this.wasmService.initialize();
    }

    private registerCommands(): void {
        this.context.subscriptions.push(
            vscode.commands.registerCommand('get-unused-imports.scanWorkspace', async () => {
                await this.scanWorkspace();
            })
        );

        this.context.subscriptions.push(
            vscode.commands.registerCommand('get-unused-imports.scanFile', async (uri: vscode.Uri) => {
                await this.scanFile(uri);
            })
        );

        this.context.subscriptions.push(
            vscode.commands.registerCommand('get-unused-imports.scanFolder', async (uri: vscode.Uri) => {
                await this.scanFolder(uri);
            })
        );

        this.context.subscriptions.push(
            vscode.commands.registerCommand('get-unused-imports.openWebview', async () => {
                this.webviewProvider.create(() => this.scanWorkspace());
            })
        );

        this.context.subscriptions.push(
            vscode.commands.registerCommand('get-unused-imports.goTo', async (issue: CodeIssue, filePath: string) => {
                await this.highlightIssues(filePath, [issue]);
            })
        );

        this.context.subscriptions.push(
            vscode.commands.registerCommand('get-unused-imports.highlightFile', async (filePath: string, issues: CodeIssue[]) => {
                await this.highlightIssues(filePath, issues);
            })
        );
    }

    private async highlightIssues(filePath: string, issues: CodeIssue[]): Promise<void> {
        this.decorationCollection.forEach(d => d.dispose());
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
            border: DECORATION_BORDER
        });

        const ranges: vscode.Range[] = issues.map(issue => {
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
            this.decorationCollection.forEach(d => d.dispose());
            this.decorationCollection = [];
        }, DECORATION_TIMEOUT_MS);
    }

    private registerTreeView(): void {
        this.treeView = vscode.window.createTreeView('results', {
            treeDataProvider: this.treeProvider,
            showCollapseAll: true
        });

        this.treeView.onDidChangeSelection(async (e) => {
            if (e.selection.length > 0) {
                const item = e.selection[0];
                if (item.contextValue === 'issue' && item.command) {
                    vscode.commands.executeCommand(item.command.command, ...(item.command.arguments || []));
                }
            }
        });

        this.treeView.onDidChangeVisibility(async (e) => {
            if (e.visible && this.isAutoAnalyzerEnabled()) {
                const results = (this.treeProvider as any).results as FileIssue[] | undefined;
                if (!results || results.length === 0) {
                    await this.scanWorkspace();
                }
            }
        });

        this.context.subscriptions.push(this.treeView);
    }

    private async scanWorkspace(): Promise<void> {
        const workspaceFolders = vscode.workspace.workspaceFolders;
        if (!workspaceFolders || workspaceFolders.length === 0) {
            vscode.window.showInformationMessage('No workspace folder open');
            return;
        }

        this.isScanning = true;
        this.treeProvider.clear();
        this.fileHashes.clear();
        vscode.window.showInformationMessage('Scanning workspace for unused code...');

        const files = await vscode.workspace.findFiles(this.getFileExtensions(), this.getExcludePattern());

        if (!files) {
            vscode.window.showInformationMessage('No files found to scan');
            this.isScanning = false;
            return;
        }

        const workspaceFiles: { content: string; filename: string }[] = [];
        
        for (const file of files) {
            try {
                const doc = await vscode.workspace.openTextDocument(file.fsPath);
                const content = doc.getText();
                const hash = computeHash(content);
                this.fileHashes.set(file.fsPath, hash);
                workspaceFiles.push({
                    content,
                    filename: file.fsPath
                });
            } catch (error) {
                console.error(`Error reading ${file.fsPath}:`, error);
            }
        }

        const resultsMap = await this.wasmService.analyzeWorkspace(workspaceFiles);
        
        const allResults: AnalysisResult = { imports: [], variables: [], parameters: [] };
        const fileResults: FileIssue[] = [];
        
        for (const [filename, result] of resultsMap) {
            const totalIssues = result.imports.length + result.variables.length + result.parameters.length;
            
            if (totalIssues > 0) {
                fileResults.push({
                    file: vscode.workspace.asRelativePath(filename),
                    issues: result
                });
                
                allResults.imports.push(...result.imports);
                allResults.variables.push(...result.variables);
                allResults.parameters.push(...result.parameters);
            }
        }

        this.treeProvider.addResults(fileResults);
        this.webviewProvider.updateContent(allResults);

        this.isScanning = false;
        const totalResults = (this.treeProvider as any).results?.length || 0;
        vscode.window.showInformationMessage(`Scan complete! Found ${totalResults} files with issues.`);
    }

    private async scanFile(uri: vscode.Uri): Promise<void> {
        try {
            if (uri.fsPath.includes('.') === false) {
                vscode.window.showInformationMessage('Please select a file, not a folder');
                return;
            }

            const doc = await vscode.workspace.openTextDocument(uri);
            const content = doc.getText();
            const hash = computeHash(content);
            this.fileHashes.set(uri.fsPath, hash);
            
            const workspaceFiles: { content: string; filename: string }[] = [];
            workspaceFiles.push({
                content,
                filename: uri.fsPath
            });

            const resultsMap = await this.wasmService.analyzeWorkspace(workspaceFiles);
            
            const result = resultsMap.get(uri.fsPath);
            if (!result) {
                vscode.window.showInformationMessage('No issues found');
                return;
            }
            
            const fileIssue: FileIssue = {
                file: vscode.workspace.asRelativePath(uri.fsPath),
                issues: result
            };
            
            this.treeProvider.addResults([fileIssue]);
            vscode.window.showInformationMessage(`Analyzed: ${fileIssue.file}`);
        } catch (error) {
            vscode.window.showErrorMessage(`Analysis failed: ${error}`);
        }
    }

    private async scanFolder(uri: vscode.Uri): Promise<void> {
        if (this.isScanning) {
            vscode.window.showInformationMessage('Already scanning...');
            return;
        }

        this.isScanning = true;
        this.treeProvider.clear();
        vscode.window.showInformationMessage('Scanning folder...');

        const ext = this.getFileExtensions().replace('**/*', '');
        const pattern = new vscode.RelativePattern(uri.fsPath, '**/*' + ext);
        const files = await vscode.workspace.findFiles(pattern, this.getExcludePattern());

        if (!files || files.length === 0) {
            vscode.window.showInformationMessage('No files found to scan');
            this.isScanning = false;
            return;
        }

        const workspaceFiles: { content: string; filename: string }[] = [];
        
        for (const file of files) {
            try {
                const doc = await vscode.workspace.openTextDocument(file.fsPath);
                const content = doc.getText();
                const hash = computeHash(content);
                this.fileHashes.set(file.fsPath, hash);
                workspaceFiles.push({
                    content,
                    filename: file.fsPath
                });
            } catch (error) {
                console.error(`Error reading ${file.fsPath}:`, error);
            }
        }

        const resultsMap = await this.wasmService.analyzeWorkspace(workspaceFiles);
        
        const fileResults: FileIssue[] = [];
        
        for (const [filename, result] of resultsMap) {
            const totalIssues = result.imports.length + result.variables.length + result.parameters.length;
            
            if (totalIssues > 0) {
                fileResults.push({
                    file: vscode.workspace.asRelativePath(filename),
                    issues: result
                });
            }
        }

        this.treeProvider.addResults(fileResults);

        this.isScanning = false;
        const totalResults = (this.treeProvider as any).results?.length || 0;
        vscode.window.showInformationMessage(`Scan complete! Found ${totalResults} files with issues.`);
    }

    dispose(): void {
        this.wasmService.terminate();
    }
}

export async function activate(context: vscode.ExtensionContext): Promise<void> {
    console.log('[Extension] Starting activation...');
    const ext = new Extension(context);
    console.log('[Extension] Extension created, initializing WASM...');
    await ext.init();
    console.log('[Extension] Activation complete');
}

export function deactivate(): void {}
