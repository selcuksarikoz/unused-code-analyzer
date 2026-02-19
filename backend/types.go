package main

type Language string

type AnalysisResult struct {
	Imports    []CodeIssue `json:"imports"`
	Variables  []CodeIssue `json:"variables"`
	Parameters []CodeIssue `json:"parameters"`
}

type CodeIssue struct {
	ID   string `json:"id"`
	Line int    `json:"line"`
	Text string `json:"text"`
	File string `json:"file"`
}

type AnalyzeRequest struct {
	Content  string `json:"content"`
	Filename string `json:"filename"`
	Language string `json:"language"`
	Hash     string `json:"hash"`
}

type WorkspaceAnalyzeRequest struct {
	Files []AnalyzeFile `json:"files"`
}

type AnalyzeFile struct {
	Content  string `json:"content"`
	Filename string `json:"filename"`
	Hash     string `json:"hash"`
}

type Definition struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Type     string `json:"type"`
	Exported bool   `json:"exported"`
}

type Import struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Source string `json:"source"`
}
