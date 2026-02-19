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
