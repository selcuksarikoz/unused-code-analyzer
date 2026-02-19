import * as ts from 'typescript';
import type { AnalysisResult, CodeIssue, ImportInfo, VariableInfo, ParameterInfo } from '../types';
import { computeHash } from '../utils/hash';
import { generateUUID } from '../helpers/uuidHelper';
import { isGlobalIdentifier } from '../helpers/identifierHelpers';

interface NativeCacheEntry {
    hash: string;
    result: AnalysisResult;
}

export class NativeJsAnalyzer {
  private compilerOptions: ts.CompilerOptions;
  private cache: Map<string, NativeCacheEntry> = new Map();

  constructor() {
    this.compilerOptions = {
      target: ts.ScriptTarget.ES2020,
      module: ts.ModuleKind.ESNext,
      jsx: ts.JsxEmit.React,
      allowJs: true,
      checkJs: false,
      noEmit: true,
      skipLibCheck: true,
      esModuleInterop: true,
      allowSyntheticDefaultImports: true,
    };
  }

  analyze(content: string, filename: string): AnalysisResult {
    const hash = computeHash(content);
    const cached = this.cache.get(filename);
    
    if (cached && cached.hash === hash) {
      console.log(`[NativeJsAnalyzer] Cache hit for: ${filename}`);
      return cached.result;
    }

    console.log(`[NativeJsAnalyzer] Analyzing: ${filename}`);

    const language = this.detectLanguage(filename);
    
    let result: AnalysisResult;
    if (language === 'svelte' || language === 'vue') {
      result = this.analyzeVueSvelte(content, filename, language);
    } else {
      const sourceFile = ts.createSourceFile(
        filename,
        content,
        this.compilerOptions.target || ts.ScriptTarget.ES2020,
        true,
        this.getScriptKind(filename)
      );

      const imports = this.findImports(sourceFile);
      const variables = this.findVariables(sourceFile);
      const parameters = this.findParameters(sourceFile);

      const usedNames = this.findUsedNames(sourceFile, imports, variables, parameters);

      result = this.buildResult(imports, variables, parameters, usedNames, filename);
    }

    this.cache.set(filename, { hash, result });
    return result;
  }

  private detectLanguage(filename: string): string {
    const ext = filename.toLowerCase().split('.').pop();
    switch (ext) {
      case 'svelte':
        return 'svelte';
      case 'vue':
        return 'vue';
      case 'ts':
        return 'typescript';
      case 'tsx':
        return 'typescript';
      case 'js':
        return 'javascript';
      case 'jsx':
        return 'javascript';
      default:
        return 'javascript';
    }
  }

  private getScriptKind(filename: string): ts.ScriptKind {
    const ext = filename.toLowerCase().split('.').pop();
    switch (ext) {
      case 'ts':
        return ts.ScriptKind.TS;
      case 'tsx':
        return ts.ScriptKind.TSX;
      case 'jsx':
        return ts.ScriptKind.JSX;
      case 'js':
      default:
        return ts.ScriptKind.JS;
    }
  }

  private analyzeVueSvelte(content: string, filename: string, type: 'vue' | 'svelte'): AnalysisResult {
    let scriptContent = content;
    
    if (type === 'vue') {
      const scriptMatch = content.match(/<script[^>]*>([\s\S]*?)<\/script>/);
      if (!scriptMatch) {
        return { imports: [], variables: [], parameters: [] };
      }
      scriptContent = scriptMatch[1];
    } else if (type === 'svelte') {
      const scriptMatch = content.match(/<script[^>]*>([\s\S]*?)<\/script>/);
      if (scriptMatch) {
        scriptContent = scriptMatch[1];
      }
    }

    return this.analyze(scriptContent, filename + '.ts');
  }

