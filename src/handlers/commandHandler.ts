import * as vscode from 'vscode';
import * as path from 'path';
import { WasmService } from '../services/wasmService';
import { WebviewProvider } from '../providers/webviewProvider';

export class CommandHandler {
    constructor(
        private wasmService: WasmService,
        private webviewProvider: WebviewProvider
    ) {}

    register(): void {
        vscode.commands.registerCommand('get-unused-imports.analyze', async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) {
                vscode.window.showInformationMessage('No active editor found');
                return;
            }

            const document = editor.document;
            const content = document.getText();
            const filename = path.basename(document.uri.fsPath);
            const language = document.languageId;

            try {
                const result = await this.wasmService.analyze({ content, filename, language });
                this.webviewProvider.updateContent(result, language);
            } catch (error) {
                vscode.window.showErrorMessage(`Analysis failed: ${error}`);
            }
        });
    }
}
