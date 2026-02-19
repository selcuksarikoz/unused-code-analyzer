const GLOBAL_IDENTIFIERS = new Set([
    'window', 'document', 'console', 'process', 'global', 'require',
    'module', 'exports', 'Buffer', 'setTimeout', 'setInterval',
    'clearTimeout', 'clearInterval', 'setImmediate', 'clearImmediate',
    '__dirname', '__filename', 'undefined', 'null', 'true', 'false',
    'NaN', 'Infinity', 'eval', 'parseInt', 'parseFloat', 'isNaN', 'isFinite',
    'decodeURI', 'decodeURIComponent', 'encodeURI', 'encodeURIComponent',
    'escape', 'unescape',
]);

export function isGlobalIdentifier(name: string): boolean {
    return GLOBAL_IDENTIFIERS.has(name);
}

export function isPrivateIdentifier(name: string): boolean {
    return name.startsWith('_');
}

export function shouldSkipIdentifier(name: string): boolean {
    return isGlobalIdentifier(name) || isPrivateIdentifier(name);
}
