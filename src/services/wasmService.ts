import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import * as crypto from 'crypto';
import type { AnalysisResult, AnalyzeRequest } from '../types';
import { EXTENSION_ID } from '../constants';
import { NativeJsAnalyzer } from './nativeJsAnalyzer';

let wasmInitialized = false;
let analyzeCodeFn: any = null;
let analyzeWorkspaceFn: any = null;
let detectLanguageFn: any = null;

interface WorkspaceFile {
  content: string;
  filename: string;
  hash?: string;
}

export class WasmService {
  private wasmModule: any = null;
  private extensionId = EXTENSION_ID;
  private cache: Map<string, { hash: string; result: AnalysisResult }> = new Map();
  private workspaceHashes: Map<string, string> = new Map();
  private nativeJsAnalyzer: NativeJsAnalyzer;

  constructor() {
    this.nativeJsAnalyzer = new NativeJsAnalyzer();
  }

  private computeHash(content: string): string {
    return crypto.createHash('md5').update(content).digest('hex');
  }

  private isJsTsFile(filename: string): boolean {
    const ext = filename.toLowerCase().split('.').pop();
    return ['ts', 'tsx', 'js', 'jsx', 'mjs', 'cjs', 'vue', 'svelte'].includes(ext || '');
  }

  private isGoFile(filename: string): boolean {
    return filename.toLowerCase().endsWith('.go');
  }

  private isPythonFile(filename: string): boolean {
    return filename.toLowerCase().endsWith('.py');
  }

  async initialize(): Promise<void> {
    if (wasmInitialized) {
      console.log('[WASMService] WASM already initialized');
      return;
    }
    
    console.log('[WASMService] Starting initialization...');
    
    const timeoutMs = 30000;
    const timeoutPromise = new Promise((_, reject) => 
      setTimeout(() => reject(new Error(`Initialization timed out after ${timeoutMs}ms`)), timeoutMs)
    );
    
    const initPromise = this.doInitialize();
    
    try {
      await Promise.race([initPromise, timeoutPromise]);
    } catch (error) {
      console.error('[WASMService] Initialization failed:', error);
      throw error;
    }
  }
  
  private async doInitialize(): Promise<void> {
    try {
      console.log('[WASMService] Getting extension path...');
      const extensionPath = vscode.extensions.getExtension(this.extensionId)?.extensionPath;
      console.log('[WASMService] Extension path:', extensionPath);
      if (!extensionPath) {
        throw new Error('Extension path not found');
      }

      const wasmPath = path.join(extensionPath, 'out', 'main.wasm');
      const wasmExecPath = path.join(extensionPath, 'out', 'wasm_exec.js');
      
      console.log('[WASMService] Checking files...');
      console.log('[WASMService] WASM path:', wasmPath);
      console.log('[WASMService] WASM exists:', fs.existsSync(wasmPath));
      console.log('[WASMService] wasm_exec exists:', fs.existsSync(wasmExecPath));

      if (!fs.existsSync(wasmPath)) {
        throw new Error(`WASM file not found at ${wasmPath}`);
      }

      console.log('[WASMService] Reading WASM binary...');
      const wasmBinary = fs.readFileSync(wasmPath);
      console.log('[WASMService] WASM binary size:', wasmBinary.length);
      
      console.log('[WASMService] Loading wasm_exec.js...');
      require(wasmExecPath);
      console.log('[WASMService] wasm_exec.js loaded');
      
      console.log('[WASMService] Creating Go instance...');
      const go = new (globalThis as any).Go();
      console.log('[WASMService] Go instance created');
      
      console.log('[WASMService] Instantiating WASM...');
      const WebAssembly = globalThis.WebAssembly;
      const result = await WebAssembly.instantiate(wasmBinary, go.importObject);
      console.log('[WASMService] WASM instantiated');
      
      console.log('[WASMService] Running Go...');
      go.run(result.instance);
      console.log('[WASMService] Go started (non-blocking)');
      
      analyzeCodeFn = (globalThis as any).analyzeCode;
      analyzeWorkspaceFn = (globalThis as any).analyzeWorkspace;
      detectLanguageFn = (globalThis as any).detectLanguage;
      
      console.log('[WASMService] analyzeCode function:', typeof analyzeCodeFn);
      console.log('[WASMService] analyzeWorkspace function:', typeof analyzeWorkspaceFn);
      console.log('[WASMService] detectLanguage function:', typeof detectLanguageFn);
      
      wasmInitialized = true;
      console.log('[WASMService] Initialization complete!');
    } catch (error) {
      console.error('[WASMService] Initialization error:', error);
      throw error;
    }
  }

  async ensureInitialized(): Promise<void> {
    if (!wasmInitialized) {
      await this.initialize();
    }
  }

