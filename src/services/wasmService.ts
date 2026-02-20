import * as fs from "fs";
import * as path from "path";
import * as vscode from "vscode";
import { EXTENSION_ID } from "../constants";
import type { AnalysisResult, AnalyzeRequest, WorkspaceFile } from "../types";
import { isJsTsFile } from "../utils/fileUtils";
import { computeHash } from "../utils/hash";

let wasmInitialized = false;
let analyzeCodeFn: any = null;
let analyzeWorkspaceFn: any = null;
let detectLanguageFn: any = null;

export class WasmService {
  private wasmModule: any = null;
  private extensionId = EXTENSION_ID;
  private cache: Map<string, { hash: string; result: AnalysisResult }> =
    new Map();
  private workspaceHashes: Map<string, string> = new Map();
  private analyzerChecked = false;

  constructor() {}

  private async detectBestAnalyzer(): Promise<"tsserver"> {
    if (this.analyzerChecked) {
      return "tsserver";
    }

    console.log("[WASMService] Using WASM analyzer");
    this.analyzerChecked = true;
    return "tsserver";
  }

  async initialize(forceWasmLoad = false): Promise<void> {
    await this.detectBestAnalyzer();

    if (!forceWasmLoad) return;

    if (wasmInitialized) {
      console.log("[WASMService] WASM already initialized");
      return;
    }

    const timeoutMs = 30000;
    const timeoutPromise = new Promise((_, reject) =>
      setTimeout(
        () =>
          reject(new Error(`Initialization timed out after ${timeoutMs}ms`)),
        timeoutMs,
      ),
    );

    const initPromise = this.doInitialize();

    try {
      await Promise.race([initPromise, timeoutPromise]);
    } catch (error) {
      console.error("[WASMService] Initialization failed:", error);
      throw error;
    }
  }

  private async doInitialize(): Promise<void> {
    try {
      console.log("[WASMService] Getting extension path...");
      const extensionPath = vscode.extensions.getExtension(
        this.extensionId,
      )?.extensionPath;
      console.log("[WASMService] Extension path:", extensionPath);
      if (!extensionPath) {
        throw new Error("Extension path not found");
      }

      const wasmPath = path.join(extensionPath, "out", "main.wasm");
      const wasmExecPath = path.join(extensionPath, "out", "wasm_exec.js");

      console.log("[WASMService] Checking files...");
      console.log("[WASMService] WASM path:", wasmPath);
      console.log("[WASMService] WASM exists:", fs.existsSync(wasmPath));
      console.log(
        "[WASMService] wasm_exec exists:",
        fs.existsSync(wasmExecPath),
      );

      if (!fs.existsSync(wasmPath)) {
        throw new Error(`WASM file not found at ${wasmPath}`);
      }

      console.log("[WASMService] Reading WASM binary...");
      const wasmBinary = fs.readFileSync(wasmPath);
      console.log("[WASMService] WASM binary size:", wasmBinary.length);

      console.log("[WASMService] Loading wasm_exec.js...");
      require(wasmExecPath);
      console.log("[WASMService] wasm_exec.js loaded");

      console.log("[WASMService] Creating Go instance...");
      const go = new (globalThis as any).Go();
      console.log("[WASMService] Go instance created");

      console.log("[WASMService] Instantiating WASM...");
      const WebAssembly = globalThis.WebAssembly;
      const result = await WebAssembly.instantiate(wasmBinary, go.importObject);
      console.log("[WASMService] WASM instantiated");

      console.log("[WASMService] Running Go...");
      go.run(result.instance);
      console.log("[WASMService] Go started (non-blocking)");

      analyzeCodeFn = (globalThis as any).analyzeCode;
      analyzeWorkspaceFn = (globalThis as any).analyzeWorkspace;
      detectLanguageFn = (globalThis as any).detectLanguage;

      console.log("[WASMService] analyzeCode function:", typeof analyzeCodeFn);
      console.log(
        "[WASMService] analyzeWorkspace function:",
        typeof analyzeWorkspaceFn,
      );
      console.log(
        "[WASMService] detectLanguage function:",
        typeof detectLanguageFn,
      );

      wasmInitialized = true;
      console.log("[WASMService] Initialization complete!");
    } catch (error) {
      console.error("[WASMService] Initialization error:", error);
      throw error;
    }
  }

  async ensureInitialized(): Promise<void> {
    if (!wasmInitialized) {
      await this.initialize(true);
    }
  }

  async analyzeWorkspace(
    files: WorkspaceFile[],
  ): Promise<Map<string, AnalysisResult>> {
    if (!this.analyzerChecked) {
      await this.detectBestAnalyzer();
    }
    return this.analyzeWorkspaceWasm(files);
  }

  private async analyzeWorkspaceWasm(
    files: WorkspaceFile[],
  ): Promise<Map<string, AnalysisResult>> {
    await this.ensureInitialized();

    if (!analyzeWorkspaceFn) {
      console.error("[WASMService] analyzeWorkspace not initialized");
      return new Map();
    }

    try {
      const start = Date.now();
      const result = analyzeWorkspaceFn(
        JSON.stringify({ files }),
      );
      const elapsed = Date.now() - start;
      console.log("[WASMService] Workspace analysis took:", elapsed, "ms");

      if (!result) {
        console.log("[WASMService] No result from WASM");
        return new Map();
      }

      const parsed = JSON.parse(result);
      const resultsKey = parsed.results ? "results" : "Results";
      const resultsMap = new Map<string, AnalysisResult>();

      for (const [filename, analysisResult] of Object.entries(
        parsed[resultsKey] || {},
      )) {
        const res = analysisResult as any;
        const analysis: AnalysisResult = {
          imports: res.imports || res.Imports || [],
          variables: res.variables || res.Variables || [],
          parameters: res.parameters || res.Parameters || [],
        };
        resultsMap.set(filename, analysis);
      }

      return resultsMap;
    } catch (error) {
      console.error("[WASMService] Workspace analysis error:", error);
      return new Map();
    }
  }

  async analyze(request: AnalyzeRequest): Promise<AnalysisResult> {
    if (!this.analyzerChecked) {
      await this.detectBestAnalyzer();
    }
    return this.analyzeWasm(request);
  }

  private async analyzeWasm(request: AnalyzeRequest): Promise<AnalysisResult> {
    await this.ensureInitialized();

    if (!analyzeCodeFn) {
      console.error("[WASMService] analyzeCode not initialized");
      return { imports: [], variables: [], parameters: [] };
    }

    try {
      const start = Date.now();
      const result = analyzeCodeFn(JSON.stringify(request));
      const elapsed = Date.now() - start;
      console.log(
        "[WASMService] Single file analysis took:",
        elapsed,
        "ms",
      );

      if (!result) {
        return { imports: [], variables: [], parameters: [] };
      }

      const parsed = JSON.parse(result);
      if (!parsed) {
        return { imports: [], variables: [], parameters: [] };
      }

      return {
        imports: parsed.imports || parsed.Imports || [],
        variables: parsed.variables || parsed.Variables || [],
        parameters: parsed.parameters || parsed.Parameters || [],
      };
    } catch (error) {
      console.error("[WASMService] Analysis error:", error);
      return { imports: [], variables: [], parameters: [] };
    }
  }

  detectLanguage(filename: string): string {
    if (isJsTsFile(filename)) {
      return "javascript/typescript";
    }

    if (!detectLanguageFn) {
      return "unknown";
    }
    return detectLanguageFn(filename) || "unknown";
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
