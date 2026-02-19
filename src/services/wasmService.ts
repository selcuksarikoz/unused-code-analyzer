import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';
import type { AnalysisResult, AnalyzeRequest, WorkspaceFile } from '../types';
import { EXTENSION_ID } from '../constants';
import { computeHash } from '../utils/hash';
import { isJsTsFile, isGoFile, isPythonFile } from '../utils/fileUtils';
import { NativeJsAnalyzer } from './nativeJsAnalyzer';

let wasmInitialized = false;
let analyzeCodeFn: any = null;
let analyzeWorkspaceFn: any = null;
let detectLanguageFn: any = null;

export class WasmService {
  private wasmModule: any = null;
  private extensionId = EXTENSION_ID;
  private cache: Map<string, { hash: string; result: AnalysisResult }> = new Map();
  private workspaceHashes: Map<string, string> = new Map();
  private nativeJsAnalyzer: NativeJsAnalyzer;

  constructor() {
    this.nativeJsAnalyzer = new NativeJsAnalyzer();
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
      if (isJsTsFile(file.filename)) {
        jsTsFiles.push(file);
      } else if (isGoFile(file.filename) || isPythonFile(file.filename)) {
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
      hash: computeHash(f.content)
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
      console.log('[WASMService] Raw result:', result?.substring(0, 500));
      
      if (!result) {
        console.log('[WASMService] No result from WASM');
        return cachedResults;
      }
      
      const parsed = JSON.parse(result);
      const resultsKey = parsed.results ? 'results' : 'Results';
      console.log('[WASMService] Parsed results count:', Object.keys(parsed[resultsKey] || {}).length);
      
      for (const [filename, analysisResult] of Object.entries(parsed[resultsKey])) {
        const res = analysisResult as any;
        console.log('[WASMService] File:', filename, '-> imports:', res.imports?.length || res.Imports?.length, 'variables:', res.variables?.length || res.Variables?.length, 'parameters:', res.parameters?.length || res.Parameters?.length);
        const result: AnalysisResult = {
          imports: res.imports || res.Imports || [],
          variables: res.variables || res.Variables || [],
          parameters: res.parameters || res.Parameters || []
        };
        this.cache.set(filename, { hash: computeHash(files.find(f => f.filename === filename)?.content || ''), result });
        cachedResults.set(filename, result);
      }
      
      return cachedResults;
    } catch (error) {
      console.error('[WASMService] Workspace analysis error:', error);
      return cachedResults;
    }
  }

  async analyze(request: AnalyzeRequest): Promise<AnalysisResult> {
    if (isJsTsFile(request.filename)) {
      return this.nativeJsAnalyzer.analyze(request.content, request.filename);
    }

    return this.analyzeWasm(request);
  }

  private async analyzeWasm(request: AnalyzeRequest): Promise<AnalysisResult> {
    await this.ensureInitialized();

    const contentHash = computeHash(request.content);
    
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
      console.log('[WASMService] Single file analysis took:', elapsed, 'ms, result:', result?.substring(0, 200));
      
      if (!result) {
        return { imports: [], variables: [], parameters: [] };
      }
      
      const parsed = JSON.parse(result);
      console.log('[WASMService] Parsed single result:', JSON.stringify(parsed).substring(0, 200));
      if (!parsed) {
        return { imports: [], variables: [], parameters: [] };
      }

      const analysisResult: AnalysisResult = {
        imports: (parsed.imports || parsed.Imports || []),
        variables: (parsed.variables || parsed.Variables || []),
        parameters: (parsed.parameters || parsed.Parameters || [])
      };

      this.cache.set(request.filename, { hash: contentHash, result: analysisResult });
      return analysisResult;
    } catch (error) {
      console.error('[WASMService] Analysis error:', error);
      return { imports: [], variables: [], parameters: [] };
    }
  }

  detectLanguage(filename: string): string {
    if (isJsTsFile(filename)) {
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