  private findImports(sourceFile: ts.SourceFile): ImportInfo[] {
    const imports: ImportInfo[] = [];

    const visit = (node: ts.Node) => {
      if (ts.isImportDeclaration(node)) {
        const moduleSpecifier = node.moduleSpecifier;
        if (!ts.isStringLiteral(moduleSpecifier)) {
          ts.forEachChild(node, visit);
          return;
        }

        const source = moduleSpecifier.text;
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;

        if (node.importClause) {
          if (node.importClause.name) {
            imports.push({
              name: node.importClause.name.text,
              line,
              text: `import ${node.importClause.name.text}`,
              source,
            });
          }

          if (node.importClause.namedBindings) {
            if (ts.isNamedImports(node.importClause.namedBindings)) {
              node.importClause.namedBindings.elements.forEach((element) => {
                const name = element.name.text;
                const originalName = element.propertyName?.text || name;
                imports.push({
                  name,
                  line,
                  text: `import { ${originalName}${name !== originalName ? ` as ${name}` : ''} }`,
                  source,
                });
              });
            } else if (ts.isNamespaceImport(node.importClause.namedBindings)) {
              imports.push({
                name: node.importClause.namedBindings.name.text,
                line,
                text: `import * as ${node.importClause.namedBindings.name.text}`,
                source,
              });
            }
          }
        }
      } else if (ts.isCallExpression(node) &&
                 ts.isIdentifier(node.expression) &&
                 node.expression.text === 'require' &&
                 node.arguments.length > 0 &&
                 ts.isStringLiteral(node.arguments[0])) {
        const parent = node.parent;
        if (ts.isVariableDeclaration(parent)) {
          const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
          const source = node.arguments[0].text;
          
          if (ts.isIdentifier(parent.name)) {
            imports.push({
              name: parent.name.text,
              line,
              text: `const ${parent.name.text} = require('${source}')`,
              source,
            });
          } else if (ts.isObjectBindingPattern(parent.name)) {
            parent.name.elements.forEach((element) => {
              if (ts.isIdentifier(element.name)) {
                imports.push({
                  name: element.name.text,
                  line,
                  text: `const { ${element.name.text} } = require('${source}')`,
                  source,
                });
              }
            });
          }
        }
      }

      ts.forEachChild(node, visit);
    };

    visit(sourceFile);
    return imports;
  }

  private isInsideDeclareGlobal(node: ts.Node): boolean {
    let current: ts.Node | undefined = node.parent;
    while (current) {
      // Check if we're inside a ModuleDeclaration (namespace) with declare modifier
      if (ts.isModuleDeclaration(current)) {
        // Check if it's a global augmentation (declare global { ... })
        if (current.name && ts.isIdentifier(current.name) && current.name.text === 'global') {
          return true;
        }
      }
      current = current.parent;
    }
    return false;
  }

  private findVariables(sourceFile: ts.SourceFile): VariableInfo[] {
    const variables: VariableInfo[] = [];
    const seenNames = new Set<string>();

    const addVariable = (name: string, line: number, type: string, isExported: boolean) => {
      if (seenNames.has(name) || isGlobalIdentifier(name)) return;
      seenNames.add(name);
      variables.push({ name, line, type: type as VariableInfo['type'], exported: isExported });
    };

    const checkExported = (node: ts.Node): boolean => {
      let current: ts.Node | undefined = node.parent;
      while (current) {
        if (ts.isVariableStatement(current)) {
          return current.modifiers?.some((m: ts.ModifierLike) => m.kind === ts.SyntaxKind.ExportKeyword) ?? false;
        }
        if (ts.isFunctionDeclaration(current) || ts.isClassDeclaration(current) || 
            ts.isInterfaceDeclaration(current) || ts.isTypeAliasDeclaration(current) ||
            ts.isEnumDeclaration(current)) {
          return current.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword) ?? false;
        }
        current = current.parent;
      }
      return false;
    };