  async analyzeWorkspace(files: WorkspaceFile[]): Promise<Map<string, AnalysisResult>> {
    const jsTsFiles: WorkspaceFile[] = [];
    const wasmFiles: WorkspaceFile[] = [];

    for (const file of files) {
      if (this.isJsTsFile(file.filename)) {
        jsTsFiles.push(file);
      } else if (this.isGoFile(file.filename) || this.isPythonFile(file.filename)) {
        wasmFiles.push(file);
      }
    }

    const results = new Map<string, AnalysisResult>();

    if (jsTsFiles.length > 0) {
      console.log(`[WASMService] Analyzing ${jsTsFiles.length} JS/TS files natively...`);
      const jsResults = this.nativeJsAnalyzer.analyzeWorkspace(jsTsFiles);
      for (const [filename, result] of jsResults) {
        results.set(filename, result);
      }
    }

    if (wasmFiles.length > 0) {
      const wasmResults = await this.analyzeWorkspaceWasm(wasmFiles);
      for (const [filename, result] of wasmResults) {
        results.set(filename, result);
      }
    }

    return results;
  }

  private async analyzeWorkspaceWasm(files: WorkspaceFile[]): Promise<Map<string, AnalysisResult>> {
    await this.ensureInitialized();

    if (!analyzeWorkspaceFn) {
      console.error('[WASMService] analyzeWorkspace not initialized');
      return new Map();
    }

    const filesWithHash = files.map(f => ({
      ...f,
      hash: this.computeHash(f.content)
    }));

    const cachedResults = new Map<string, AnalysisResult>();
    const filesToAnalyze: WorkspaceFile[] = [];
    
    for (const file of filesWithHash) {
      const cached = this.cache.get(file.filename);
      if (cached && cached.hash === file.hash) {
        cachedResults.set(file.filename, cached.result);
      } else {
        filesToAnalyze.push(file);
      }
    }

    console.log(`[WASMService] Cache hit: ${cachedResults.size}, to analyze: ${filesToAnalyze.length}`);

    if (filesToAnalyze.length === 0) {
      return cachedResults;
    }

    try {
      const start = Date.now();
      const result = analyzeWorkspaceFn(JSON.stringify({ files: filesToAnalyze }));
      const elapsed = Date.now() - start;
      console.log('[WASMService] Workspace analysis took:', elapsed, 'ms');
      
      if (!result) {
        return cachedResults;
      }
      
      const parsed = JSON.parse(result);
      if (!parsed || !parsed.results) {
        return cachedResults;
      }

      for (const [filename, analysisResult] of Object.entries(parsed.results)) {
        const res = analysisResult as AnalysisResult;
        this.cache.set(filename, { hash: this.computeHash(files.find(f => f.filename === filename)?.content || ''), result: res });
        cachedResults.set(filename, res);
      }
      
      return cachedResults;
    } catch (error) {
      console.error('[WASMService] Workspace analysis error:', error);
      return cachedResults;
    }
  }

  async analyze(request: AnalyzeRequest): Promise<AnalysisResult> {
    if (this.isJsTsFile(request.filename)) {
      return this.nativeJsAnalyzer.analyze(request.content, request.filename);
    }

    return this.analyzeWasm(request);
  }

  private async analyzeWasm(request: AnalyzeRequest): Promise<AnalysisResult> {
    await this.ensureInitialized();

    const contentHash = this.computeHash(request.content);
    
    const cached = this.cache.get(request.filename);
    if (cached && cached.hash === contentHash) {
      console.log('[WASMService] Using cache for:', request.filename);
      return cached.result;
    }

    console.log('[WASMService] Analyzing:', request.filename);

    const requestWithHash = {
      ...request,
      hash: contentHash
    };
    
    if (!analyzeCodeFn) {
      console.error('[WASMService] analyzeCode not initialized');
      return { imports: [], variables: [], parameters: [] };
    }
    
    try {
      const start = Date.now();
      const result = analyzeCodeFn(JSON.stringify(requestWithHash));
      const elapsed = Date.now() - start;
      
      if (!result) {
        return { imports: [], variables: [], parameters: [] };
      }
      
      const parsed = JSON.parse(result);
      if (!parsed) {
        return { imports: [], variables: [], parameters: [] };
      }

      const analysisResult: AnalysisResult = {
        imports: parsed.imports || [],
        variables: parsed.variables || [],
        parameters: parsed.parameters || []
      };

      this.cache.set(request.filename, { hash: contentHash, result: analysisResult });
      return analysisResult;
    } catch (error) {
      console.error('[WASMService] Analysis error:', error);
      return { imports: [], variables: [], parameters: [] };
    }
  }

  detectLanguage(filename: string): string {
    if (this.isJsTsFile(filename)) {
      return 'javascript/typescript';
    }
    
    if (!detectLanguageFn) {
      return 'unknown';
    }
    return detectLanguageFn(filename) || 'unknown';
  }

  terminate(): void {
    if (this.wasmModule) {
      this.wasmModule.exit();
    }
  }

  clearCache(): void {
    this.cache.clear();
  }
}
