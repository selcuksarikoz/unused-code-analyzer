export interface CodeIssue {
    id: string;
    line: number;
    text: string;
    file: string;
}

export interface AnalysisResult {
    imports: CodeIssue[];
    variables: CodeIssue[];
    parameters: CodeIssue[];
}

export interface AnalyzeRequest {
    content: string;
    filename: string;
    language?: string;
    hash?: string;
}

export interface WorkspaceAnalyzeRequest {
    files: AnalyzeRequest[];
}

export interface ImportInfo {
    name: string;
    line: number;
    text: string;
    source: string;
}

export interface VariableInfo {
    name: string;
    line: number;
    type: 'var' | 'function' | 'class' | 'interface' | 'type' | 'enum';
    exported: boolean;
}

export interface ParameterInfo {
    name: string;
    line: number;
    funcName: string;
}

export interface CacheEntry<T> {
    hash: string;
    result: T;
}

export interface WorkspaceFile {
    content: string;
    filename: string;
    hash?: string;
}