    const visit = (node: ts.Node) => {
      // Variable declarations - simple: const foo = ...
      if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name)) {
        const name = node.name.text;
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        addVariable(name, line, 'var', checkExported(node));
      } 
      // Variable declarations - destructuring: const { a, b } = ...
      else if (ts.isVariableDeclaration(node) && ts.isObjectBindingPattern(node.name)) {
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        this.extractBindingNames(node.name).forEach(name => {
          addVariable(name, line, 'var', checkExported(node));
        });
      }
      // Variable declarations - array destructuring: const [a, b] = ...
      else if (ts.isVariableDeclaration(node) && ts.isArrayBindingPattern(node.name)) {
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        this.extractBindingNames(node.name).forEach(name => {
          addVariable(name, line, 'var', checkExported(node));
        });
      }
      // Function declarations
      else if (ts.isFunctionDeclaration(node) && node.name) {
        const name = node.name.text;
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        addVariable(name, line, 'function', checkExported(node));
      } 
      // Class declarations
      else if (ts.isClassDeclaration(node) && node.name) {
        const name = node.name.text;
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        addVariable(name, line, 'class', checkExported(node));
      }
      // Interface declarations
      else if (ts.isInterfaceDeclaration(node) && node.name) {
        const name = node.name.text;
        if (seenNames.has(name)) return;
        
        if (this.isInsideDeclareGlobal(node)) {
          return;
        }
        
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        const isExported = node.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword);
        
        seenNames.add(name);
        variables.push({
          name,
          line,
          type: 'interface',
          exported: !!isExported,
        });
      } 
      // Type alias declarations
      else if (ts.isTypeAliasDeclaration(node)) {
        const name = node.name.text;
        if (seenNames.has(name)) return;
        
        if (this.isInsideDeclareGlobal(node)) {
          return;
        }
        
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        const isExported = node.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword);
        
        seenNames.add(name);
        variables.push({
          name,
          line,
          type: 'type',
          exported: !!isExported,
        });
      } 
      // Enum declarations
      else if (ts.isEnumDeclaration(node)) {
        const name = node.name.text;
        if (seenNames.has(name)) return;
        
        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        const isExported = node.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword);
        
        seenNames.add(name);
        variables.push({
          name,
          line,
          type: 'enum',
          exported: !!isExported,
        });
      }

      ts.forEachChild(node, visit);
    };

    visit(sourceFile);
    return variables;
  }

  private extractBindingNames(pattern: ts.BindingPattern): string[] {
    const names: string[] = [];

    const visit = (element: ts.BindingElement | ts.ArrayBindingElement) => {
      if (ts.isBindingElement(element)) {
        if (ts.isIdentifier(element.name)) {
          names.push(element.name.text);
        } else if (ts.isObjectBindingPattern(element.name) || ts.isArrayBindingPattern(element.name)) {
          this.extractBindingNames(element.name).forEach((n) => names.push(n));
        }
      }
    };

    pattern.elements.forEach(visit);
    return names;
  }

  private findParameters(sourceFile: ts.SourceFile): ParameterInfo[] {
    const parameters: ParameterInfo[] = [];
    const seenParams = new Set<string>();

    const visit = (node: ts.Node) => {
      if ((ts.isFunctionDeclaration(node) || ts.isMethodDeclaration(node) || ts.isArrowFunction(node)) && node.parameters) {
        const funcName = this.getFunctionName(node);
        const funcLine = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;

        node.parameters.forEach((param) => {
          if (ts.isIdentifier(param.name)) {
            const name = param.name.text;
            const key = `${funcName}:${name}`;
            if (name !== 'this' && name !== '_' && !seenParams.has(key)) {
              seenParams.add(key);
              parameters.push({
                name,
                line: funcLine,
                funcName: funcName || '(anonymous)',
              });
            }
          } else if (ts.isObjectBindingPattern(param.name) || ts.isArrayBindingPattern(param.name)) {
            this.extractBindingElements(param.name).forEach((name) => {
              const key = `${funcName}:${name}`;
              if (name !== '_' && !seenParams.has(key)) {
                seenParams.add(key);
                parameters.push({
                  name,
                  line: funcLine,
                  funcName: funcName || '(anonymous)',
                });
              }
            });
          }
        });
      }

      ts.forEachChild(node, visit);
    };

    visit(sourceFile);
    return parameters;
  }

  private getFunctionName(node: ts.FunctionDeclaration | ts.MethodDeclaration | ts.ArrowFunction): string | null {
    if (ts.isFunctionDeclaration(node) && node.name) {
      return node.name.text;
    } else if (ts.isMethodDeclaration(node) && ts.isIdentifier(node.name)) {
      return node.name.text;
    } else if (ts.isVariableDeclaration(node.parent) && ts.isIdentifier(node.parent.name)) {
      return node.parent.name.text;
    }
    return null;
  }

  private extractBindingElements(pattern: ts.BindingPattern): string[] {
    const names: string[] = [];

    const visit = (element: ts.BindingElement | ts.ArrayBindingElement) => {
      if (ts.isBindingElement(element)) {
        if (ts.isIdentifier(element.name)) {
          names.push(element.name.text);
        } else if (ts.isBindingElement(element) && (ts.isObjectBindingPattern(element.name) || ts.isArrayBindingPattern(element.name))) {
          this.extractBindingElements(element.name).forEach((n) => names.push(n));
        }
      }
    };

    pattern.elements.forEach(visit);
    return names;
  }

  private findUsedNames(
    sourceFile: ts.SourceFile,
    imports: ImportInfo[],
    variables: VariableInfo[],
    parameters: ParameterInfo[]
  ): Set<string> {
    const usedNames = new Set<string>();

    const allItems = new Map<string, number>();
    imports.forEach((imp) => allItems.set(imp.name, imp.line));
    variables.forEach((v) => allItems.set(v.name, v.line));
    parameters.forEach((p) => allItems.set(p.name, p.line));

    const isDefinitionSite = (node: ts.Identifier): boolean => {
      const parent = node.parent;
      // These are places where an identifier is being defined/declared, not used
      if (ts.isImportClause(parent) && parent.name === node) return true;
      if (ts.isImportSpecifier(parent) && parent.name === node) return true;
      if (ts.isNamespaceImport(parent) && parent.name === node) return true;
      if (ts.isVariableDeclaration(parent) && parent.name === node) return true;
      if (ts.isFunctionDeclaration(parent) && parent.name === node) return true;
      if (ts.isClassDeclaration(parent) && parent.name === node) return true;
      if (ts.isInterfaceDeclaration(parent) && parent.name === node) return true;
      if (ts.isTypeAliasDeclaration(parent) && parent.name === node) return true;
      if (ts.isEnumDeclaration(parent) && parent.name === node) return true;
      if (ts.isParameter(parent) && parent.name === node) return true;
      if (ts.isPropertySignature(parent) && parent.name === node) return true;
      if (ts.isPropertyDeclaration(parent) && parent.name === node) return true;
      if (ts.isPropertyAssignment(parent) && parent.name === node) return true;
      if (ts.isShorthandPropertyAssignment(parent) && parent.name === node) return true;
      if (ts.isBindingElement(parent) && parent.name === node) return true;
      
      // Check if this identifier is the property name in a property access (obj.prop)
      // In this case, 'prop' is not being used as a variable, it's a property name
      if (ts.isPropertyAccessExpression(parent) && parent.name === node) return true;
      
      return false;
    };

    const visit = (node: ts.Node) => {
      if (ts.isIdentifier(node)) {
        const name = node.text;
        
        // Check if this is a name we care about
        if (!allItems.has(name)) {
          ts.forEachChild(node, visit);
          return;
        }

        const line = ts.getLineAndCharacterOfPosition(sourceFile, node.getStart()).line + 1;
        const definitionLine = allItems.get(name);

        // Skip if on definition line
        if (definitionLine === line) {
          ts.forEachChild(node, visit);
          return;
        }

        // Skip if this is a definition site
        if (isDefinitionSite(node)) {
          ts.forEachChild(node, visit);
          return;
        }

        // This is a usage!
        usedNames.add(name);
      }

      ts.forEachChild(node, visit);
    };

    visit(sourceFile);
    return usedNames;
  }

  private buildResult(
    imports: ImportInfo[],
    variables: VariableInfo[],
    parameters: ParameterInfo[],
    usedNames: Set<string>,
    filename: string
  ): AnalysisResult {
    const unusedImports: CodeIssue[] = [];
    const unusedVars: CodeIssue[] = [];
    const unusedParams: CodeIssue[] = [];

    imports.forEach((imp) => {
      if (!usedNames.has(imp.name)) {
        unusedImports.push({
          id: generateUUID(),
          line: imp.line,
          text: imp.text,
          file: filename,
        });
      }
    });

    variables.forEach((v) => {
      if (!v.exported && !usedNames.has(v.name)) {
        unusedVars.push({
          id: generateUUID(),
          line: v.line,
          text: `${v.type} ${v.name}`,
          file: filename,
        });
      }
    });

    parameters.forEach((p) => {
      if (!usedNames.has(p.name)) {
        unusedParams.push({
          id: generateUUID(),
          line: p.line,
          text: `parameter ${p.name}`,
          file: filename,
        });
      }
    });

    return {
      imports: unusedImports,
      variables: unusedVars,
      parameters: unusedParams,
    };
  }

  analyzeWorkspace(
    files: Array<{ content: string; filename: string }>
  ): Map<string, AnalysisResult> {
    const results = new Map<string, AnalysisResult>();
    const fileData = new Map<string, {
      sourceFile: ts.SourceFile;
      imports: ImportInfo[];
      variables: VariableInfo[];
      parameters: ParameterInfo[];
    }>();

    // Phase 1: Parse all files and collect imports/definitions
    files.forEach(({ content, filename }) => {
      const sourceFile = ts.createSourceFile(
        filename,
        content,
        this.compilerOptions.target || ts.ScriptTarget.ES2020,
        true,
        this.getScriptKind(filename)
      );

      const imports = this.findImports(sourceFile);
      const variables = this.findVariables(sourceFile);
      const parameters = this.findParameters(sourceFile);

      fileData.set(filename, { sourceFile, imports, variables, parameters });
    });

    // Phase 2: Collect all usages across ALL files
    const globalUsedNames = new Set<string>();
    fileData.forEach(({ sourceFile }) => {
      // Collect all names defined in all files
      const allNames = new Set<string>();
      fileData.forEach(({ imports, variables, parameters }) => {
        imports.forEach(imp => allNames.add(imp.name));
        variables.forEach(v => allNames.add(v.name));
        parameters.forEach(p => allNames.add(p.name));
      });

      // Find usages in this file
      this.findUsedNamesInFile(sourceFile, allNames, globalUsedNames);
    });

    // Phase 3: Build results for each file
    fileData.forEach(({ imports, variables, parameters }, filename) => {
      const result = this.buildResult(imports, variables, parameters, globalUsedNames, filename);
      results.set(filename, result);
    });

    return results;
  }

  private findUsedNamesInFile(
    sourceFile: ts.SourceFile,
    allNames: Set<string>,
    globalUsedNames: Set<string>
  ): void {
    const isDefinitionSite = (node: ts.Identifier): boolean => {
      const parent = node.parent;
      if (ts.isImportClause(parent) && parent.name === node) return true;
      if (ts.isImportSpecifier(parent) && parent.name === node) return true;
      if (ts.isNamespaceImport(parent) && parent.name === node) return true;
      if (ts.isVariableDeclaration(parent) && parent.name === node) return true;
      if (ts.isFunctionDeclaration(parent) && parent.name === node) return true;
      if (ts.isClassDeclaration(parent) && parent.name === node) return true;
      if (ts.isInterfaceDeclaration(parent) && parent.name === node) return true;
      if (ts.isTypeAliasDeclaration(parent) && parent.name === node) return true;
      if (ts.isEnumDeclaration(parent) && parent.name === node) return true;
      if (ts.isParameter(parent) && parent.name === node) return true;
      if (ts.isPropertySignature(parent) && parent.name === node) return true;
      if (ts.isPropertyDeclaration(parent) && parent.name === node) return true;
      if (ts.isPropertyAssignment(parent) && parent.name === node) return true;
      if (ts.isShorthandPropertyAssignment(parent) && parent.name === node) return true;
      if (ts.isBindingElement(parent) && parent.name === node) return true;
      if (ts.isPropertyAccessExpression(parent) && parent.name === node) return true;
      return false;
    };

    const visit = (node: ts.Node) => {
      if (ts.isIdentifier(node)) {
        const name = node.text;
        
        // Only check names we're interested in
        if (!allNames.has(name)) {
          ts.forEachChild(node, visit);
          return;
        }

        // Skip if this is a definition site
        if (isDefinitionSite(node)) {
          ts.forEachChild(node, visit);
          return;
        }

        // This is a usage - mark as used globally
        globalUsedNames.add(name);
      }

      ts.forEachChild(node, visit);
    };

    visit(sourceFile);
  }

  private findExports(sourceFile: ts.SourceFile): Set<string> {
    const exports = new Set<string>();

    const visit = (node: ts.Node) => {
      if (ts.isExportDeclaration(node)) {
        if (node.exportClause && ts.isNamedExports(node.exportClause)) {
          node.exportClause.elements.forEach((element) => {
            exports.add(element.name.text);
          });
        }
      } else if ((ts.isVariableStatement(node) || ts.isFunctionDeclaration(node) || 
                  ts.isClassDeclaration(node) || ts.isInterfaceDeclaration(node) ||
                  ts.isTypeAliasDeclaration(node) || ts.isEnumDeclaration(node)) &&
                 node.modifiers?.some(m => m.kind === ts.SyntaxKind.ExportKeyword)) {
        if (ts.isVariableStatement(node)) {
          node.declarationList.declarations.forEach((decl) => {
            if (ts.isIdentifier(decl.name)) {
              exports.add(decl.name.text);
            }
          });
        } else if ((ts.isFunctionDeclaration(node) || ts.isClassDeclaration(node) || 
                    ts.isInterfaceDeclaration(node)) && node.name) {
          exports.add(node.name.text);
        } else if (ts.isTypeAliasDeclaration(node)) {
          exports.add(node.name.text);
        } else if (ts.isEnumDeclaration(node)) {
          exports.add(node.name.text);
        }
      }

      ts.forEachChild(node, visit);
    };

    visit(sourceFile);
    return exports;
  }
}
