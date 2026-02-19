export interface ImportInfo {
  name: string;
  line: number;
  text: string;
  source: string;
}

export interface VariableInfo {
  name: string;
  line: number;
  type: string;
  exported: boolean;
}

export interface ParameterInfo {
  name: string;
  line: number;
  funcName: string;
}

export interface CacheEntry {
  hash: string;
  result: import('../types').AnalysisResult;
}
