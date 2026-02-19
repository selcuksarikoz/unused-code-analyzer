import * as vscode from 'vscode';
import type { AnalysisResult } from '../types';

const WEBVIEW_HTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="https://cdn.tailwindcss.com"></script>
    <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
        .tab-active { border-bottom: 2px solid #3b82f6; color: #3b82f6; }
        .tab-inactive { color: #6b7280; }
        .tab-inactive:hover { color: #374151; }
    </style>
</head>
<body class="bg-gray-50 h-screen flex flex-col">
    <div class="p-4 bg-white border-b border-gray-200 flex items-center justify-between">
        <h1 class="text-xl font-semibold text-gray-800 flex items-center gap-2">
            <span class="material-icons text-blue-500">code</span>
            Unused Code Analyzer
            <span id="language-badge" class="ml-2 bg-gray-100 text-gray-600 text-xs font-semibold px-2 py-1 rounded">-</span>
        </h1>
        <button onclick="triggerScan()" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg flex items-center gap-2">
            <span class="material-icons text-sm">refresh</span>
            Scan
        </button>
    </div>
    
    <div class="flex border-b border-gray-200 bg-white">
        <button id="tab-imports" class="flex-1 px-6 py-3 text-sm font-medium tab-active flex items-center justify-center gap-2" onclick="switchTab('imports')">
            <span class="material-icons text-sm">download</span>
            Imports
            <span id="imports-count" class="bg-blue-100 text-blue-800 text-xs font-semibold px-2 py-0.5 rounded">0</span>
        </button>
        <button id="tab-variables" class="flex-1 px-6 py-3 text-sm font-medium tab-inactive flex items-center justify-center gap-2" onclick="switchTab('variables')">
            <span class="material-icons text-sm">variable</span>
            Variables
            <span id="variables-count" class="bg-green-100 text-green-800 text-xs font-semibold px-2 py-0.5 rounded">0</span>
        </button>
        <button id="tab-parameters" class="flex-1 px-6 py-3 text-sm font-medium tab-inactive flex items-center justify-center gap-2" onclick="switchTab('parameters')">
            <span class="material-icons text-sm">settings</span>
            Parameters
            <span id="parameters-count" class="bg-yellow-100 text-yellow-800 text-xs font-semibold px-2 py-0.5 rounded">0</span>
        </button>
    </div>

    <div class="flex-1 overflow-auto p-4">
        <div id="content-imports" class="space-y-2">
            <p class="text-gray-500 text-sm">Click "Scan" to analyze your workspace.</p>
        </div>
        <div id="content-variables" class="space-y-2 hidden">
            <p class="text-gray-500 text-sm">Click "Scan" to analyze your workspace.</p>
        </div>
        <div id="content-parameters" class="space-y-2 hidden">
            <p class="text-gray-500 text-sm">Click "Scan" to analyze your workspace.</p>
        </div>
    </div>

    <script>
        const vscode = acquireVsCodeApi();
        let currentTab = 'imports';
        let issues = { imports: [], variables: [], parameters: [] };

        function triggerScan() {
            vscode.postMessage({ type: 'triggerScan' });
        }

        function switchTab(tab) {
            document.querySelectorAll('[id^="tab-"]').forEach(el => {
                el.classList.remove('tab-active');
                el.classList.add('tab-inactive');
            });
            document.querySelectorAll('[id^="content-"]').forEach(el => {
                el.classList.add('hidden');
            });

            document.getElementById('tab-' + tab).classList.remove('tab-inactive');
            document.getElementById('tab-' + tab).classList.add('tab-active');
            document.getElementById('content-' + tab).classList.remove('hidden');
            currentTab = tab;
        }

        function renderIssues() {
            const tabs = ['imports', 'variables', 'parameters'];
            const colors = {
                imports: 'blue',
                variables: 'green',
                parameters: 'yellow'
            };
            const icons = {
                imports: 'download',
                variables: 'variable',
                parameters: 'settings'
            };

            tabs.forEach(tab => {
                const container = document.getElementById('content-' + tab);
                const countEl = document.getElementById(tab + '-count');
                const items = issues[tab];
                
                countEl.textContent = items.length;

                if (items.length === 0) {
                    container.innerHTML = '<p class="text-gray-500 text-sm">No unused ' + tab + ' found.</p>';
                    return;
                }

                container.innerHTML = items.map(item => \`
                    <div class="bg-white border border-gray-200 rounded-lg p-3 hover:shadow-md transition-shadow cursor-pointer" onclick="goToLine(\${item.line}, '\${item.file}')">
                        <div class="flex items-start gap-3">
                            <span class="material-icons text-\${colors[tab]}-500 mt-0.5">\${icons[tab]}</span>
                            <div class="flex-1 min-w-0">
                                <div class="text-sm font-mono text-gray-700 truncate">\${item.text}</div>
                                <div class="text-xs text-gray-500 mt-1">
                                    Line \${item.line} \${item.file ? '• ' + item.file : ''}
                                </div>
                            </div>
                        </div>
                    </div>
                \`).join('');
            });
        }

        function goToLine(line, file) {
            vscode.postMessage({ type: 'goToLine', line: line, file: file });
        }

        window.addEventListener('message', event => {
            const message = event.data;
            if (message.type === 'updateIssues') {
                issues = message.issues;
                if (message.language) {
                    document.getElementById('language-badge').textContent = message.language;
                }
                renderIssues();
            } else if (message.type === 'triggerScan') {
                triggerScan();
            }
        });

        window.addEventListener('load', () => {
            vscode.postMessage({ type: 'ready' });
        });
    </script>
</body>
</html>`;

export class WebviewProvider {
    private panel: vscode.WebviewPanel | null = null;
    private onScanCallback: (() => void) | null = null;

    create(onScan?: () => void): vscode.WebviewPanel {
        this.onScanCallback = onScan || null;
        
        this.panel = vscode.window.createWebviewPanel(
            'unusedImportsViewer',
            'Unused Code Analyzer',
            vscode.ViewColumn.One,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        this.panel.webview.onDidReceiveMessage((message) => {
            if (message.type === 'triggerScan' && this.onScanCallback) {
                this.onScanCallback();
            }
        });

        this.panel.webview.html = WEBVIEW_HTML;

        this.panel.onDidDispose(() => {
            this.panel = null;
            this.onScanCallback = null;
        });

        return this.panel;
    }

    updateContent(issues: AnalysisResult, language?: string): void {
        if (this.panel) {
            this.panel.webview.postMessage({ type: 'updateIssues', issues, language });
        }
    }
}
