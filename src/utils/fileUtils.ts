export function getFileExtension(filename: string): string {
    return filename.toLowerCase().split('.').pop() || '';
}

export function isJsTsFile(filename: string): boolean {
    const ext = getFileExtension(filename);
    return ['ts', 'tsx', 'js', 'jsx', 'mjs', 'cjs', 'vue', 'svelte'].includes(ext);
}

export function isGoFile(filename: string): boolean {
    return filename.toLowerCase().endsWith('.go');
}

export function isPythonFile(filename: string): boolean {
    return filename.toLowerCase().endsWith('.py');
}

export function isRubyFile(filename: string): boolean {
    return filename.toLowerCase().endsWith('.rb');
}

export function isPHPFile(filename: string): boolean {
    return filename.toLowerCase().endsWith('.php');
}

export function detectLanguage(filename: string): string {
    const ext = getFileExtension(filename);
    switch (ext) {
        case 'svelte':
            return 'svelte';
        case 'vue':
            return 'vue';
        case 'ts':
        case 'tsx':
            return 'typescript';
        case 'js':
        case 'jsx':
        case 'mjs':
        case 'cjs':
            return 'javascript';
        case 'py':
            return 'python';
        case 'go':
            return 'go';
        case 'rb':
            return 'ruby';
        case 'php':
            return 'php';
        default:
            return 'unknown';
    }
}

export function isRelevantFile(filePath: string, extensions: string[], excludeFolders: string[]): boolean {
    const ext = getFileExtension(filePath);
    if (!extensions.includes(ext)) {
        return false;
    }
    
    for (const folder of excludeFolders) {
        if (filePath.includes('/' + folder + '/') || filePath.endsWith('/' + folder)) {
            return false;
        }
    }
    
    return true;
}
